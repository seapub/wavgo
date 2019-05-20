package wavgo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"time"
)

const (
	wavFormatPCM       = 1
	wavFormatIEEEFloat = 3
	wavs32le           = 0xfffe //  ffmpeg -i input.wav -c:a pcm_s32le -ar 48000 -ac 1 output.wav
)

// Header contains Wav fmt chunk data.
type Header struct {
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
}

// Wav reads wav files.
type Wav struct {
	Header
	// Samples is the total number of available samples.
	Samples int
	// Duration is the estimated duration based on reported samples.
	Duration time.Duration

	r io.Reader //io.Reader
}

// NewWav reads the WAV header from r.
func NewWav(r io.Reader) (*Wav, error) {
	var w Wav
	header := make([]byte, 16)
	if _, err := io.ReadFull(r, header[:12]); err != nil {
		return nil, err
	}
	if string(header[0:4]) != "RIFF" { // [0,3]
		return nil, fmt.Errorf("wav: missing RIFF")
	}
	if string(header[8:12]) != "WAVE" { // [8,11]
		return nil, fmt.Errorf("wav: missing WAVE")
	}
	hasFmt := false
	for {
		if _, err := io.ReadFull(r, header[:8]); err != nil { //[12,19]
			return nil, err
		}
		sz := binary.LittleEndian.Uint32(header[4:]) // [16,19]
		switch typ := string(header[:4]); typ {      //[12,15]
		case "fmt ": //[12,15]
			if sz < 16 {
				return nil, fmt.Errorf("wav: bad fmt size")
			}
			f := make([]byte, sz)
			if _, err := io.ReadFull(r, f); err != nil {
				return nil, err
			}
			if err := binary.Read(bytes.NewBuffer(f), binary.LittleEndian, &w.Header); err != nil {
				return nil, err
			}
			switch w.AudioFormat {
			case wavFormatPCM:
			case wavFormatIEEEFloat:
			case wavs32le:
			default:
				return nil, fmt.Errorf("wav: unknown audio format: %02x", w.AudioFormat)
			}
			hasFmt = true
		case "data":
			if !hasFmt {
				return nil, fmt.Errorf("wav: unexpected fmt chunk")
			}
			w.Samples = int(sz) / int(w.BitsPerSample) * 8
			w.Duration = time.Duration(w.Samples) * time.Second / time.Duration(w.SampleRate) / time.Duration(w.NumChannels)
			w.r = io.LimitReader(r, int64(sz))
			return &w, nil
		default:
			io.CopyN(ioutil.Discard, r, int64(sz))
		}
	}
}

// ReadSamples returns a [n]T, where T is uint8, int16, or float32, based on the
// wav data. n is the number of samples to return.
// only support `binary.LittleEndian`
func (w *Wav) ReadSamples(n int) (interface{}, error) {
	var data interface{}
	switch w.AudioFormat {
	case wavFormatPCM:
		switch w.BitsPerSample {
		case 8:
			data = make([]uint8, n)
		case 16:
			data = make([]int16, n)
		case 32:
			data = make([]int32, n)
		default:
			return nil, fmt.Errorf("wav: unknown bits per sample: %v", w.BitsPerSample)
		}
	case wavs32le: // seapub 针对ffmpeg定制
		data = make([]int32, n)
	case wavFormatIEEEFloat:
		data = make([]float32, n)
	default:
		return nil, fmt.Errorf("wav: unknown audio format")
	}
	if err := binary.Read(w.r, binary.LittleEndian, data); err != nil {
		return nil, fmt.Errorf("binary.Read Err:%s", err)
	}
	return data, nil
}

// ReadFloats is like ReadSamples, but it converts any underlying data to a float64.
func (w *Wav) ReadFloats(n int) ([]float64, error) {
	d, err := w.ReadSamples(n)
	if err != nil {
		return nil, fmt.Errorf("ReadSamples Err:%s", err)
	}
	var f []float64
	switch d := d.(type) {
	case []uint8:
		f = make([]float64, len(d))
		for i, v := range d {
			f[i] = float64(v) / math.MaxUint8
		}
	case []int16:
		f = make([]float64, len(d))
		for i, v := range d {
			// f[i] = (float64(v) - math.MinInt16) / (math.MaxInt16 - math.MinInt16) // unsigned
			f[i] = float64(v) / (math.MaxInt16 + 1) // seapub signed
		}
	case []int32: // seapub 定制 signed
		f = make([]float64, len(d))
		for i, v := range d {
			// f[i] = (float64(v) - math.MinInt32) / (math.MaxInt32 - math.MinInt32) // unsigned
			f[i] = float64(v) / (math.MaxInt32 + 1) // ffmpeg s32le 定制 190507
		}
	case []float32:
		f = make([]float64, len(d))
		for i, v := range d {
			f[i] = float64(v)
		}
	default:
		return nil, fmt.Errorf("wav: unknown type: %T", d)
	}
	return f, nil
}
