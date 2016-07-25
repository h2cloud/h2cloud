package main

import (
    "kernel/filetype"
    "outapi"
    "flag"
    "fmt"
    "strings"
    "io/ioutil"
    "os"
)

const SWIFT_LOCALE="swift://"
func main() {
    flag.Parse()
    var args=flag.Args()
    if len(args)<1 {
        fmt.Fprintln(os.Stderr, "The input file should be specified.")
        os.Exit(1)
    }
    var path=args[0]
    if strings.HasPrefix(path, SWIFT_LOCALE) {
        var whole=strings.SplitN(path[len(SWIFT_LOCALE):], "/", 2)
        if len(whole)<2 {
            fmt.Fprintln(os.Stderr, "The container in Swift should be specified.")
            os.Exit(1)
        }
        var io=outapi.NewSwiftio(outapi.DefaultConnector, whole[0])
        var meta, file, err=io.Get(whole[1])
        if err!=nil {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }
        if meta==nil {
            fmt.Fprintln(os.Stderr, "File non exists.")
            os.Exit(1)
        }

        fmt.Fprintln(os.Stdout, "==FILE META==")
        for k, v:=range meta {
            fmt.Fprintln(os.Stdout, k, ":\t\t", v)
        }
        switch file:=file.(type) {
        case *filetype.Kvmap:
            fmt.Fprintln(os.Stdout, "==KVMAP FILE==")

            file.CheckOut()
            for k, v:=range file.Kvm {
                fmt.Fprintln(os.Stdout, k, ":\t\t", *v)
            }
        case *filetype.Nnode:
            fmt.Fprintln(os.Stdout, "==NNODE FILE==")
            fmt.Fprintln(os.Stdout, "Pointed:\t\t", file.DesName)
        default:
            fmt.Fprintln(os.Stdout, "==TYPE NOT FOUND==")
        }
        os.Exit(0)
    } else {
        var res, err=ioutil.ReadFile(path)
        if err!=nil {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }
        fmt.Fprintln(os.Stdout, string(res))
        os.Exit(0)
    }
}
