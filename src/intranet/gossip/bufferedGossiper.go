package gossip

import (
    . "definition"
    "sync"
    "errors"
    . "logger"
    "fmt"
    "time"
)

type BufferedGossiper struct {
    *stdGossiperListImplementation

    // <0 for no launch
    PeriodInMillisecond int

    BufferSize int

    // EnsureTellCount is the times of propagation for each posted gossip
    EnsureTellCount int

    // TellMaxCount is the max number of gossips that can be delivered in one tick
    TellMaxCount int

    // ParallelTell is the number of nodes told in one tick
    ParallelTell int

    // ===============================

    do func(addr Tout, content []Tout) error

    // the next of the last
    tail int
    head int

    lenLock sync.RWMutex
    len int

    buffer []Tout
    gCount []int
    notifyFail []bool
}
func (this *BufferedGossiper)SetGossipingFunc(do func(addr Tout, content []Tout) error) {
    this.do=do
}

func NewBufferedGossiper(bufferSize int) *BufferedGossiper {
    return &BufferedGossiper {
        BufferSize: bufferSize,
        head: 0,
        tail: 0,
        len: 0,
        buffer: make([]Tout, bufferSize),
        gCount: make([]int, bufferSize),
        notifyFail: make([]bool, bufferSize),
        stdGossiperListImplementation: &stdGossiperListImplementation{},
    }
}

var BUFFER_IS_FULL=errors.New("The buffer for buffered gossiper is full. New gossip cannot be checked in.")
func (this *BufferedGossiper)PostGossip(content Tout) error {
    this.lenLock.Lock()
    defer this.lenLock.Unlock()

    if this.len==this.BufferSize {
        return BUFFER_IS_FULL
    }
    this.len++

    this.buffer[this.tail]=content
    this.gCount[this.tail]=this.EnsureTellCount
    this.notifyFail[this.tail]=true
    this.tail++
    if this.tail>=this.BufferSize {
        this.tail-=this.BufferSize
    }

    return nil
}
func (this *BufferedGossiper)PostGossipSilent(content Tout) error {
    this.lenLock.Lock()
    defer this.lenLock.Unlock()

    if this.len==this.BufferSize {
        return BUFFER_IS_FULL
    }
    this.len++

    this.buffer[this.tail]=content
    this.gCount[this.tail]=this.EnsureTellCount
    this.notifyFail[this.tail]=false
    this.tail++
    if this.tail>=this.BufferSize {
        this.tail-=this.BufferSize
    }

    return nil
}

func (this *BufferedGossiper)gossip(content []Tout, notify bool) {
    if len(content)==0 {
        return
    }
    var taskList, err=this.stdGossiperListImplementation.RandList(this.ParallelTell)
    if err!=nil && notify {
        Secretary.Error("gossip::BufferedGossiper.gossip()", "Unable to gossip due to "+err.Error())
        return
    }
    for _, e:=range taskList {
        go (func(x Tout) {
            if err:=this.do(x, content); err!=nil && notify {
                Secretary.Warn("gossip::BufferedGossiper.gossip()", "Failed to gossip to "+fmt.Sprint(x)+": "+err.Error())
                // TODO: retry others?
            }
        })(e)
    }
}
func (this *BufferedGossiper)onTick() {
    this.lenLock.Lock()

    var c=this.len
    if c>this.TellMaxCount {
        c=this.TellMaxCount
    }

    var res=make([]Tout, c)

    var p=this.head
    var notify=false
    for i:=0; i<c; i++ {
        res[i]=this.buffer[p]
        this.gCount[p]-=this.ParallelTell
        notify=notify || this.notifyFail[p]
        p++
        if p>=this.BufferSize {
            p-=this.BufferSize
        }
    }
    for this.len>0 && this.gCount[this.head]<=0 {
        this.len--
        this.head++
        if this.head>=this.BufferSize {
            this.head-=this.BufferSize
        }
    }
    this.lenLock.Unlock()

    this.gossip(res, notify)
}

func (this *BufferedGossiper)Launch() error {
    if this.PeriodInMillisecond<0 {
        return nil
    }
    Secretary.Log("gossip::BufferedGossiper.Launch()", "Gossiper is launched.")
    var dur=time.Duration(this.PeriodInMillisecond)*time.Millisecond
    for {
        this.onTick()
        time.Sleep(dur)
    }
}

// ONLY for monitoring. IT WILL CAST GREAT NEGATIVE IMPACT on working efficiency
func (this *BufferedGossiper)DumpGossip() ([]Tout, []int) {
    this.lenLock.Lock()
    defer this.lenLock.Unlock()

    var x1=make([]Tout, this.len)
    var x2=make([]int, this.len)

    var c=this.len
    var p=this.head

    for i:=0; i<c; i++ {
        x1[i]=this.buffer[p]
        x2[i]=this.gCount[p]
        p++
        if p>=this.BufferSize {
            p-=this.BufferSize
        }
    }

    return x1, x2
}

func (this *BufferedGossiper)GenerateProfile() map[string]Tout {
    var bf, gc=this.DumpGossip()
    return map[string]Tout {
        "content_buffer": bf,
        "count_rest": gc,
    }
}
