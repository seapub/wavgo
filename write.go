package wavgo

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// Writer .
type Writer struct {
	w      io.Writer
	Format *Header
}

// NewWavWriter .
func NewWavWriter(w io.Writer,
	numSamples uint32,
	numChannels uint16,
	sampleRate uint32,
	bitsPerSample uint16) (
	*Writer, error) {
	// 8 16 : wavFormatPCM   24,32:wavFormatIEEEFloat
	var wavFormat uint16
	if bitsPerSample <= 16 {
		wavFormat = wavFormatPCM
	} else {
		wavFormat = wavFormatIEEEFloat
	}

	blockAlign := numChannels * bitsPerSample / 8
	byteRate := sampleRate * uint32(blockAlign)
	format := &Header{
		AudioFormat:   wavFormat, // wavFormatPCM,
		NumChannels:   numChannels,
		SampleRate:    sampleRate,
		ByteRate:      byteRate,
		BlockAlign:    blockAlign,
		BitsPerSample: bitsPerSample,
	}
	dataSize := numSamples * uint32(format.BlockAlign)
	riffSize := uint32(4 + 8 + 16 + 8 + dataSize)

	_, err := w.Write([]byte("RIFF"))
	if err != nil {
		return nil, err
	}
	err = binary.Write(w, binary.LittleEndian, riffSize)
	if err != nil {
		fmt.Printf("a21 \n")
		return nil, err
	}
	_, err = w.Write([]byte("WAVE"))
	if err != nil {
		return nil, err
	}

	_, err = w.Write([]byte("fmt "))
	if err != nil {
		return nil, err
	}
	err = binary.Write(w, binary.LittleEndian, uint32(16))
	if err != nil {
		fmt.Printf("a %s\n", err)
		return nil, err
	}
	err = binary.Write(w, binary.LittleEndian, format)
	if err != nil {
		return nil, err
	}

	_, err = w.Write([]byte("data"))
	if err != nil {
		fmt.Printf("b \n")
		return nil, err
	}
	err = binary.Write(w, binary.LittleEndian, dataSize)
	if err != nil {
		fmt.Printf("c \n")
		return nil, err
	}

	writer := &Writer{
		w:      w,
		Format: format,
	}
	return writer, nil
}

// WriteSamples .
func (w *Writer) WriteSamples(data interface{}) error {
	switch w.Format.AudioFormat {
	case wavFormatPCM:
		switch w.Format.BitsPerSample {
		case 8:
			if data1, ok := data.([]uint8); ok {
				return binary.Write(w.w, binary.LittleEndian, data1)
			}
			return fmt.Errorf("type error 8")
		case 16:
			if data1, ok := data.([]int16); ok {
				return binary.Write(w.w, binary.LittleEndian, data1)
			}
			return fmt.Errorf("type error 16")
		case 32:
			if data1, ok := data.([]int32); ok {
				return binary.Write(w.w, binary.LittleEndian, data1)
			}
			return fmt.Errorf("type error 32")
		default:
			return fmt.Errorf("wav: unknown bits per sample: %v", w.Format.BitsPerSample)
		}
	case wavs32le:
		if data1, ok := data.([]int32); ok {
			return binary.Write(w.w, binary.LittleEndian, data1)
		}
		return fmt.Errorf("type error pcms32le")
	case wavFormatIEEEFloat:
		if data1, ok := data.([]float32); ok {
			return binary.Write(w.w, binary.LittleEndian, data1)
		}
		return fmt.Errorf("type error 3")
	default:
		return fmt.Errorf("wav: unknown audio format")
	}
}

// WriteFloats is like ReadSamples, but it converts any underlying data to a float64
// 目前未测试 8bit 宽度的wav
func (w *Writer) WriteFloats(data []float64) error {
	switch w.Format.AudioFormat {
	case wavFormatPCM:
		switch w.Format.BitsPerSample {
		case 8:
			data1 := make([]uint8, len(data))
			for i, v := range data {
				data1[i] = uint8(v * math.MaxUint8)
			}
			return binary.Write(w.w, binary.LittleEndian, data1)
		case 16:
			data1 := make([]int16, len(data))
			for i, v := range data {
				data1[i] = int16(v * (math.MaxInt16 + 1)) // signed
				// data1[i] = int16(v*(math.MaxInt16-math.MinInt16)) + math.MinInt16 // unsigned
			}
			return binary.Write(w.w, binary.LittleEndian, data1)
		case 32:
			data1 := make([]int32, len(data))
			for i, v := range data {
				// data1[i] = int32(v*(math.MaxInt32-math.MinInt32)) + math.MinInt32
				data1[i] = int32(v * (math.MaxInt32 + 1)) // signed
			}
			return binary.Write(w.w, binary.LittleEndian, data1)
		default:
			return fmt.Errorf("wav: unknown bits per sample: %v", w.Format.BitsPerSample)
		}
	case wavs32le:
		data1 := make([]int32, len(data))
		for i, v := range data {
			// data1[i] = int32(v*(math.MaxInt32-math.MinInt32)) + math.MinInt32 // unsigned
			data1[i] = int32(v * (math.MaxInt32 + 1)) // signed
		}
		return binary.Write(w.w, binary.LittleEndian, data1)
	case wavFormatIEEEFloat:
		data1 := make([]float32, len(data))
		for i, v := range data {
			data1[i] = float32(v)
		}
		return binary.Write(w.w, binary.LittleEndian, data1)
	default:
		return fmt.Errorf("wav: unknown audio format")
	}
}
