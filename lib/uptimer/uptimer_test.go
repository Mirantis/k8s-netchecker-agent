package uptimer

import (
	"testing"
)

func Test_getUptimeString(t *testing.T) {
	oldReadFile := readFile
	defer func() { readFile = oldReadFile }()
	readFile = func(filename string) ([]byte, error) {
		return []byte("888.999 333.222"), nil
	}

	u := NewUptimer()
	if r := u.getUptimeString(); "888.999" != r {
		t.Errorf("returned '%s', should be 888.999", r)
	}
}

func Test_Get_(t *testing.T) {
	oldReadFile := readFile
	defer func() { readFile = oldReadFile }()
	readFile = func(filename string) ([]byte, error) {
		return []byte("888.999 333.222"), nil
	}

	u := NewUptimer()
	if r := u.Get(); 888 != r {
		t.Errorf("returned %d, should be 888", r)
	}
	if r := u.GetFloat(); 888.999 != r {
		t.Errorf("returned %d, should be 888.999", r)
	}
}
