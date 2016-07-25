package public

import (
    . "logger"
    "os"
    "os/signal"
    "sync"
)

var gLock=sync.Mutex{}
func OnSignaled(sig os.Signal) bool {
    gLock.Lock()
    // prevent others entering
    Secretary.Log("mainpkg::Terminated", "Receive signal "+sig.String())

    return false
}

func WaitForSig(exit chan bool) {
    defer (func(){
        exit<-false
    })()

    c:=make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, os.Kill)

    for sig:=range c {
        if !OnSignaled(sig) {
            return
        }
    }
}
