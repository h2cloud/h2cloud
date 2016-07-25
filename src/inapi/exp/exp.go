package exp


import (
    . "github.com/levythu/gurgling"
    "kernel/filesystem"
    "fmt"
    "outapi"
    sysio "io"
    "strconv"
    //"logger"
)


func __expgo_nouse() {
    fmt.Println("nouse")
}

func ExpRouter() Router {
    var rootRouter=ARegexpRouter()

    rootRouter.Put(`/batchput`, batchPutHandler)
    rootRouter.Get(`/rawget/([^/]+)/(.*)`, rawGetHandler)

    return rootRouter
}

func rawGetHandler(req Request, res Response) {
    var rr=req.F()["RR"].([]string)
    var io=outapi.NewSwiftio(outapi.DefaultConnector, rr[1])
    var _, rc, err=io.GetStreamX(rr[2])
    if err!=nil {
        res.SendCode(500)
    } else if rc==nil {
        res.SendCode(404)
    } else {
        var w=res.R()
        if _, copyErr:=sysio.Copy(w, rc); copyErr!=nil {
            rc.Close()
            return
        }
        if err2:=rc.Close(); err2!=nil {
            return
        }
    }
}

func batchPutHandler(req Request, res Response) {
    var container=req.Get("P-Container")
    var frominode=req.Get("P-From-Inode")
    var fromn=req.Get("P-From")
    var ton=req.Get("P-To")
    var prefix=req.Get("P-Prefix")

    var content="The quick brown fox jumps over the lazy dog"

    var fs=filesystem.GetFs(outapi.NewSwiftio(outapi.DefaultConnector, container))
    i, _:=strconv.Atoi(fromn)
    j, _:=strconv.Atoi(ton)
    if err:=fs.BatchPutDir(prefix, frominode, i, j, content); err!=nil {
        res.Send(err.Error())
    } else {
        res.Send("OK")
    }
}
