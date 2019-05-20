package wavgo

import (
	"fmt"
	"io"
	"math"
	"os"
)

// 边界情况
// 2147483647 2147483647
// -2147483648 -2147483648
func edgeTest() {
	a := math.MaxInt32

	af := float64(a) / (math.MaxInt32 + 1)
	a2 := int32(af * (math.MaxInt32 + 1))
	fmt.Println(a, a2)

	b := math.MinInt32

	bf := float64(b) / (math.MaxInt32 + 1)
	b2 := int32(bf * (math.MaxInt32 + 1))
	fmt.Println(b, b2)
}

func foo2(srcPath, dstPath string) {
	file, err := os.Open(srcPath)
	if err != nil {
		fmt.Printf("Open(srcPath) %v\n", err)
		return
	}
	defer file.Close()
	f, err := NewWav(file)
	if err != nil {
		fmt.Printf("%#v", err)
	}

	fo, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("Open(dstPath) %v\n", err)
		return
	}
	defer fo.Close()
	w, err := NewWavWriter(fo, uint32(f.Samples), f.NumChannels, f.SampleRate, f.BitsPerSample)
	if err != nil {
		fmt.Println("NewWavWriter", err)
	}

	fmt.Printf("%v, %v, %v\n", f.Header, f.Samples, f.Duration)
	for {
		samples, err := f.ReadFloats(2) // 不能读取太多 ReadSamples
		if err == io.EOF {
			fmt.Printf("EOF\n")
			break
		}

		err = w.WriteFloats(samples) // WriteSamples
		if err != nil {
			fmt.Println("WriteSamples", err)
		}
	}
}
