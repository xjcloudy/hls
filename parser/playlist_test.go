package parser

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestNewPlayList(t *testing.T) {
	play, _ := NewPlayList("http://playertest.longtailvideo.com/adaptive/bipbop/gear4/prog_index.m3u8", "test")
	err := play.prepare()
	play.parse()
	if err != nil {
		t.Error(err)
	}
	play.download()
	play.merge()
}
func TestDirTest(t *testing.T) {
	filepath.WalkDir("./", func(path string, fs fs.DirEntry, err error) error {
		if !fs.IsDir() {
			fmt.Println(path, fs.IsDir(), fs.Name())
		}
		return nil
	})
}
func TestMerge(t *testing.T) {
	var outputFile *os.File
	var err error
	if _, exists := os.Stat("/Users/xj/workspaces/hls/parser/output.mp4"); os.IsNotExist(exists) {
		// create outputfile
		outputFile, err = os.Create("/Users/xj/workspaces/hls/parser/output.mp4")
	} else {
		outputFile, err = os.CreateTemp("/Users/xj/workspaces/hls/parser", "output"+"_*"+".mp4")
	}

	if err != nil {
		t.Error(err)

	}
	defer outputFile.Close()

	walkerr := filepath.WalkDir("test_2662136025", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(path, ".ts") {
			f, err := os.Open(path)
			defer f.Close()
			if err != nil {
				return err
			}
			writebyte, copyerr := io.Copy(outputFile, f)
			t.Log("write", writebyte)
			if copyerr != nil {
				return copyerr
			}

		}
		return nil
	})
	if walkerr != nil {
		t.Error(walkerr)
	}
}
func TestFileStat(t *testing.T) {
	filename := "/Users/xj/workspaces/hls/parser/output.mp4"
	_, err := os.Stat(filename)
	fmt.Println(os.IsExist(err))
}
func TestScanner(t *testing.T) {
	rg := regexp.MustCompile(`(METHOD|URI|IV|KEYFORMAT)="?([^,"]+)"?`)
	rgs := rg.FindAllStringSubmatch(`METHOD="AES-128",IV=12`, -1)
	fmt.Println(rgs)
}
func TestPoint(t *testing.T) {
	s, _ := NewPlayList("http://baidu.com", "test")
	s.currentKey = &Key{
		FormatVersion: "1",
	}
	v := &s.currentKey
	s.currentKey = &Key{
		FormatVersion: "2",
	}
	fmt.Println(s.currentKey)
	fmt.Println(*v)
}
