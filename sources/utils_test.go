package sources

import (
	"testing"
	"unicode/utf8"
)

func Test_getValidUtf8String(t *testing.T){

	in  := "a{b=\"\xff\"} 1\n# EOF\n"
	expected := "a{b=\"\"} 1\n# EOF\n"
	out := getValidUtf8String(in)
	if out != expected {
		t.Fatalf("Retrieved an unexpected string. Expected: %s, Got: %s", expected, out)
	}

	if !utf8.ValidString(out){
		t.Fatalf("Retrieved an unexpected utf-8 string")
	}
}
