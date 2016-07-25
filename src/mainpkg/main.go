package main

import (
    "inapi"
    . "mainpkg/public"
    "intranet"
    "intranet/gossipd"
    dvc "kernel/distributedvc"
    "intranet/ping"
    . "logger"
)

func _no_use_() {
    inapi.Entry(nil)
}

func main() {
    StartUp()

    dvc.MergeManager.Launch()

    var exitCh=make(chan bool)

    go intranet.Entry(exitCh)
    go gossipd.Entry(exitCh)
    go inapi.Entry(exitCh)
    go WaitForSig(exitCh)
    go ping.Entry(exitCh)

    _=<-exitCh
    Secretary.Log("mainpkg::main", "Midware-MH2 is about to terminate...")

}
