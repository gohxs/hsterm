package ansi_test

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/gohxs/termu/ansi"
)

type testValue struct {
	input  string
	expect ansi.Value
	comp   bool
}

var (
	keySeq = map[string]testValue{
		"private use": {
			input:  "\033[?25h",
			expect: ansi.Value{Raw: "\033[?25h", Type: ansi.TypeEscape, Attr: []int{25}, Value: "\033[?h"},
			comp:   true,
		},
		"modereset": {
			input:  "\033[?25;05l",
			expect: ansi.Value{Raw: "\033[?25;05l", Type: ansi.TypeEscape, Attr: []int{25, 5}, Value: "\033[?l"},
			comp:   true,
		},
		"Suffixes": {
			input:  "\033[31;3+m",
			expect: ansi.Value{Raw: "\033[31;3+m", Type: ansi.TypeEscape, Attr: []int{31, 3}, Value: "\033[+m"},
			comp:   true,
		},
		"fkey": {
			input:  "\033OP",
			expect: ansi.Value{Raw: "\033OP", Type: ansi.TypeMeta, Attr: nil, Value: "\033OP"},
			comp:   true,
		},
		"keyUp": {
			input:  "\033[A",
			expect: ansi.Value{Raw: "\033[A", Type: ansi.TypeEscape, Attr: nil, Value: "\033[A"},
			comp:   true,
		},
		"param": {
			input:  "\033[03;35m",
			expect: ansi.Value{Raw: "\033[03;35m", Type: ansi.TypeEscape, Attr: []int{3, 35}, Value: "\033[m"},
			comp:   true,
		},
		"fail": {
			input:  "\033[aa",
			expect: ansi.Value{Raw: "\033[a", Type: ansi.TypeEscape, Attr: []int{1}, Value: "a"},
			comp:   false,
		},
	}
)

func TestRead(t *testing.T) {
	for k, v := range keySeq {
		rd := ansi.NewReader(bytes.NewReader([]byte(v.input)))
		val, err := rd.ReadEscape()
		if err != nil {
			t.Fatal(err)
		}
		if reflect.DeepEqual(val, v.expect) != v.comp {
			t.Fatalf("Mismatch '%s' %#v==%#v %v", k, val, v.expect, v.comp)
		} else {
			t.Logf("Success '%s' %#v == %#v %v", k, val, v.expect, v.comp)
		}
	}
}

func TestReadString(t *testing.T) {
	testString := "\033[01;33mHello \033[01;35mTesting\033[3~\033[36m 指事字\033[01;35mok\033[m\n\033[?5htest est"

	rd := ansi.NewReader(bytes.NewReader([]byte(testString)))

	data := make([]byte, 40)

	n, err := rd.Read(data)
	t.Logf("Readed: '%d'", n)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	result := string(data[:n])
	t.Log("Original:", testString)
	t.Log("Result:", result)
	expect := "Hello Testing 指事字ok\ntest est"
	if result != expect {
		t.Fatalf("String does not match '%#v' != '%#v'", result, expect)
	}

}
func TestSingleEsc(t *testing.T) {
	br := bytes.NewBuffer([]byte("\033"))
	rd := ansi.NewReader(br)
	r, err := rd.ReadEscape()
	t.Logf("%#v %v", r, err)

}
