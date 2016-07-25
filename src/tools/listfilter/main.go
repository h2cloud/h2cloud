package main

import (
    "github.com/ncw/swift"
    "outapi"
    "flag"
    "fmt"
    "time"
    "os"
    "strings"
)

var c=outapi.DefaultConnector.DumpConn()

func main() {
    var pContainer=flag.String("container", "", "The container to manipulate.")
    var pPrefix=flag.String("prefix", "", "Prefix")
    flag.Parse()

    if *pContainer=="" {
        fmt.Println("Container must be specified.")
        os.Exit(1)
    }

    var nowTime=time.Now().UnixNano()

    var opt=&swift.ObjectsOpts {}
    var objList, err=c.ObjectsAll(*pContainer, opt)
    var resMap=map[string]bool{}

    if err!=nil {
        fmt.Println("Error:", err)
        os.Exit(1)
    }
    for _, e:=range objList {
        if strings.HasPrefix(e.Name, *pPrefix) {
            var trimmed=e.Name[len(*pPrefix):]
            if i:=strings.Index(trimmed, "/"); i>=0 {
                trimmed=trimmed[:i]
            }
            resMap[trimmed]=true
        }
    }

    nowTime=time.Now().UnixNano()-nowTime

    for k, _:=range resMap {
        fmt.Println(k)
    }
    fmt.Println("Total:", len(resMap))
    fmt.Println("Time consumed:", nowTime, "ns")
    return
}
