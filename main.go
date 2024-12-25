package main

import (
	"io"
	"log"

	"github.com/vicpsec/URLFinder/cmd"
	"github.com/vicpsec/URLFinder/crawler"
	"github.com/vicpsec/URLFinder/util"
)

func main() {
	log.SetOutput(io.Discard)
	util.GetUpdate()
	cmd.Parse()
	crawler.Run()
}
