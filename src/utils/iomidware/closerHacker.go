package iomidware

import (
    "io"
)

type Callback func(error) error

//==========================================================
type CloserHackerPost struct {
    cb Callback
    original io.Closer
}
func NewCloserHackerPost(ori io.Closer, _cb Callback) *CloserHackerPost {
    return &CloserHackerPost{
        cb: _cb,
        original: ori,
    }
}

func (this *CloserHackerPost)Close() error {
    err1:=this.original.Close()
    return this.cb(err1)
}
//==============================================================
type WritecloserHackerPost struct {
    cb Callback
    original io.WriteCloser
}
func NewWritecloserHackerPost(ori io.WriteCloser, _cb Callback) *WritecloserHackerPost {
    return &WritecloserHackerPost{
        cb: _cb,
        original: ori,
    }
}

func (this *WritecloserHackerPost)Close() error {
    err1:=this.original.Close()
    return this.cb(err1)
}
func (this *WritecloserHackerPost)Write(p []byte) (n int, err error) {
    return this.original.Write(p)
}
//==============================================================
