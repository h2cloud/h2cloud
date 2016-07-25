package gossipd

import (
    . "github.com/levythu/gurgling"
    "io/ioutil"
    dvc "kernel/distributedvc"
    . "logger"
    gsp "intranet/gossip"
    . "intranet/gossipd/interactive"
    "intranet/ping"
)

// @ async
func checkGossipedData(src []*GossipEntry) {
    // TODO: whether use multi-routine?
    for _, e:=range src {
        if e.OutAPI==OUTAPI_PLACEHOLDER_PING_FLAG {
            if ping.Pong(e) {
                if err:=gsp.GlobalGossiper.PostGossip(e); err!=nil {
                    Secretary.Warn("gossipd::checkGossipedData", "Fail to post heartbeat gossiping to other nodes: "+err.Error())
                }
            }
            continue
        }
        Secretary.Log("gossipd::checkGossipedData", "Gossip received: "+e.Filename+" @ "+e.OutAPI)
        dvc.MergeManager.SubmitGossipingTask(e)
    }
}
/*
** GOSSIP API: Posted
** Method:      POST
** URL:         [:intranet]/gossip
** Parameter:   Content(in Body): the raw body is the parameter content itself.
*/
func OnPostedGossip(req Request, res Response) {
    if ct, err:=ioutil.ReadAll(req.R().Body); err!=nil {
        Secretary.Error("gossipd::OnPostedGossip", "Fail to read data from gossiped request: "+err.Error())
        res.SendCode(500)
        return
    } else {
        var pList=ParseAll(string(ct))
        if pList==nil {
            Secretary.Error("gossipd::OnPostedGossip", "Format error for gossiped data")
            res.SendCode(403)
            return
        }

        go checkGossipedData(pList)
        res.SendCode(200)
    }
}

func GetGossipRouter() Router {
    var r=ARouter()
    r.Post("/", OnPostedGossip)

    return r
}
