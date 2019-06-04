package wavgo

import "testing"

func TestCopySeg(t *testing.T) {
	var err error
	err = CopySeg("./test_float.wav", "./test_float_seg.wav", 2, 8)
	if err != nil {
		t.Error(err)
	}
	err = CopySeg("./libsbvYixCTxPTc0ahBod_VjUZye.wav", "./test_float_seg.wav", 2, 8)
	if err != nil {
		t.Error(err)
	}
}
