package parquet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/emb/play/parquet/meta"
)

var magic = []byte("PAR1")

type readAtCloser interface {
	io.ReaderAt
	io.Closer
}

// Parquet describes a parquet file
type Parquet struct {
	readAtCloser
	size int64
}

func NewParquetFile(name string) (*Parquet, error) {
	pfile, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	info, err := pfile.Stat()
	if err != nil {
		return nil, err
	}

	p := &Parquet{readAtCloser: pfile, size: info.Size()}
	ok, err := p.Valid()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("invalid parquet: expecting magic %q number in header/footer", magic)
	}
	return p, nil
}

// readMagicAt is supposed to be a helper but really it is just a wrapper and not a good one either.
func (p *Parquet) readMagicAt(offset int64) ([]byte, error) {
	buf := make([]byte, 4)
	n, err := p.ReadAt(buf, offset)
	if err != nil {
		return nil, fmt.Errorf("read magic at offset %d: %s", err)
	}
	if n != len(buf) {
		return nil, fmt.Errorf("read magic read %d bytes instead 4", n)
	}
	return buf, nil
}

// Valid checks the header and footer for the parquet magic number
func (p *Parquet) Valid() (bool, error) {
	header, err := p.readMagicAt(0)
	if err != nil {
		return false, err
	}
	// The last f
	footer, err := p.readMagicAt(p.size - 4)
	if err != nil {
		return false, err
	}
	return bytes.Equal(magic, header) && bytes.Equal(magic, footer), nil
}

func (p *Parquet) metaDataSize() (uint32, error) {
	buf := make([]byte, 4)
	// The footer contains 4 bytes of metadata length and 4 bytes of magic number
	n, err := p.ReadAt(buf, p.size-8)
	if err != nil {
		return 0, err
	}

	if n != 4 {
		return 0, fmt.Errorf("metaDataSize: partial read")
	}

	return binary.LittleEndian.Uint32(buf), nil
}

// MetaData reads the footer and decodes FileMetaData using thrift TCompactProtocol
func (p *Parquet) MetaData() (*meta.FileMetaData, error) {
	metaSizeUint32, err := p.metaDataSize()
	if err != nil {
		return nil, err
	}
	metaSize := int64(metaSizeUint32)
	buf := make([]byte, metaSize)
	n, err := p.ReadAt(buf, p.size-metaSize-8) // 8 (footer) = 4 (magic number) + 4 (metadata size)
	if err != nil {
		return nil, err
	}
	// TODO is this necessary?
	if int64(n) != metaSize {
		return nil, fmt.Errorf("metaData: partial read")
	}

	rs := io.NewSectionReader(p, p.size-metaSize-8, metaSize)
	protocol := thrift.NewTCompactProtocolFactory().GetProtocol(thrift.NewStreamTransportR(rs))
	metadata := meta.NewFileMetaData()
	if err := metadata.Read(protocol); err != nil {
		return nil, fmt.Errorf("metaData: failed to deserialize: %s", err)
	}
	return metadata, nil
}
