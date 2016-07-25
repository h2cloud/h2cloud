package ping

import (
    gsp "intranet/gossip"
    . "intranet/gossipd/interactive"
    conf "definition/configinfo"
    . "logger"
    "time"
    . "utils/timestamp"
    "strconv"
)

func Entry(exit chan bool) {
    // MUST GUARANTEE that "intranet/gossipd" has been inited

    if conf.HEARTBEAT_PING_INTERVAL<=0 {
        // exit without nofifying
        return
    }
    defer (func(){
        exit<-false
    })()

    var sleepInterval=time.Millisecond*time.Duration(conf.HEARTBEAT_PING_INTERVAL)
    Secretary.Log("ping::Entry()", "Start to ping every "+strconv.Itoa(conf.HEARTBEAT_PING_INTERVAL)+" ms")
    for {
        var err=gsp.GlobalGossiper.PostGossipSilent(&GossipEntry {
            Filename: "",
            OutAPI: OUTAPI_PLACEHOLDER_PING_FLAG,
            UpdateTime: GetTimestamp(),
            NodeNumber: conf.NODE_NUMBER,
        })
        if err!=nil {
            Secretary.Warn("ping::Entry()", "Fail to post ping gossiping: "+err.Error())
        }
        time.Sleep(sleepInterval)
    }
}
