package random

import (
    "fmt"
    "testing"
)

func TestBatRand(t *testing.T) {
    var p=NewBatchRandom(20)
    fmt.Println(p.Get(1))
    fmt.Println(p.Get(10))
    fmt.Println(p.Get(30))
    fmt.Println(p.Get(2))
    fmt.Println(p.Get(5))
    fmt.Println(p.Get(5))
    p.Resize(100)
    fmt.Println(p.Get(10))
    fmt.Println(p.Get(2))
    fmt.Println(p.Get(5))
    fmt.Println(p.Get(5))

    fmt.Println(p.innerStorage)
    var m=make(map[int]bool)
    for i:=0; i<p.n0; i++ {
        if m[p.innerStorage[i]] {
            fmt.Println("FAIL.")
            t.Fail()
        }
        m[p.innerStorage[i]]=true
    }
}
