package main

import (
	"finedust"
	"flag"
	"path"
)

func main() {
	imgpath := flag.String("imgpath", "frames", "image sequence directory")
	binpath := flag.String("binpath", "tools/ffmpeg", "ffmpeg path")
	quality := flag.Int("quality", 4, "1:lossless")
	scale := flag.Int("scale", 720, "height scale")
	start := flag.String("start", "", "start image file ")
	end := flag.String("end", "", "end image file")

	flag.Parse()

	var err error
	err = finedust.RollbackImageSequence(*imgpath)
	if err != nil {
		panic(err)
	}
	err = finedust.RenameImageSequence(*imgpath, *start, *end)
	if err != nil {
		panic(err)
	}
	err = finedust.GenVideoByFFMPEG(*binpath, *imgpath, *quality, *scale)
	if err != nil {
		panic(err)
	}
	err = finedust.RollbackImageSequence(*imgpath)
	if err != nil {
		panic(err)
	}
}
