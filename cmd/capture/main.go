package main

import (
	"finedust"
	"fmt"
)

func main() {
	fmt.Println("start finedust capture")
	conf := &finedust.Config{}
	err := conf.Load("config.json")
	if err != nil {
		panic(err)
	}
	err = finedust.CaptureBySelenium(*conf)
	if err != nil {
		panic(err)
	}
	fmt.Println("done finedust capture")
}
