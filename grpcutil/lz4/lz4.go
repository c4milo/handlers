package lz4

import (
	"io"
	"io/ioutil"

	"github.com/pierrec/lz4"
	"google.golang.org/grpc"
)

// Compressor compresses gRPC messages using LZ4
type compressor struct{}

func (c *compressor) Do(w io.Writer, p []byte) error {
	l := lz4.NewWriter(w)
	if _, err := l.Write(p); err != nil {
		return err
	}
	return l.Close()
}

func (c *compressor) Type() string {
	return "lz4"
}

// NewCompressor returns a new LZ4 compressor instance.
func NewCompressor() grpc.Compressor {
	return &compressor{}
}

// Decompressor decompresses gRPC messages using LZ4
type decompressor struct{}

// NewDecompressor returns a new LZ4 decompressor instance.
func NewDecompressor() grpc.Decompressor {
	return &decompressor{}
}

func (c *decompressor) Do(r io.Reader) ([]byte, error) {
	lr := lz4.NewReader(r)
	return ioutil.ReadAll(lr)
}

func (c *decompressor) Type() string {
	return "lz4"
}
