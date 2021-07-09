package cmd

import (
	"flag"
)

var targetURL string
var output string

func init() {
	// target url
	flag.StringVar(&targetURL, "url", "", "address of target m3u8 file.")
	// output file name
	flag.StringVar(&output, "file", "output", "output file name.")
}

func Run() {

}
