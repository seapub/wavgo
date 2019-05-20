package main

import (
	"flag"
	"fmt"

	"github.com/seapub/wavgo"
)

var srcPath string
var dstDir string
var barEnergy float64
var spanSilence int64 // ms
var spanMargin int64  // ms
var spanMin int64     // ms

func init() {
	flag.StringVar(&srcPath, "srcPath", ".", "srcPath")
	flag.StringVar(&dstDir, "dstDir", "./tmp", "dstDir")
	flag.Float64Var(&barEnergy, "barEnergy", 500, "barEnergy")
	flag.Int64Var(&spanSilence, "spanSilence", 500, "spanSilence")
	flag.Int64Var(&spanMargin, "spanMargin", 200, "spanMargin")
	flag.Int64Var(&spanMin, "spanMin", 400, "spanMin")
}

func main() {
	flag.Parse()
	var err error
	// res, err := wavgo.SplitWav(srcPath, wavgo.SplitArgs{
	// 	BarEnergy:   barEnergy,
	// 	SpanSilence: spanSilence,
	// 	SpanMargin:  spanMargin,
	// 	SpanMin:     spanMin,
	// })
	// fmt.Printf("%v\n err:%v\n", res.NotEmpty, err)
	fmt.Println(srcPath, dstDir)
	err = wavgo.SplitSavWav(srcPath, dstDir, wavgo.SplitArgs{
		BarEnergy:   barEnergy,
		SpanSilence: spanSilence,
		SpanMargin:  spanMargin,
		SpanMin:     spanMin,
	})
	if err != nil {
		fmt.Printf("%s:\t%s\n", srcPath, err)
	} else {
		fmt.Printf("%s:\tDONE\n", srcPath)
	}
}

/*
gor splitwav.go \
-srcPath="/Volumes/seagate8/Gmark2/HANGUP.wav" \
-dstDir="/Volumes/seagate8/Gmark2/output" \
-barEnergy="0.000036" \
-spanSilence="800" \
-spanMargin="400" \
-spanMin="200"

splitwav -srcPath="/Volumes/seagate8/Gmark2/HANGUP.wav" \
-dstDir="/Volumes/seagate8/Gmark2/output" \
-barEnergy="0.000036" \
-spanSilence="800" \
-spanMargin="400" \
-spanMin="200"
*/
