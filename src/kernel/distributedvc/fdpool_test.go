package distributedvc

// Unit test for kernel/distributedvc

import (
    "testing"
    "fmt"
    "strconv"
    "sync"
    "math/rand"
    . "outapi"
)

var testOutIO=&EmptyIO{}

func _TestAutoDormant(t *testing.T) {
    var wg sync.WaitGroup
    for i:=0; i<10050; i++ {
        wg.Add(1)
        go (func(num int) {
            //fmt.Println("Thread #", num, "is running.")
            var name="name "+strconv.Itoa(num)
            var des=GetFD(name, testOutIO)
            des.GraspReader()
            des.Read()
            des.ReleaseReader()
            des.Release()
            wg.Done()
        })(i)
    }
    wg.Wait()
    fmt.Println(dormant.Length)
}

func _TestAllRound(t *testing.T) {
    var wg sync.WaitGroup
    for i:=0; i<10050; i++ {
        wg.Add(1)
        go (func(num int) {
            //fmt.Println("Thread #", num, "is running.")
            var name="name "+strconv.Itoa(num)
            var des=GetFD(name, testOutIO)
            des.GraspReader()
            des.Read()
            des.ReleaseReader()
            des.Release()
            wg.Done()
        })(rand.Intn(499))
    }
    wg.Wait()
    fmt.Println(dormant.Length)
}
