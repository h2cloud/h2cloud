package ping

import (
    . "utils/timestamp"
    "sync"
    . "intranet/gossipd/interactive"
    "strconv"
)

var activeList=map[int]ClxTimestamp{}
var aLLock sync.RWMutex

// nonexist returns zero
func QueryConn(nodenum int) ClxTimestamp {
    aLLock.RLock()
    defer aLLock.RUnlock()

    return activeList[nodenum]
}

func DumpConn() map[string]uint64 {
    aLLock.RLock()
    defer aLLock.RUnlock()

    var ret=map[string]uint64{}
    for k, v:=range activeList {
        ret[strconv.Itoa(k)]=v.Val()
    }

    return ret
}

// returning value indicates whether the gossip should be passed on
func Pong(context *GossipEntry) bool {
    aLLock.Lock()
    defer aLLock.Unlock()

    if activeList[context.NodeNumber]<context.UpdateTime {
        activeList[context.NodeNumber]=context.UpdateTime
        return true
    } else {
        // it has been notified before. there is no need to propagate the gossip
        return false
    }
}
