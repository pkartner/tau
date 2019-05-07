package tau

// Modified from strings.Reader
type wrappingStringReader struct {
	s string
	i int64 // current reading index
}

func (r *wrappingStringReader) Read(b []byte) (n int, err error) {
	for n < len(b) {
		n += copy(b, r.s[r.i:])
		r.i += int64(n)
		if r.i >= int64(len(r.s)) {
			r.i = 0
		}
	}
	return
}

func newWrappingStringReader(s string) *wrappingStringReader {
	return &wrappingStringReader{
		s: s,
	}
}
