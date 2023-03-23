package main

import (
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/huntresslabs/win-service-updater/updater"
)

func main() {
	args := updater.Args{}
	info := updater.Info{}
	wys, err := info.ParseWYSFromFilePath(os.Args[1], args)
	if nil != err {
		log.Fatal(err)
	}
	spew.Dump("%+v", wys)
}
