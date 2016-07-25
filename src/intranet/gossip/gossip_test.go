package gossip

import (
    "testing"
    "time"
    "fmt"
    . "definition"
    conf "definition/configinfo"
)

func TestBatRand(t *testing.T) {
    GlobalGossiper=(func() Gossiper {
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
    })()
    go GlobalGossiper.Launch()
    var i=0
    for {
        time.Sleep(time.Second)
        if err:=GlobalGossiper.PostGossip(i); err==nil {
            fmt.Println("POSTED:", i)
        } else {
            fmt.Println("ERROR:", err)
        }
        i++
    }
}
