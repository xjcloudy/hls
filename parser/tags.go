package parser

const (
	// basic tags
	EXTM3U  = "#EXTM3U"
	VERSION = "#EXT-X-VERSION"

	// media or master playlist tags
	INDEPENDENT_SEGMENTS = "#EXT-X-INDEPENDENT-SEGMENTS"
	START                = "#EXT-X-START"
	DEFINE               = "#EXT-X-DEFINE"

	// media playlist tags
	TARGETDURATION          = "#EXT-X-TARGETDURATION"
	MEDIA_SEQUENCE          = "#EXT-X-MEDIA-SEQUENCE"
	DISCOUNTINUITY_SEQUENCE = "#EXT-X-DISCONTINUITY-SEQUENCE"
	ENDLIST                 = "#EXT-X-ENDLIST"
	PLAY_LIST_TYPE          = "#EXT-X-PLAYLIST-TYPE"
	I_FRAMES_ONLY           = "#EXT-X-I-FRAMES-ONLY"
	PART_INF                = "#EXT-X-PART-INF"
	SERVER_CONTROL          = "#EXT-X-SERVER-CONTROL"

	// media segment tags
	EXTINF            = "#EXTINF"
	BYTERANGE         = "#EXT-X-BYTERANGE"
	DISCONTINUITY     = "#EXT-X-DISCONTINUITY"
	KEY               = "#EXT-X-KEY"
	MAP               = "#EXT-X-MAP"
	PROGRAM_DATE_TIME = "#EXT-X-PROGRAM-DATE-TIME"
	GAP               = "#EXT-X-GAP"
)

// PlayListTypeEnum https://tools.ietf.org/html/draft-pantos-hls-rfc8216bis-08#page-19
type PlayListTypeEnum string

const (
	PLAY_LIST_TYPE_VOD   PlayListTypeEnum = "VOD"
	PLAY_LIST_TYPE_EVENT PlayListTypeEnum = "EVENT"
)

type KeyMethodEnums string

const (
	KEY_METHOD_NONE   = "NONE"
	KEY_METHOD_AES128 = "AES-128"
	KEY_METHOD_SAMPLE = "SAMPLE-AES"
)

// Key https://tools.ietf.org/html/draft-pantos-hls-rfc8216bis-08#page-24
type Key struct {
	Method        KeyMethodEnums
	URI           string
	IV            uint64
	Format        string
	FormatVersion string
}

// StartPoint https://tools.ietf.org/html/draft-pantos-hls-rfc8216bis-08#page-16
type StartPoint struct {
	TimeOffset float64
	Precise    bool
}

// Define https://tools.ietf.org/html/draft-pantos-hls-rfc8216bis-08#page-17
type Define struct {
	Name   string
	Value  string
	Import string
}

const DEFAULT_OUTPUT_FILENAME = "output"
