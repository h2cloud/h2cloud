package distributedvc

// Unit test for kernel/distributedvc

import (
    "testing"
    "fmt"
    . "outapi"
    . "definition/configinfo"
    . "kernel/filetype"
    . "utils/timestamp"
    "time"
)

var swiftc=ConnectbyAuth(KEYSTONE_USERNAME, KEYSTONE_PASSWORD, KEYSTONE_TENANT)
var io=NewSwiftio(swiftc, "testcon")

func _TestFDGet(t *testing.T) {
    fmt.Println("+++++ TestFDGet::start")
    var huahua=GetFD("huahuad", io)
    huahua.GraspReader()
    fmt.Println(huahua.Read())
    huahua.ReleaseReader()
    huahua.Release()
    fmt.Println("----- TestFDGet::end")

    for {
        time.Sleep(time.Hour)
    }
}

func _TestFDSubmit(t *testing.T) {
    fmt.Println("+++++ TestFDSubmit::start")
    var huahua=GetFD("huahuad", io)
    var toSubmit=NewKvMap()
    toSubmit.CheckOut()
    fmt.Println("+ Checked out")
    toSubmit.Kvm["huahuax"]=&KvmapEntry {
        Key: "huahuax",
        Val: "baomihua",
        Timestamp: GetTimestamp(),
    }
    toSubmit.CheckIn()
    fmt.Println("+ Checked in")
    fmt.Println(huahua.Submit(toSubmit))

    huahua.Release()
    fmt.Println("----- TestFDSubmit::start")


    for {
        time.Sleep(time.Hour)
    }
}
