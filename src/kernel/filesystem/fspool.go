package filesystem

/*
** A pool for single-instance implementation of Fs and cache it as needed.
** GetFs() to get one Fs and Release it after use.
**
** Currently there's only one container. Thus, no auto-clearing has been
** implemented.
*/

import (
    "outapi"
    "sync"
)

var globalFsMap=map[string]*Fs{}
var fsMapLock=sync.RWMutex{}

// use outapi::Outapi.GenerateUniqueID() as identifier
// a nil may be returned indicating error
func GetFs(io outapi.Outapi) *Fs {
    var id=io.GenerateUniqueID()

    fsMapLock.RLock()
    if elem, ok:=globalFsMap[id]; ok {
        fsMapLock.RUnlock()
        return elem
    }
    fsMapLock.RUnlock()
    fsMapLock.Lock()
    defer fsMapLock.Unlock()

    if elem, ok:=globalFsMap[id]; ok {
        return elem
    }
    var ret=newFs(io)
    globalFsMap[id]=ret

    return ret
}

func (this *Fs)Release() {
    // nothing to do, cuz auto clearing is not implemented.
    return
}
