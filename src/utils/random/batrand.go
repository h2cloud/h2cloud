package random

import (
    "sync"
    "math/rand"
    "time"
)

var R=rand.New(rand.NewSource(time.Now().UnixNano()))

// Batch Random is capable of generating m [0, n) random numbers without any identical
// pairs. It can be accomplished with time complexity O(m) and space complexity O(n)
type BatchRandom struct {
    innerStorage []int
    lock sync.Mutex
    n0 int
}

func NewBatchRandom(n int) *BatchRandom {
    var t=make([]int, n)
    for i:=0; i<n; i++ {
        t[i]=i
    }
    return &BatchRandom {
        innerStorage: t,
        n0: n,
    }
}

// if n>=n0, just append.
// if n<n0, a innerMap will be regenerated
func (this *BatchRandom)Resize(n int) {
    this.lock.Lock()
    defer this.lock.Unlock()

    var n0=this.n0
    if n==n0 {
        return
    }
    if n<n0 {
        var t=make([]int, n)
        for i:=0; i<n; i++ {
            t[i]=i
        }
        this.innerStorage=t
        this.n0=n
        return
    }
    for i:=n0; i<n; i++ {
        this.innerStorage=append(this.innerStorage, i)
    }
    this.n0=n
}

func swap(x1 *int, x2 *int) {
    var x3=*x1
    *x1=*x2
    *x2=x3
}

// if m<=0, panic;
// if m>n, return a value with m-n zeros
func (this *BatchRandom)Get(m int) []int {
    if m==1 {
        return []int{R.Intn(this.n0)}
    }
    var ret=make([]int, m)
    if m>this.n0 {
        m=this.n0
    }

    this.lock.Lock()
    defer this.lock.Unlock()

    var st=R.Intn(this.n0)
    var des int
    for i:=0; i<m; i++ {
        des=st+R.Intn(this.n0-i)
        if des>=this.n0 {
            des-=this.n0
        }
        swap(&this.innerStorage[st], &this.innerStorage[des])
        ret[i]=this.innerStorage[st]
        st++
        if st>=this.n0 {
            st-=this.n0
        }
    }

    return ret
}
