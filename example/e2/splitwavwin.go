package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/seapub/wavgo"
)

var srcPath string
var dstDir string
var barEnergy float64
var spanSilence int64 // ms
var spanMargin int64  // ms
var spanMin int64     // ms

func main() {
	// fmt.Printf("%#v\n", os.Args)
	// parse args
	if len(os.Args) != 7 {
		fmt.Println("Usage: ./splitwav 0.000036 800 400 200 srcPath dstDir")
		os.Exit(-2)
	}
	var err error
	barEnergy, err := strconv.ParseFloat(os.Args[1], 10)
	if err != nil {
		fmt.Printf("args Err: %s\n", err)
		os.Exit(-3)
	}
	spanSilence, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		fmt.Printf("args Err: %s\n", err)
		os.Exit(-3)
	}
	spanMargin, err := strconv.ParseInt(os.Args[3], 10, 64)
	if err != nil {
		fmt.Printf("args Err: %s\n", err)
		os.Exit(-3)
	}
	spanMin, err := strconv.ParseInt(os.Args[4], 10, 64) // 覆盖 spanMin?
	if err != nil {
		fmt.Printf("args Err: %s\n", err)
		os.Exit(-3)
	}
	srcPath = strings.Trim(os.Args[5], " ")
	dstDir = strings.Trim(os.Args[6], " ")

	// res, err := wavgo.SplitWav(srcPath, wavgo.SplitArgs{
	// 	BarEnergy:   barEnergy,
	// 	SpanSilence: spanSilence,
	// 	SpanMargin:  spanMargin,
	// 	SpanMin:     spanMin,
	// })
	// fmt.Printf("%v\n err:%v\n", res.NotEmpty, err)
	// fmt.Println(srcPath, dstDir)
	err = wavgo.SplitSavWav(srcPath, dstDir, wavgo.SplitArgs{
		BarEnergy:   barEnergy,
		SpanSilence: spanSilence,
		SpanMargin:  spanMargin,
		SpanMin:     spanMin,
	})
	if err != nil {
		fmt.Printf("%s:\t%s\n", srcPath, err)
	} else {
		// fmt.Printf("%s:\tDONE\n", srcPath)
		fmt.Println("ok")
	}
}

/*
GOOS=windows GOARCH=amd64 go build splitwavwin.go

./splitwavwin 0.000036 800 400 200 srcPath dstDir

./splitwavwin 0.000036 800 400 200 HEYTICO.wav ./output
*/
