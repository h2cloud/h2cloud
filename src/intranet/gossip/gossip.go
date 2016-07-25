package gossip

import (
    . "definition"
    "sync"
    "utils/random"
    "errors"
    _ "fmt"
    //conf "definition/configinfo"
)

type Gossiper interface {
    PostGossip(content Tout) error
    PostGossipSilent(content Tout) error

    // the list passed in will be replicated.
    SetGossiperList(list []Tout) error

    // the do func will get invoked asynchonously
    SetGossipingFunc(do func(addr Tout, content []Tout) error)

    // a deamon function
    Launch() error

    GenerateProfile() map[string]Tout
}

var GlobalGossiper Gossiper=nil
/*var GlobalGossiper Gossiper=(func() Gossiper {
    var ret=NewBufferedGossiper(conf.GOSSIP_BUFFER_SIZE)
    ret.PeriodInMillisecond=conf.GOSSIP_PERIOD_IN_MS
    ret.EnsureTellCount=conf.GOSSIP_RETELL_TIMES
    ret.TellMaxCount=conf.GOSSIP_MAX_DELIVERED_IN_ONE_TICK
    ret.ParallelTell=conf.GOSSIP_MAX_TELLING_IN_ONE_TICK
    ret.SetGossiperList([]Tout{1,2,3,4,5})
    ret.SetGossipingFunc(func(addr Tout, content []Tout) error {
        fmt.Println(addr, ": ", content)
        return nil
    })

    return ret
})()*/

type stdGossiperListImplementation struct {
    list []Tout
    lock sync.RWMutex

    ranBatcher *random.BatchRandom
}

func (this *stdGossiperListImplementation)SetGossiperList(list []Tout) error {
    this.lock.Lock()
    defer this.lock.Unlock()

    this.list=[]Tout{}
    for _, e:=range list {
        this.list=append(this.list, e)
    }
    if this.ranBatcher==nil {
        this.ranBatcher=random.NewBatchRandom(len(list))
    } else {
        this.ranBatcher.Resize(len(list))
    }
    return nil
}

func (this *stdGossiperListImplementation)RandList(total int) ([]Tout, error) {
    this.lock.RLock()
    defer this.lock.RUnlock()

    if this.list==nil {
        return nil, errors.New("The gossiper list has not been set yet.")
    }

    if total>len(this.list) {
        total=len(this.list)
    }
    var ret=make([]Tout, total)
    var result=this.ranBatcher.Get(total)
    for i, e:=range result {
        ret[i]=this.list[e]
    }

    return ret, nil
}
func (this *stdGossiperListImplementation)Get(numberInList int) (Tout, error) {
    this.lock.RLock()
    defer this.lock.RUnlock()

    if this.list==nil {
        return nil, errors.New("The gossiper list has not been set yet.")
    }
    if numberInList<0 || numberInList>=len(this.list) {
        return nil, errors.New("Invalid Access to gossiper list.")
    }

    return this.list[numberInList], nil
}
func (this *stdGossiperListImplementation)GetLen() (int, error) {
    this.lock.RLock()
    defer this.lock.RUnlock()

    if this.list==nil {
        return 0, errors.New("The gossiper list has not been set yet.")
    }

    return len(this.list), nil
}
