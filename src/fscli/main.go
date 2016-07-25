package main

import (
    . "mainpkg/public"
    . "github.com/levythu/gurgling"
    . "outapi"
    conf "definition/configinfo"
    fs "kernel/filesystem"
    "strings"
    . "logger"
    "fmt"
)

func main() {
    StartUp()

    var Testio=NewSwiftio(ConnectbyAuth(conf.KEYSTONE_USERNAME, conf.KEYSTONE_PASSWORD, conf.KEYSTONE_TENANT), "testcon")
    var session=fs.NewSession(Testio)
    var router=ARouter().Use(func(req Request, res Response) {
        var cmd=req.Path()[1:]
        fmt.Println(cmd)
        if len(cmd)>=2 && cmd=="ls" {
            var ret, err=session.Ls()
            if err!=nil {
                res.Status(err.Error(), 500)
            } else {
                res.Send(strings.Join(ret, "\n"))
            }
        } else if len(cmd)>=6 && cmd[:6]=="mkdir " {
            var name=cmd[6:]
            if err:=session.Mkdir(name); err!=nil {
                res.Status(err.Error(), 500)
            } else {
                res.Send("Created.")
            }
        } else if len(cmd)>=3 && cmd[:3]=="rm " {
            var name=cmd[3:]
            if err:=session.Rm(name); err!=nil {
                res.Status(err.Error(), 500)
            } else {
                res.Send("Removed.")
            }
        } else if len(cmd)>=3 && cmd[:3]=="cd " {
            var name=cmd[3:]
            if err:=session.Cd(name); err!=nil {
                res.Status(err.Error(), 500)
            } else {
                res.Send("OK.")
            }
        } else if len(cmd)==4 && cmd=="pwdx" {
            res.Send(session.PwdInode())
        } else {
            res.Status("404 NOT FOUND", 404)
        }
    })

    Secretary.LogD("Running on port 8192.")
    router.Launch(":8192")
}
