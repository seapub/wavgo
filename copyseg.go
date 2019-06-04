package wavgo

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopySeg 保存
func CopySeg(srcPath, dstPath string,
	start, end float64) error {
	if start > end || start < 0 || end < 0 {
		return errors.New("invalid args")
	}

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

	framerate := float64(f.SampleRate)
	numFrames := float64(f.Samples)
	if start*framerate > numFrames || end*framerate > numFrames {
		return errors.New("invalid args: start/end too big")
	}

	_, err = f.ReadFloats(int(start * framerate)) // todo 优化效率
	if err != nil {
		return err
	}

	fo, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("Open(dstPath) %v", err)
	}
	defer fo.Close()

	w, err := NewWavWriter(fo,
		uint32((end-start)*framerate),
		f.NumChannels,
		f.SampleRate,
		f.BitsPerSample,
	)
	if err != nil {
		return fmt.Errorf("wav.NewWavWriter %s", err)
	}
	samples, errR := f.ReadFloats(int((end - start) * framerate)) // 不能读取太多 ReadSamples
	if errR != nil && errR != io.EOF {
		return fmt.Errorf("ReadFloats %s", errR)
	}
	err = w.WriteFloats(samples) // WriteSamples
	if err != nil {
		return fmt.Errorf("WriteSamples %s", err)
	}

	return nil
}
