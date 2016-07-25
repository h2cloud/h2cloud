package inapi

import (
    . "github.com/levythu/gurgling"
    //"github.com/levythu/gurgling/midwares/analyzer"
    "inapi/containermanage"
    "inapi/fsmanage"
    "inapi/streamio"
    "inapi/exp"
    conf "definition/configinfo"
    . "logger"
)

func Entry(exit chan bool) {
    defer (func(){
        exit<-false
    })()

    var rootRouter=ARouter()
    //rootRouter.Use(analyzer.ASimpleAnalyzer())

    rootRouter.Use("/fs", fsmanage.FMRouter())
    rootRouter.Use("/io", streamio.IORouter())
    rootRouter.Use("/cn", containermanage.CMRouter())
    rootRouter.Use("/exp", exp.ExpRouter())

    Secretary.Log("inapi::Entry()", "Now launching public service at "+conf.OUTER_SERVICE_LISTENER)
    var err=rootRouter.Launch(conf.OUTER_SERVICE_LISTENER)
    if err!=nil {
        Secretary.Error("inapi::Entry()", "HTTP Server terminated: "+err.Error())
    }
}
