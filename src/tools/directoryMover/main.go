package main

import (
    "github.com/ncw/swift"
    "outapi"
    "flag"
    "fmt"
    "time"
    "sync"
    "os"
)

var c=outapi.DefaultConnector.DumpConn()

func main() {
    var pContainer=flag.String("container", "", "The container to manipulate.")
    var pFromPath=flag.String("from", "", "The from path.")
    var pToPath=flag.String("to", "", "The to path")
    var pThread=flag.Int("thread", 1, "The thread to issue concurrently")
    var pDelete=flag.Bool("delete", true, "Perform delete")
    var pCopy=flag.Bool("copy", true, "Perform copy")
    flag.Parse()

    if *pContainer=="" {
        fmt.Println("Container must be specified.")
        os.Exit(1)
    }
    if *pFromPath==*pToPath {
        fmt.Println("FromPath==ToPath, abort.")
        return
    }
    var nowTime=time.Now().UnixNano()

    var objList, err=c.ObjectsAll(*pContainer, &swift.ObjectsOpts {
        Prefix: *pFromPath,
    })
    if err!=nil {
        fmt.Println(err)
        os.Exit(1)
    }

    var io=outapi.NewSwiftio(outapi.DefaultConnector, *pContainer)
    var wg sync.WaitGroup
    var rollingList=func(begg int, endd int) {
        if endd<0 || endd>len(objList) {
            endd=len(objList)
        }
        for i:=begg; i<endd; i++ {
            var e=objList[i]
            var fromName=e.Name
            var toName=*pToPath+e.Name[len(*pFromPath):]
            fmt.Println("Moving", fromName, "->", toName)

            if *pCopy {
                if err:=io.Copy(fromName, toName, nil); err!=nil {
                    fmt.Println("Error:", err, "when trying to copy", fromName)
                    os.Exit(1)
                }
            }
            if *pDelete {
                if err:=io.Delete(fromName); err!=nil {
                    fmt.Println("Error:", err, "when trying to delete", fromName)
                    os.Exit(1)
                }
            }
        }
        wg.Done()
    }

    var slice=len(objList)/(*pThread)
    if slice<1 {
        slice=1
    }
    var now=0
    for now<len(objList) {
        go rollingList(now, now+slice)
        now=now+slice
        wg.Add(1)
    }
    wg.Wait()

    fmt.Println("Time consumed:", time.Now().UnixNano()-nowTime, "ns")
    return
}
