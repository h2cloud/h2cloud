// K-V map file, storing string-to-string map based on the diction sort of keys,
// kernel file type to store folder index.
// File structure:
//     - BYTE0~3: magic chars "KVMP"
//     =========Rep=========
//     - 4B-int: n, 4B-int:m
//     - 8B-int: timestamp
//     - n B: unicoded key
//     - m B: unicoded value
//     =====================
//     - 0
// Store structure per entry:
//     (key,(value,timestamp))
// How to edit:
//     1. Construct with a string(create)/tuple(modify) as 2nd parameter
//     2. checkOut
//     3. edit kvm
//     4. checkIn
//     5. writeBack
// How to merge:
//     1. Construct with a tuple as 2nd parameter
//     2. mergeWith
//     3. [IF needs modification, checkIn/Out]
//     4. writeBack

package filetype

import (
    . "utils/timestamp"
    "io"
    "definition/exception"
    "encoding/binary"
    "log"
    "sort"
    "sync"
    "fmt"
)

const fileMagic="KVMP"
const REMOVE_SPECIFIED="$@REMOVED@$)*!*"

type KvmapEntry struct {
    Timestamp ClxTimestamp
    Key string
    Val string
}

type Kvmap struct {
    finishRead bool

    Kvm map[string]*KvmapEntry
    rmed map[string]*KvmapEntry

    readData []*KvmapEntry
    dataSource io.Reader

    fileTS ClxTimestamp

    // Attentez: the lock does not protect Kvm & rmed
    lock *sync.RWMutex
}

func Kvmap_verbose() {
    // NOUSE, only for crunching the warning.
    fmt.Print("useless")
}
func NewKvMap() *Kvmap {
    var nkv Kvmap
    rkv:=&nkv

    rkv.Init(nil, GetTimestamp())
    rkv.finishRead=true
    return rkv
}

func (this *Kvmap)Init(dtSource io.Reader, dtTimestamp ClxTimestamp) {
    this.readData=make([]*KvmapEntry, 0)
    this.dataSource=dtSource
    this.fileTS=dtTimestamp
    this.finishRead=false
    this.lock=&sync.RWMutex{}
}

func (this *Kvmap)TSet(dtTimestamp ClxTimestamp) {
    this.lock.Lock()
    defer this.lock.Unlock()

    this.fileTS=dtTimestamp
}
func (this *Kvmap)TGet() ClxTimestamp {
    this.lock.RLock()
    defer this.lock.RUnlock()

    return this.fileTS
}

func (this *Kvmap)LoadIn(dtSource io.Reader) error {
    this.Init(dtSource, 0)
    return this.EnsureRead()
}
func (this *Kvmap)GetType() string {
    return "key-value map file"
}

func ParseString(inp io.Reader ,length uint32) (string, error) {
    buf:=make([]byte, length)
    n, err:=inp.Read(buf)
    if err!=nil || uint32(n)<length {
        return "", exception.EX_IMPROPER_DATA
    }
    return string(buf[:n]), nil
}

func (this *Kvmap)CheckOut() map[string]*KvmapEntry {
    // Attentez: All the modification will not be stored before executing CheckIn
    if this.LoadIntoMem()!=nil {
        return nil
    }

    this.lock.RLock()
    defer this.lock.RUnlock()

    this.Kvm=make(map[string]*KvmapEntry)
    this.rmed=make(map[string]*KvmapEntry)
    for _, elem:=range this.readData {
        if elem.Val==REMOVE_SPECIFIED {
            this.rmed[elem.Key]=elem
        } else {
            this.Kvm[elem.Key]=elem
        }
    }

    return this.Kvm
}
func (this *Kvmap)CheckOutReadOnly() map[string]*KvmapEntry {
    // Attentez: All the modification will not be stored before executing CheckIn
    if this.LoadIntoMem()!=nil {
        return nil
    }

    this.lock.RLock()
    defer this.lock.RUnlock()

    var ret=make(map[string]*KvmapEntry)
    for _, elem:=range this.readData {
        if elem.Val!=REMOVE_SPECIFIED {
            ret[elem.Key]=elem
        }
    }

    return ret
}
func (this *Kvmap)CheckIn() {
    this.lock.Lock()
    defer this.lock.Unlock()

    if this.Kvm==nil {
        log.Fatal("<Kvmap::CheckIn> Have not checkout yet.")
    }
    tRes:=make([]*KvmapEntry, 0)
    keyArray:=make([]string, 0)

    for key:=range this.Kvm {
        keyArray=append(keyArray,key)
    }
    for key:=range this.rmed {
        if _, ok:=this.Kvm[key]; !ok {
            keyArray=append(keyArray,key)
        }
    }
    sort.Strings(keyArray)

    for _, key:=range keyArray {
        val4kvm, ok4kvm:=this.Kvm[key]
        val4rm, ok4rm:=this.rmed[key]
        if ok4kvm && ok4rm {
            if val4kvm.Timestamp<val4rm.Timestamp {
                tRes=append(tRes, val4rm)
            } else {
                tRes=append(tRes, val4kvm)
            }
        }
        if ok4kvm && !ok4rm {
            tRes=append(tRes, val4kvm)
        }
        if !ok4kvm && ok4rm {
            tRes=append(tRes, val4rm)
        }
    }

    this.readData=tRes
}

// Attentez: deadlock may happen with incorrect co-merge!!
func (this *Kvmap)MergeWith(file2 *Kvmap) (*Kvmap, error) {
    file2.lock.Lock()
    defer file2.lock.Unlock()

    this.lock.Lock()
    defer this.lock.Unlock()

    tRes:=make([]*KvmapEntry, 0)
    file2x:=file2
    i,j:=0,0

    for {
        if this.lazyRead_NoError(i)==nil {
            for file2x.lazyRead_NoError(j)!=nil {
                tRes=append(tRes,file2x.lazyRead_NoError(j))
                j=j+1
            }
            break
        }
        if file2x.lazyRead_NoError(j)==nil {
            for this.lazyRead_NoError(i)!=nil {
                tRes=append(tRes,this.lazyRead_NoError(i))
                i=i+1
            }
            break
        }
        for this.lazyRead_NoError(i)!=nil && file2x.lazyRead_NoError(j)!=nil && this.lazyRead_NoError(i).Key<file2x.lazyRead_NoError(j).Key {
            tRes=append(tRes,this.lazyRead_NoError(i))
            i=i+1
        }
        for file2x.lazyRead_NoError(j)!=nil && this.lazyRead_NoError(i)!=nil && this.lazyRead_NoError(i).Key>file2x.lazyRead_NoError(j).Key {
            tRes=append(tRes,file2x.lazyRead_NoError(j))
            j=j+1
        }
        for file2x.lazyRead_NoError(j)!=nil && this.lazyRead_NoError(i)!=nil && this.lazyRead_NoError(i).Key==file2x.lazyRead_NoError(j).Key {
            if this.lazyRead_NoError(i).Timestamp>file2x.lazyRead_NoError(j).Timestamp {
                tRes=append(tRes,this.lazyRead_NoError(i))
            } else if this.lazyRead_NoError(i).Timestamp<file2x.lazyRead_NoError(j).Timestamp {
                tRes=append(tRes,file2x.lazyRead_NoError(j))
            } else {
                // Attentez: this conflict resolving strategy may be altered.
                tRes=append(tRes,this.lazyRead_NoError(i))
            }
            i=i+1
            j=j+1
        }
    }
    this.readData=tRes
    this.fileTS=MergeTimestamp(this.fileTS,file2x.fileTS)

    return this, nil
}

func (this *Kvmap)WriteBack(dtDes io.Writer) error {
    if err:=this.LoadIntoMem(); err!=nil {
        return err
    }

    this.lock.RLock()
    defer this.lock.RUnlock()

    if _,err:=dtDes.Write([]byte(fileMagic)); err!=nil {
        return err
    }
    for _, elem:=range this.readData {
        K:=[]byte(elem.Key)
        V:=[]byte(elem.Val)
        if err:=binary.Write(dtDes, binary.LittleEndian, uint32(len(K))); err!=nil {
            return err
        }
        if err:=binary.Write(dtDes, binary.LittleEndian, uint32(len(V))); err!=nil {
            return err
        }
        if err:=binary.Write(dtDes, binary.LittleEndian, elem.Timestamp); err!=nil {
            return err
        }
        if _,err:=dtDes.Write(K); err!=nil {
            return err
        }
        if _,err:=dtDes.Write(V); err!=nil {
            return err
        }
    }
    if err:=binary.Write(dtDes, binary.LittleEndian, uint32(0)); err!=nil {
        return err
    }
    return nil
}
func (this *Kvmap)LoadIntoMem() error {
    this.lock.Lock()
    defer this.lock.Unlock()

    for !this.finishRead {
        _, err:=this.lazyRead(len(this.readData))
        if err!=nil {
            return err
        }
    }
    return nil
}
func (this *Kvmap)EnsureRead() error {
    return this.LoadIntoMem()
}
func (this *Kvmap)lazyRead_NoError(pos int) *KvmapEntry {
    res, err:=this.lazyRead(pos)
    if err!=nil {
        return nil
    }
    return res
}
func (this *Kvmap)lazyRead(pos int) (*KvmapEntry, error) {
    if pos<len(this.readData) {
        return this.readData[pos], nil
    }
    if this.finishRead {
        return nil, nil
    }
    if len(this.readData)==0 {
        // Open the target, check it.
        tmpString, err:=ParseString(this.dataSource, 4)
        if (err!=nil) {
            return nil, exception.EX_WRONG_FILEFORMAT
        }
        if tmpString!=fileMagic {
            return nil, exception.EX_WRONG_FILEFORMAT
        }
    }

    for pos>=len(this.readData) {
        var m, n uint32
        var ts ClxTimestamp
        if binary.Read(this.dataSource, binary.LittleEndian, &n)!=nil {
            return nil, exception.EX_WRONG_FILEFORMAT
        }
        if n==0 {
            this.finishRead=true
            return nil, nil
        }
        if binary.Read(this.dataSource, binary.LittleEndian, &m)!=nil {
            return nil, exception.EX_WRONG_FILEFORMAT
        }
        if binary.Read(this.dataSource, binary.LittleEndian, &ts)!=nil {
            return nil, exception.EX_WRONG_FILEFORMAT
        }

        K, err:=ParseString(this.dataSource, n)
        if (err!=nil) {
            return nil, exception.EX_WRONG_FILEFORMAT
        }

        V, err:=ParseString(this.dataSource, m)
        if (err!=nil) {
            return nil, exception.EX_WRONG_FILEFORMAT
        }

        this.readData=append(this.readData, &KvmapEntry{
            Timestamp: ts,
            Key: K,
            Val: V,
        })
    }

    return this.readData[pos], nil
}

// Get the latest TS, from the removed version as well
// If not exist, return 0
// Must be checked out first.
func (this *Kvmap)GetRelativeTS(entry string) ClxTimestamp {
    if this.Kvm==nil {
        log.Fatal("<Kvmap::CheckIn> Have not checkout yet.")
    }
    var v1, v2 ClxTimestamp
    if v, ok:=this.Kvm[entry]; ok {
        v1=v.Timestamp
    } else {
        v1=0
    }

    if v, ok:=this.rmed[entry]; ok {
        v2=v.Timestamp
    } else {
        v2=0
    }

    return MergeTimestamp(v1, v2)
}

// Make a Kvmap with following keys in the map and empty vals, setting Timestamp to the
// system time at the present.
func FastMake(stringList ...string) *Kvmap {
    var ret=NewKvMap()
    var nowTime=GetTimestamp()
    ret.CheckOut()
    for _, elem:=range stringList {
        ret.Kvm[elem]=&KvmapEntry {
            Key: elem,
            Val: "",
            Timestamp: nowTime,
        }
    }
    ret.CheckIn()
    ret.TSet(nowTime)

    return ret
}

// Make a Kvmap with following keys in the map and vals set to REMOVED, setting Timestamp to the
// system time at the present.
func FastAntiMake(stringList ...string) *Kvmap {
    var ret=NewKvMap()
    var nowTime=GetTimestamp()
    ret.CheckOut()
    for _, elem:=range stringList {
        ret.Kvm[elem]=&KvmapEntry {
            Key: elem,
            Val: REMOVE_SPECIFIED,
            Timestamp: nowTime,
        }
    }
    ret.CheckIn()
    ret.TSet(nowTime)

    return ret
}
