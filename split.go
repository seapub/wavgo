package wavgo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// SplitArgs 音频切割所需的参数
type SplitArgs struct {
	BarEnergy   float64
	SpanSilence int64 // ms
	SpanMargin  int64 // ms
	SpanMin     int64 // ms
}

// SplitRes .
type SplitRes struct {
	FramePerWindows int64
	TimePerWindows  int64 // 一个音素为20ms
	Energy          []float64
	OverFlowCnt     []int64
	NotEmpty        [][2]int64
}

// energy Calculate the energy of the audio frames
func energy(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := float64(0.0)
	for _, v := range data {
		sum += float64(v * v)
	}
	return sum / float64(len(data))
}

func overflowCnt(bitsPerFrame int64, data []float64) int64 {
	res := int64(0)
	for _, v := range data {
		switch bitsPerFrame {
		case 8: // math.MaxUint8
			if v > 0.999999999 || v < -0.999999999 {
				res++
			}
		case 16: // math.MaxInt8 / (math.MaxInt16 + 1)
			if v > 0.999999999 || v < -0.999999999 {
				res++
			}
		case 24:
			if v > 0.999999 || v < -0.999999999 {
				res++
			}
		case 32:
			if v > 0.999999999 || v < -0.999999999 {
				res++
			}
		default:
			if v > 0.999999999 || v < -0.999999999 {
				res++
			}
		}
	}
	return res
}

// SplitSavWav split and save
func SplitSavWav(srcPath, outDir string, splitArgs SplitArgs) error {
	splitRes, err := SplitWav(srcPath, splitArgs)
	if err != nil {
		return fmt.Errorf("SplitWav Err:%s", err)
	}
	return saveSplitRes(srcPath, outDir, splitArgs, splitRes)
}

// saveSplitRes 保存
func saveSplitRes(srcPath, dstDir string,
	splitArgs SplitArgs, splitRes *SplitRes) error {

	if filepath.Ext(srcPath) != ".wav" {
		return fmt.Errorf("format not supported")
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		fmt.Printf("Open(srcPath) %v\n", err)
		return err
	}
	defer srcFile.Close()

	f, err := NewWav(srcFile)
	if err != nil {
		return err
	}

	// mkdir
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		// mask := syscall.Umask(0)
		os.MkdirAll(dstDir, 0777)
		// syscall.Umask(mask)
	}

	ptrNow := int64(0) // init
	// oldName := path.Base(srcPath)         // Get the original filename
	oldName := filepath.Base(srcPath)     // Get the original filename
	oldName = oldName[0 : len(oldName)-4] // The suffix of the filename is `.wav`

	for i, v := range splitRes.NotEmpty {
		var ofHint string //  As a sign of overflow
		if splitRes.OverFlowCnt[i] != 0 {
			ofHint = fmt.Sprintf("overflow%d", splitRes.OverFlowCnt[i]) // hard coding
		} else {
			ofHint = "0"
		}
		dstPath := filepath.Join(dstDir, fmt.Sprintf("%s_%s_%08d.wav", oldName, ofHint, i+1))

		var padding []float64
		var paddingE []float64
		frameOfMargin := int64(splitArgs.SpanMargin) * int64(f.SampleRate) / 1000
		ptrBegin := v[0]*splitRes.FramePerWindows - frameOfMargin
		if ptrNow < ptrBegin {
			f.ReadFloats(int(ptrBegin - ptrNow)) // 调整坐标
		} else { //
			padding = make([]float64, ptrNow-ptrBegin) // default value is 0
			ptrBegin = ptrNow
		}

		ptrEnd := v[1] * splitRes.FramePerWindows
		if ptrEnd+frameOfMargin < int64(f.Samples) {
			ptrEnd = ptrEnd + frameOfMargin
		} else if ptrEnd < int64(f.Samples) {
			paddingE = make([]float64, frameOfMargin-(int64(f.Samples)-ptrEnd))
			ptrEnd = int64(f.Samples)
		} else {
			paddingE = make([]float64, frameOfMargin)
			ptrEnd = int64(f.Samples)
		}
		if ptrEnd <= ptrBegin {
			continue
		}

		fo, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			return fmt.Errorf("Open(dstPath) %v", err)
		}
		// defer fo.Close() // don't uses defer in for loop

		w, err := NewWavWriter(fo,
			uint32(ptrEnd-ptrBegin)+
				uint32(len(padding)+len(paddingE)),
			f.NumChannels,
			f.SampleRate,
			f.BitsPerSample,
		)
		if err != nil {
			fo.Close()
			return fmt.Errorf("wav.NewWavWriter %s", err)
		}
		if len(padding) != 0 { //
			err = w.WriteFloats(padding) // WriteSamples
			if err != nil {
				fo.Close()
				return fmt.Errorf("wavgo.WriteFloats %s", err)
			}
		}
		samples, errR := f.ReadFloats(int(ptrEnd - ptrBegin)) // 不能读取太多 ReadSamples
		if errR != nil && errR != io.EOF {
			fo.Close()
			return fmt.Errorf("ReadFloats %s", errR)
		}
		err = w.WriteFloats(samples) // WriteSamples
		if err != nil {
			fo.Close()
			return fmt.Errorf("WriteSamples %s", err)
		}
		if len(paddingE) != 0 {
			err = w.WriteFloats(paddingE) // WriteSamples
			if err != nil {
				fo.Close()
				return fmt.Errorf("wavgo.WriteFloats %s", err)
			}
		}

		fo.Close()
		if errR == io.EOF {
			return nil
		}
		ptrNow = ptrEnd
	}

	return nil
}

// SplitWav .
func SplitWav(srcPath string, splitArgs SplitArgs) (*SplitRes, error) {
	_, err := os.Stat(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not exist %s", err)
		}
		return nil, fmt.Errorf("os.Stat Err:%s", err)
	}
	splitRes, err := EnergySlice(srcPath)
	if err != nil {
		return nil, fmt.Errorf("EnergySlice Err:%s", err)
	}
	// fmt.Printf("len(splitRes.Energy):%d", len(splitRes.Energy))
	if len(splitRes.Energy) == 0 {
		return nil, fmt.Errorf("read file error or file is empty")
	}
	// 利用空白切割
	splitRes = notEmpty(splitRes, splitArgs)
	// 过滤最长
	splitRes = filterSpanMin(splitRes, splitArgs)

	return splitRes, nil
}

// filterSpanMin .
func filterSpanMin(splitRes *SplitRes, splitArgs SplitArgs) *SplitRes {
	spanMin := splitArgs.SpanMin / splitRes.TimePerWindows // windows per spanMin
	notEmptyLen := int64(len(splitRes.NotEmpty))
	notEmpty := make([][2]int64, 0, notEmptyLen)
	for i := int64(0); i < notEmptyLen; i++ {
		if splitRes.NotEmpty[i][1]-splitRes.NotEmpty[i][0] > spanMin {
			notEmpty = append(notEmpty, splitRes.NotEmpty[i])
		}
	}
	splitRes.NotEmpty = notEmpty
	return splitRes
}

// NotEmpty .
func notEmpty(splitRes *SplitRes, splitArgs SplitArgs) *SplitRes {
	enAve := float64(0)
	for _, i := range splitRes.Energy {
		enAve += i
	}
	enAve /= float64(len(splitRes.Energy))

	barEnergy := splitArgs.BarEnergy                                 //
	winPerSilence := splitArgs.SpanSilence / splitRes.TimePerWindows // unit: 20ms

	notEmptyS := make([]int64, 0, 8) // start. point to first non-empty window
	notEmptyE := make([]int64, 0, 8) // end. point to windwo next to last  non-empty window  指向尾部第一个空白
	i := 0
	for ; i < len(splitRes.Energy); i++ {
		if splitRes.Energy[i] > barEnergy {
			notEmptyS = append(notEmptyS, int64(i))
			break
		}
	}
	emptylen := int64(0)
	for ; i < len(splitRes.Energy); i++ {
		if splitRes.Energy[i] < barEnergy {
			emptylen++
		} else {
			if emptylen > winPerSilence {
				notEmptyE = append(notEmptyE, int64(i)-emptylen)
				notEmptyS = append(notEmptyS, int64(i))
			}
			emptylen = 0
		}
	}
	if len(notEmptyS) != 0 {
		if emptylen != 0 {
			notEmptyE = append(notEmptyE, int64(i)-emptylen)
		} else {
			notEmptyE = append(notEmptyE, int64(i))
		}
	}

	notEmpty := make([][2]int64, len(notEmptyE))
	for i := 0; i < len(notEmpty); i++ {
		notEmpty[i] = [2]int64{notEmptyS[i], notEmptyE[i]}
	}
	splitRes.NotEmpty = notEmpty
	return splitRes
}

// EnergySlice 计算wav文件的能量变化 20ms为一个音素
func EnergySlice(srcPath string) (*SplitRes, error) {
	srcExt := filepath.Ext(srcPath)
	if strings.ToLower(srcExt) != ".wav" {
		return nil, fmt.Errorf("format not supported")
	}
	// 打开文件
	file, err := os.Open(srcPath)
	if err != nil {
		return nil, fmt.Errorf("Open(srcPath) Err:%s", err)
	}
	defer file.Close()
	// wav
	f, err := NewWav(file)
	if err != nil {
		fmt.Printf("SplitWav New(file) %#v", err)
		return nil, fmt.Errorf("NewWav:%s", err)
	}
	bitsPerFrame := int64(f.BitsPerSample)
	framerate := int64(f.SampleRate)
	numFrames := int64(f.Samples)
	framePerWindows := framerate * 20 / 1000 // 20ms

	// 用于存储能量值的数组 (11,20)->(20,29)->2
	res := make([]float64, 0, (numFrames+framePerWindows-1)/framePerWindows)
	ofCnt := make([]int64, 0, len(res))
	for i := int64(0); i < numFrames-framePerWindows; i = i + framePerWindows {
		windowT, err := f.ReadFloats(int(framePerWindows))
		if err != nil {
			return nil, fmt.Errorf("ReadFloats Err:%s", err)
		}
		res = append(res, energy(windowT))
		ofCnt = append(ofCnt, overflowCnt(bitsPerFrame, windowT))
	}
	return &SplitRes{
		FramePerWindows: framePerWindows,
		TimePerWindows:  20,
		Energy:          res,
		OverFlowCnt:     ofCnt,
	}, nil
}
