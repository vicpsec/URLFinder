package main

import (
	"io"
	"log"

	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/crawler"
	"github.com/pingc0y/URLFinder/util"
)

func main() {
	log.SetOutput(io.Discard)
	util.GetUpdate()
	cmd.Parse()
	crawler.Run()
}
