//go:build bench

package hw10programoptimization

import (
	"archive/zip"
	"io"
	"testing"
)

func BenchmarkGetDomainStat(b *testing.B) {
	r, err := zip.OpenReader("testdata/users.dat.zip")
	if err != nil {
		b.Fatal(err)
	}
	defer r.Close()

	data, err := r.File[0].Open()
	if err != nil {
		b.Fatal(err)
	}

	content, err := io.ReadAll(data)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		GetDomainStat(readerFromBytes(content), "biz")
	}
}

func readerFromBytes(b []byte) io.Reader {
	return &byteReader{data: b, pos: 0}
}

type byteReader struct {
	data []byte
	pos  int
}

func (r *byteReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
