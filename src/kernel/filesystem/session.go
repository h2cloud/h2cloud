package filesystem
// A wrapper for fsfunc recording the working directory.

import (
    "outapi"
    "sync"
    "fmt"
)

type Session struct {
    fs *Fs
    d string
    locks []*sync.Mutex
}
func NewSession(io outapi.Outapi) *Session {
    var ret=&Session{
        fs: GetFs(io),
        d: ROOT_INODE_NAME,
        locks: []*sync.Mutex{&sync.Mutex{},&sync.Mutex{}},
    }
    if ret.fs==nil {
        return nil
    }
    return ret
}

func ____nouse__() {
    fmt.Println("123")
}

func (this *Session)Cd(path string) error {
    this.locks[0].Lock()
    defer this.locks[0].Unlock()

    var tempD, err=this.fs.Locate(path, this.d)
    if err==nil {
        this.d=tempD
    }
    return err
}

func (this *Session)Mkdir(foldername string) error {
    return this.fs.Mkdir(foldername, this.d, false)
}

func (this *Session)Rm(foldername string) error {
    return this.fs.Rm(foldername, this.d)
}

func (this *Session)Ls() ([]string, error) {
    return this.fs.List(this.d)
}

func (this *Session)PwdInode() string {
    return this.d
}

func (this *Session)Release() {
    this.fs.Release()
}
