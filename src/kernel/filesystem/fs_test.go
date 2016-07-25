package filesystem

import (
    "testing"
    "time"
    "fmt"
    "strings"
    . "github.com/levythu/gurgling"
)

// Attentez: in the test the fs never get released.
var fs4test=GetFs(Testio)

func _TestFormat(t *testing.T) {
    fmt.Println(fs4test.FormatFS())

    for {
        time.Sleep(time.Hour)
    }
}

func _TestMkDir(t *testing.T) {
    fmt.Println(fs4test.Mkdir("directory1", fs4test.rootName, false))

    for {
        time.Sleep(time.Hour)
    }
}

func _TestLS(t *testing.T) {
    fmt.Println(fs4test.List(fs4test.rootName))

    for {
        time.Sleep(time.Hour)
    }
}

func TestSession(t *testing.T) {
    var session=NewSession(Testio)
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
        } else {
            res.Status("404 NOT FOUND", 404)
        }
    })

    fmt.Println("Running on port 8192.")
    router.Launch(":8192")
}
