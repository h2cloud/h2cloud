package main

import (
    "github.com/ncw/swift"
    "outapi"
    "flag"
    "fmt"
    "time"
    "os"
)

var c=outapi.DefaultConnector.DumpConn()

func main() {
    var pContainer=flag.String("container", "", "The container to manipulate.")
    var pPrefix=flag.String("prefix", "", "Prefix")
    var pDelimiter=flag.String("delimiter", "", "Delimiter")
    var pNoRes=flag.Bool("noresult", false, "NoResult")
    flag.Parse()

    if *pContainer=="" {
        fmt.Println("Container must be specified.")
        os.Exit(1)
    }

    var nowTime=time.Now().UnixNano()

    var opt=&swift.ObjectsOpts {}
    if *pPrefix!="" {
        opt.Prefix=*pPrefix
    }
    if *pDelimiter!="" {
        opt.Delimiter=rune((*pDelimiter)[0])
    }
    var objList, err=c.ObjectsAll(*pContainer, opt)

    nowTime=time.Now().UnixNano()-nowTime

    if err==nil && !(*pNoRes) {
        for _, e:=range objList {
            fmt.Println(e.Name)
        }
    }
    fmt.Println("Total:", len(objList))
    fmt.Println("Time consumed:", nowTime, "ns")
    return
}
