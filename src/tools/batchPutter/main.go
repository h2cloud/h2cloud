package main

import (
    "outapi"
    "flag"
    "fmt"
    "strconv"
    . "kernel/distributedvc/filemeta"
    . "kernel/distributedvc/constdef"
    "os"
    sysio "io"
    "strings"
)

const STREAM_TYPE="stream type file"
func main() {
    var pcontainer=flag.String("container", "", "The container to put objects in.")
    var ppath=flag.String("path", "", "The path to put objects in.")
    var pfromnumber=flag.Int("from", 0, "From number")
    var ptonumber=flag.Int("to", 1, "To number(excluded)")
    var pprefix=flag.String("prefix", "BatchPut", "Prefix name")
    var psuffix=flag.String("suffix", "", "Suffix name")
    flag.Parse()

    var container=*pcontainer
    var path=*ppath
    var fromnumber=*pfromnumber
    var tonumber=*ptonumber
    var prefix=*pprefix
    var content="The quick brown fox jumps over the lazy dog"

    if container=="" {
        fmt.Println("Container should be specified!")
        os.Exit(1);
        return;
    }

    var io=outapi.NewSwiftio(outapi.DefaultConnector, container)
    for i:=fromnumber; i<tonumber; i++ {
        var filename=path+prefix+strconv.Itoa(i)+*psuffix
        fmt.Println("Putting file "+filename);

        var meta=NewMeta()
        meta=meta.Clone()
        meta[METAKEY_TYPE]=STREAM_TYPE
        var wc, err=io.PutStream(filename, meta)
        if err!=nil {
            fmt.Println("Error when putting "+filename, err)
            os.Exit(1)
        }
        sysio.Copy(wc, strings.NewReader(content))
        wc.Close()
    }
}
