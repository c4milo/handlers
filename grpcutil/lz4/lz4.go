package lz4

import (
	"io"
	"io/ioutil"
	"sync"

	"github.com/pierrec/lz4"
	"google.golang.org/grpc"
)

// Compressor compresses gRPC messages using LZ4
type compressor struct {
	pool sync.Pool
}

// NewCompressor returns a new LZ4 compressor instance.
func NewCompressor() grpc.Compressor {
	return &compressor{
		pool: sync.Pool{
			New: func() interface{} {
				return lz4.NewWriter(ioutil.Discard)
			},
		},
	}
}

func (c *compressor) Do(w io.Writer, p []byte) error {
	lzw := c.pool.Get().(*lz4.Writer)
	defer c.pool.Put(lzw)
	lzw.Reset(w)

	if _, err := lzw.Write(p); err != nil {
		return err
	}
	return lzw.Close()
}

func (c *compressor) Type() string {
	return "lz4"
}

// Decompressor decompresses gRPC messages using LZ4
type decompressor struct {
	pool sync.Pool
}

// NewDecompressor returns a new LZ4 decompressor instance.
func NewDecompressor() grpc.Decompressor {
	return &decompressor{
		pool: sync.Pool{
			New: func() interface{} {
				return lz4.NewReader(nil)
			},
		},
	}
}

func (d *decompressor) Do(r io.Reader) ([]byte, error) {
	lzr := d.pool.Get().(*lz4.Reader)
	defer d.pool.Put(lzr)

	lzr.Reset(r)

	return ioutil.ReadAll(lzr)
}

func (d *decompressor) Type() string {
	return "lz4"
}
