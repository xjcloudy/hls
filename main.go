package main

import (
	"flag"
	"hls/cmd"
)

func main(){
	flag.Parse()
	cmd.Run()
}
