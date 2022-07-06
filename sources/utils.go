package sources

import "unicode/utf8"

func getValidUtf8String(in string)(string) {

	if !utf8.ValidString(in){
		v := make([]rune, 0, len(in))
		for i, r := range in {
				if r == utf8.RuneError {
						_, size := utf8.DecodeRuneInString(in[i:])
						if size == 1 {
								continue
						}
				}
				v = append(v, r)
		}
		in = string(v)
	}

	return in
} 