package distributedvc

import (
    . "logger"
    "testing"
    "strconv"
    "kernel/filetype"
    "time"
    "sync"
)

func _TestParallelCommit(t *testing.T) {
    Insider.LogD("+++++ TestParallelCommit::start")

    var wg sync.WaitGroup
    var routine=func(number int) {
        defer wg.Done()

        var filename=strconv.Itoa(number)
        {
            var desParentMap=GetFD(filename, io)
            if desParentMap==nil {
                Insider.LogD("Fail to get foldermap fd for folder "+filename)
                return
            }
            if err:=desParentMap.Submit(filetype.FastMake("hasi")); err!=nil {
                Insider.LogD("Fail to submit foldermap patch for folder "+filename)
                desParentMap.Release()
                return
            }
            desParentMap.Release()
        }
        Insider.LogD("Succeed process for "+filename)
    }

    var numberOfThreads=100
    wg.Add(numberOfThreads)
    for i:=0; i<numberOfThreads; i++ {
        go routine(i)
    }
    wg.Wait()

    Insider.LogD("----- TestParallelCommit::end")

    for {
        time.Sleep(time.Hour)
    }
}
