package iomidware

import (
    "io"
)

type BlockReader struct {
    dataSource io.Reader
}

func (this *BlockReader)Read(p []byte) (int, error) {
    i:=0
    j:=len(p)
    for i<j {
        n, err:=this.dataSource.Read(p[i:])
        i=i+n
        if err!=nil {
            return i, err
        }
    }
    return i, nil
}

// Manipulate a reader and return its blocked version. When read from it, it blocks
// utils error occurs or all the buffer is used.
func Blockify(inp io.Reader) io.Reader {
    return &BlockReader{
        dataSource: inp,
    }
}
