package intranet

// used for administrative maintainance and inter-communication between servers as well

import (
    . "github.com/levythu/gurgling"
    conf "definition/configinfo"
    . "logger"
    "intranet/gossipd"
)

func Entry(exit chan bool) {
    defer (func(){
        exit<-false
    })()

    var rootRouter=ARouter()

    rootRouter.Get("/", func(res Response) {
        res.Redirect("/admin")
    })
    if r:=gossipd.GetGossipRouter(); r!=nil {
        rootRouter.Use("/gossip", r)
    }
    if r:=getAdminPageRouter(); r!=nil {
        rootRouter.Use("/admin", r)
    }

    Secretary.Log("intranet::Entry()", "Now launching intranet service at "+conf.INNER_SERVICE_LISTENER)
    var err=rootRouter.Launch(conf.INNER_SERVICE_LISTENER)
    if err!=nil {
        Secretary.Error("intranet::Entry()", "HTTP Server terminated: "+err.Error())
    }
}
