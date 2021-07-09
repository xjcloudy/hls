package parser

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// PlayList
type PlayList struct {
	// version of m3u8
	Version             int
	PlayType            PlayListTypeEnum
	IndependentSegments bool
	TargetDuration      int64
	MediaSequence       int64
	// m3u8 address
	address *url.URL
	// file name after merge
	outputFile string
	// target dir which ts file download
	tempDir string
	// origin data of m3u8 file
	file io.Reader

	// ts file list
	mediaSegments []*MediaSegment

	// latest key found in m3u8 file
	currentKey *Key
}

// NewPlayList create new MediaPlay instace
func NewPlayList(targetURL string, filename string) (*PlayList, error) {
	// verfiy url
	address, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	ins := &PlayList{}
	ins.address = address
	ins.outputFile = filename
	return ins, nil
}

// MediaSegment media segment
// https://tools.ietf.org/html/draft-pantos-hls-rfc8216bis-08#page-22
type MediaSegment struct {
	// duration and title are came from EXTINF tag
	// https://tools.ietf.org/html/draft-pantos-hls-rfc8216bis-08#page-17
	Duration int64
	Title    string

	Discontinuity bool
	Seq           int64
	URL           *url.URL
	localPath     string

	Key *Key
}

func (pl *PlayList) prepare() error {
	// download m3u8 file
	resp, err := http.Get(pl.address.String())
	if err != nil {
		return err
	}
	pl.file = resp.Body

	// create output dir
	if pl.outputFile == "" {
		pl.outputFile = DEFAULT_OUTPUT_FILENAME
	}
	dir, err := os.MkdirTemp("./", pl.outputFile+"_*")
	if err != nil {
		return err
	}
	pl.tempDir = dir

	return nil
}
func (pl *PlayList) parse() error {
	scanner := bufio.NewScanner(pl.file)
	// read line by line
	for scanner.Scan() {
		token := scanner.Text()
		switch {
		case strings.HasPrefix(token, VERSION):
			pl.parseVersion(token)
		case strings.HasPrefix(token, PLAY_LIST_TYPE):
			pl.parsePlayListType(token)
		case strings.HasPrefix(token, TARGETDURATION):
			pl.parseDuration(token)
		case strings.HasPrefix(token, MEDIA_SEQUENCE):
			pl.parseMediaSequence(token)
		case strings.HasPrefix(token, KEY):
			pl.parseKey(token)
		case strings.HasPrefix(token, EXTINF):
			// read next line
			scanner.Scan()
			pl.parseMediaSegment(token, scanner.Text())
		}

	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
func (pl *PlayList) parseVersion(token string) {
	rg := regexp.MustCompile(`^#EXT-X-VERSION:(\d+)$`)
	if rg.MatchString(token) {
		subs := rg.FindStringSubmatch(token)
		pl.Version, _ = strconv.Atoi(subs[1])
	}

}
func (pl *PlayList) parseKey(token string) {
	rg := regexp.MustCompile(`(METHOD|URI|IV|KEYFORMAT|KEYFORMATVERSIONS)="?([^,"]+)"?`)
	if rg.MatchString(token) {
		subs := rg.FindAllStringSubmatch(token, -1)
		key := Key{}
		for _, matchGroup := range subs {
			switch matchGroup[1] {
			case "METHOD":
				key.Method = KeyMethodEnums(matchGroup[2])
			case "URI":
				key.URI = matchGroup[2]
			case "IV":
				if iv, err := strconv.ParseUint(matchGroup[2], 16, 128); err == nil {
					key.IV = iv
				}
			case "KEYFORMAT":
				key.Format = matchGroup[2]
			case "KEYFORMATVERSIONS":
				key.FormatVersion = matchGroup[2]
			}
		}
		pl.currentKey = &key

	}

}
func (pl *PlayList) parseDuration(token string) {
	rg := regexp.MustCompile(`^#EXT-X-TARGETDURATION:(\d+)$`)
	if rg.MatchString(token) {
		subs := rg.FindStringSubmatch(token)
		pl.TargetDuration, _ = strconv.ParseInt(subs[1], 10, 64)
	}
}
func (pl *PlayList) parsePlayListType(token string) {
	rg := regexp.MustCompile(`^#EXT-X-PLAYLIST-TYPE:(EVENT|VOD)$`)
	if rg.MatchString(token) {
		subs := rg.FindStringSubmatch(token)
		pl.PlayType = PlayListTypeEnum(subs[1])
	}
}
func (pl *PlayList) parseMediaSequence(token string) {
	rg := regexp.MustCompile(`^#EXT-X-MEDIA-SEQUENCE:(\d+)$`)
	if rg.MatchString(token) {
		subs := rg.FindStringSubmatch(token)
		pl.MediaSequence, _ = strconv.ParseInt(subs[1], 10, 64)
	}
}
func (pl *PlayList) parseMediaSegment(token string, url string) {
	rg := regexp.MustCompile(`^#EXTINF:(\d+),?(.+)?$`)

	if rg.MatchString(token) {
		subs := rg.FindStringSubmatch(token)
		duration, _ := strconv.ParseInt(subs[1], 10, 64)
		tsurl, _ := pl.getTSFileURL(url)
		ms := MediaSegment{
			Title:    subs[2],
			Duration: duration,
			URL:      tsurl,
			Seq:      pl.MediaSequence,
			Key:      pl.currentKey,
		}
		pl.mediaSegments = append(pl.mediaSegments, &ms)
		pl.MediaSequence++
	}

}
func (pl *PlayList) getTSFileURL(tsurl string) (*url.URL, error) {
	// is absolut path
	ok, _ := regexp.MatchString(`(^http:|https:)`, tsurl)
	if !ok {
		u, err := url.Parse(tsurl)
		return pl.address.ResolveReference(u), err
	}
	return url.Parse(tsurl)
}
func (pl *PlayList) download() {
	joblist := make(chan *MediaSegment)
	var wg sync.WaitGroup
	// 3 goroutine
	for i := 0; i <= 3; i++ {
		wg.Add(1)
		go func(job <-chan *MediaSegment) {
			for ms := range job {
				//TODO retry
				downloadTSFile(ms, pl.tempDir)
			}
			wg.Done()
		}(joblist)
	}
	// distribute
	go func() {
		wg.Add(1)
		for _, ms := range pl.mediaSegments {
			joblist <- ms
		}
		// close job channel
		close(joblist)
		wg.Done()
	}()
	wg.Wait()

}

func downloadTSFile(ms *MediaSegment, dir string) error {
	resp, err := http.Get(ms.URL.String())
	if err != nil {
		return err
	}
	localpath := fmt.Sprintf("%s/%s.ts", dir, strconv.FormatInt(ms.Seq, 10))
	outputFile, err := os.Create(localpath)

	if _, copyerr := io.Copy(outputFile, resp.Body); copyerr != nil {
		return copyerr
	}
	defer resp.Body.Close()
	defer outputFile.Close()
	ms.localPath = localpath
	return nil

}
func (pl *PlayList) merge() error {
	var outputFile *os.File
	var err error
	if _, exists := os.Stat(pl.outputFile + ".mp4"); os.IsNotExist(exists) {
		// create outputfile
		outputFile, err = os.Create(pl.outputFile + ".mp4")
	} else {
		outputFile, err = os.CreateTemp("./", pl.outputFile+"_*"+".mp4")
	}

	if err != nil {
		return err
	}

	defer outputFile.Close()

	// use copy to merge
	for _, ms := range pl.mediaSegments {
		msFile, readerr := os.Open(ms.localPath)
		if readerr != nil {
			continue
		}
		_, copyerr := io.Copy(outputFile, msFile)
		if copyerr != nil {
			continue
		}
		defer msFile.Close()
	}
	return nil
}
