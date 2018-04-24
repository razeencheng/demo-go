package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var buf = []byte("skdjadialsdgasadasdhsakdjsahlskdjagloqweiqwo")

func TestEqual(t *testing.T) {
	should := require.New(t)
	should.Equal(EncodeA(buf), EncodeB(buf))
}

func BenchmarkEncodeA(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EncodeA(buf)
	}
}

func BenchmarkEncodeB(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EncodeB(buf)
	}
}
