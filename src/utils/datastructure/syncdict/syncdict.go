// synchronized version of dictionary.

package syncdict

import (
    "sync"
    . "definition"
)

type Syncdict struct {
    innerMap map[string]Tout
    lock *sync.RWMutex
}

func (this *Syncdict)Set(key string, value Tout) {
    this.lock.Lock()
    this.innerMap[key]=value
    this.lock.Unlock()
}

func (this *Syncdict)Declare(key string, value Tout) Tout {
    this.lock.Lock()
    var ret,ok=this.innerMap[key]
    if ok {
        this.lock.Unlock()
        return ret
    }
    this.innerMap[key]=value
    this.lock.Unlock()
    return value
}

// Returns NIL if not exists
func (this *Syncdict)Get(key string) Tout {
    this.lock.RLock()
    var ret=this.innerMap[key]
    this.lock.RUnlock()
    return ret
}

func NewSyncdict() *Syncdict {
    return &Syncdict {
        innerMap: make(map[string]Tout),
        lock: &sync.RWMutex{},
    }
}
