package distributedvc

import (
    "sync"
    . "logger"
    "kernel/filetype"
    . "kernel/distributedvc/filemeta"
    . "kernel/distributedvc/constdef"
    . "outapi"
    "strconv"
    ex "definition/exception"
    . "definition/configinfo"
    . "utils/timestamp"
    "fmt"
    "time"
    gsp "intranet/gossip"
    gspdi "intranet/gossipd/interactive"
    "errors"
)

/*
** FD: File Descriptor
** File Descriptor is the core data structure of the S-H2, which is responsible for
** directory meta info management.
** Each FD represents a separate directory meta and is unique in the memory. It controls
** submission of patches and auto-merging. Also, any LS operation will execute it
** to read & merge all the data, while notifying random number of peers to update their
** own patch chain.
** The first segment of member variables are used for fdPool to keep it unique and supporting
** automatically wiped out to control memory cost. It has several phases:
** 1. uninited phase:   neither the file content nor the chain info is loaded into memory
** 2. dormant phase:    when .grasp() gets invoked it will load chain info into memory,
**                      then functions like .MergeNext(), .ReadInNumberZero() and .Read()
**                      could get executed
** 3. active phase:     when .graspReader() gets invoked it will loadthe file into memory,
**                      then function .Submit could get executed.
** So always GetFD()->[GraspReader()->ReleaseReader()]->Release() in use
**
*/

func __fh_go_nouse_() {
    fmt.Println("no use")
}

type FD struct {
    /*====BEGIN: for fdPool====*/
    lock *sync.Mutex

    filename string
    io Outapi
    reader int
    peeper int

    // 1 for active, 2 for dormant, 0 for uninited, -1 for dead
    status int

    isInTrash bool
    isInDormant bool
    trashNode *fdDLinkedListNode
    dormantNode *fdDLinkedListNode
    /*====END: for fdPool====*/

    /*====BEGIN: for functionality====*/
    updateChainLock *sync.RWMutex
    nextAvailablePosition int

    // only available when active
    numberZero *filetype.Kvmap
    contentLock *sync.RWMutex
    nextToBeMerge int
    lastSyncTime int64
    latestReadableVersionTS ClxTimestamp
    modified bool
    // only if merged with other nodes' patches more than once
    needsGossiped bool
}

// Lock priority: lock > updateChainLock > contentLock

const (
    INTRA_PATCH_METAKEY_NEXT_PATCH="next-patch"
)

func newFD(filename string, io Outapi) *FD {
    var ret=&FD {
        filename: filename,
        io: io,
        reader: 0,
        peeper: 0,
        status: 0,
        lock: &sync.Mutex{},
        isInDormant: false,
        isInTrash: false,

        updateChainLock: &sync.RWMutex{},
        nextAvailablePosition: -1,

        numberZero: nil,
        contentLock: &sync.RWMutex{},
        nextToBeMerge: -1,
        lastSyncTime: 0,
        latestReadableVersionTS: 0,     // This version is for written version
        modified: false,
        needsGossiped: false,
    }
    ret.trashNode=&fdDLinkedListNode {
        carrier: ret,
    }
    ret.dormantNode=&fdDLinkedListNode {
        carrier: ret,
    }

    return ret
}

func genID_static(filename string, io Outapi) string {
    return filename+"@@"+io.GenerateUniqueID()
}
func (this *FD)ID() string {
    return genID_static(this.filename, this.io)
}

func (this *FD)__clearContentSansLock() {
    if this.status!=1 {
        return
    }

    this.contentLock.Lock()
    this.numberZero=nil
    this.nextToBeMerge=-1
    this.lastSyncTime=0
    this.modified=false
    this.needsGossiped=false
    // consider write-back for unpersisted data
    this.contentLock.Unlock()

    this.status=2
}
func (this *FD)GoDie() {
    this.lock.Lock()
    defer this.lock.Unlock()

    this.WriteBack()
    this.__clearContentSansLock()
    this.status=-1

    return
}
func (this *FD)GoDormant() {
    this.lock.Lock()
    defer this.lock.Unlock()
    this.isInDormant=false

    this.WriteBack()
    //logger.Secretary.LogD("Filehandler "+this.filename+" is going dormant.")
    this.__clearContentSansLock()
}

// If not active yet, will fetch the data from storage.
// With fetching failure, a nil will be returned.
// @ Must be Grasped Reader to use
func (this *FD)Read() (map[string]*filetype.KvmapEntry, error) {
    this.contentLock.RLock()
    var t=this.numberZero
    if t!=nil {
        var q=t.CheckOutReadOnly()
        this.contentLock.RUnlock()
        return q, nil
    }
    this.contentLock.RUnlock()

    if err:=this.ReadInNumberZero(); err!=nil {
        return nil, err
    }
    this.contentLock.RLock()
    defer this.contentLock.RUnlock()
    t=this.numberZero
    return t.CheckOutReadOnly(), nil
}

// Attentez: this method is asynchonously invoked
func (this *FD)GoGrasped() {
    this.LoadPointerMap()
}

// Attentez: this method is asynchonously invoked
func (this *FD)GoRead() {
    this.ReadInNumberZero()
}

// @ indeed static
func (this *FD)GetTSFromMeta(meta FileMeta) ClxTimestamp {
    if tTS, tOK:=meta[METAKEY_TIMESTAMP]; !tOK {
        Secretary.WarnD("File "+this.filename+"'s patch #0 has invalid timestamp.")
        return 0
    } else {
        return String2ClxTimestamp(tTS)
    }
}
var READ_ZERO_NONEXISTENCE=errors.New("Patch#0 does not exist.")
// @ Must be Grasped Reader to use
func (this *FD)ReadInNumberZero() error {
    this.lock.Lock()
    defer this.lock.Unlock()

    this.contentLock.Lock()
    defer this.contentLock.Unlock()

    if this.numberZero!=nil {
        return nil
    }

    var tMeta, tFile, tErr=this.io.Get(this.GetPatchName(0, -1))
    if tErr!=nil {
        return tErr
    }
    if tFile==nil || tMeta==nil {
        return READ_ZERO_NONEXISTENCE
    }
    var tKvmap, ok=tFile.(*filetype.Kvmap)
    if !ok {
        Secretary.WarnD("File "+this.filename+"'s patch #0 has invalid filetype. Its content will get ignored.")
        this.numberZero=filetype.NewKvMap()
    } else {
        this.numberZero=tKvmap
    }
    this.numberZero.TSet(this.GetTSFromMeta(tMeta))
    if tNext, ok2:=tMeta[INTRA_PATCH_METAKEY_NEXT_PATCH]; !ok2 {
        Secretary.WarnD("File "+this.filename+"'s patch #0 has invalid next-patch. Its precedents will get ignored.")
        this.nextToBeMerge=1
    } else {
        if nextNum, errx:=strconv.Atoi(tNext); errx!=nil {
            Secretary.WarnD("File "+this.filename+"'s patch #0 has invalid next-patch. Its precedents will get ignored.")
            this.nextToBeMerge=1
        } else {
            this.nextToBeMerge=nextNum
        }
    }
    this.status=1
    this.modified=false
    this.needsGossiped=false
    return nil
}

var FORMAT_EXCEPTION=errors.New("Kvmap file not suitable.")
// if return (nil, nil), the file just does not exist.
// a nil for file and an error instance will be returned for other errors
// if the file is not nil, the function is invoked successfully
func readInKvMapfile(io Outapi, filename string) (*filetype.Kvmap, FileMeta, error) {
    var meta, file, err=io.Get(filename)
    if err!=nil {
        return nil, nil, err
    }
    if file==nil || meta==nil {
        Secretary.Log("distributedvc::readInKvMapfile()", "File "+filename+" does not exist.")
        return nil, nil, nil
    }
    var result, ok=file.(*filetype.Kvmap)
    if !ok {
        Secretary.Warn("distributedvc::readInKvMapfile()", "Fail in reading file "+filename)
        return nil, nil, FORMAT_EXCEPTION
    }

    if tTS, tOK:=meta[METAKEY_TIMESTAMP]; !tOK {
        Secretary.WarnD("File "+filename+"'s patch #0 has invalid timestamp.")
        result.TSet(0)
    } else {
        result.TSet(String2ClxTimestamp(tTS))
    }

    return result, meta, nil
}
// if return (nil, nil), the file just does not exist.
// a nil for file and an error instance will be returned for other errors
// if the file is not nil, the function is invoked successfully
func readInKvMapfile_NoWarning(io Outapi, filename string) (*filetype.Kvmap, FileMeta, error) {
    var meta, file, err=io.Get(filename)
    if err!=nil {
        return nil, nil, err
    }
    if file==nil || meta==nil {
        return nil, nil, nil
    }
    var result, ok=file.(*filetype.Kvmap)
    if !ok {
        Secretary.Warn("distributedvc::readInKvMapfile()", "Fail in reading file "+filename)
        return nil, nil, FORMAT_EXCEPTION
    }

    if tTS, tOK:=meta[METAKEY_TIMESTAMP]; !tOK {
        Secretary.WarnD("File "+filename+"'s patch #0 has invalid timestamp.")
        result.TSet(0)
    } else {
        result.TSet(String2ClxTimestamp(tTS))
    }

    return result, meta, nil
}

var MERGE_ERROR=errors.New("Merging error")

// @ Must be Grasped Reader to use
var NOTHING_TO_MERGE=errors.New("Nothing to merge.")
func (this *FD)MergeNext() error {
    if tmpErr:=this.ReadInNumberZero(); tmpErr!=nil {
        return tmpErr
    }
    // Read one patch file , get ready for merge
    this.updateChainLock.RLock()
    var nextEmptyPatch=this.nextAvailablePosition
    this.updateChainLock.RUnlock()

    this.contentLock.Lock()
    defer this.contentLock.Unlock()

    if nextEmptyPatch==this.nextToBeMerge {
        return NOTHING_TO_MERGE
    }

    var oldMerged=this.nextToBeMerge
    var thePatch, meta, err=readInKvMapfile(this.io, this.GetPatchName(this.nextToBeMerge, -1))
    // may happen due to the unsubmission of Submit() function
    if thePatch==nil {
        Secretary.Warn("distributedvc::FD.MergeNext()", "Fail to get a supposed-to-be patch for file "+this.filename)
        if err==nil {
            return MERGE_ERROR
        } else {
            return err
        }
    }
    var theNext int
    if tNext, ok:=meta[INTRA_PATCH_METAKEY_NEXT_PATCH]; !ok {
        Secretary.Warn("distributedvc::FD.MergeNext()", "Fail to get INTRA_PATCH_METAKEY_NEXT_PATCH for file "+this.filename)
        theNext=this.nextToBeMerge+1
    } else {
        if intTNext, err:=strconv.Atoi(tNext); err!=nil {
            Secretary.Warn("distributedvc::FD.MergeNext()", "Fail to get INTRA_PATCH_METAKEY_NEXT_PATCH for file "+this.filename)
            theNext=this.nextToBeMerge+1
        } else {
            theNext=intTNext
        }
    }
    tNew, err:=this.numberZero.MergeWith(thePatch)
    if err!=nil {
        Secretary.Warn("distributedvc::FD.MergeNext()", "Fail to merge patches for file "+this.filename)
        return err
    }
    this.numberZero.MergeWith(filetype.FastMake(CONF_FLAG_PREFIX+NODE_SYNC_TIME_PREFIX+strconv.Itoa(NODE_NUMBER)))

    this.numberZero=tNew
    this.nextToBeMerge=theNext
    this.modified=true
    this.needsGossiped=true

    Secretary.Log("distributedvc::FD.MergeNext()", "Successfully merged in patch #"+strconv.Itoa(oldMerged)+" for "+this.filename)
    return nil
}

/*
** Patch list: 0(the combined version) -> 1 -> 2 -> ...
** If the #0 patch does not exist, the file does not have a separate version in the node.
** otherwise, "INTRA_PATCH_METAKEY_NEXT_PATCH" in the meta will form a linked list
** to chain all the uncombined patch.
**
** As soon as the file is loaded into system, its uncombined patch will start to combine
** and the dormant fd will store the next available patch number.
*/

func (this *FD)GetPatchName(patchnumber int, nodenumber int/*-1*/) string {
    if nodenumber<0 {
        nodenumber=NODE_NUMBER
    }
    return this.filename+".node"+strconv.Itoa(nodenumber)+".patch"+strconv.Itoa(patchnumber)
}

// @ Get Normally Grasped
func (this *FD)LoadPointerMap() error {
    this.lock.Lock()
    defer this.lock.Unlock()

    this.updateChainLock.Lock()
    defer this.updateChainLock.Unlock()

    if this.nextAvailablePosition>=0 {
        return nil
    }

    var tmpPos=0
    var needMerge=false
    for {
        tMeta, tErr:=this.io.Getinfo(this.GetPatchName(tmpPos, -1))
        if tErr!=nil {
            return tErr
        }
        if tMeta==nil {
            // the file does not exist
            if this.status<=0 {
                this.status=2
            }
            this.nextAvailablePosition=tmpPos
            if needMerge {
                MergeManager.SubmitTask(this.filename, this.io)
            }
            return nil
        }
        if tNum, ok:=tMeta[INTRA_PATCH_METAKEY_NEXT_PATCH]; !ok {
            Secretary.WarnD("File "+this.filename+"'s patch #"+strconv.Itoa(tmpPos)+" has broken/invalid metadata. All the patches after it will get lost.")
            if this.status<=0 {
                this.status=2
            }
            this.nextAvailablePosition=tmpPos+1
            if needMerge {
                MergeManager.SubmitTask(this.filename, this.io)
            }
            return nil
        } else {
            var oldPos=tmpPos
            tmpPos, tErr=strconv.Atoi(tNum)
            if tErr!=nil {
                Secretary.WarnD("File "+this.filename+"'s patch #"+strconv.Itoa(tmpPos)+" has broken/invalid metadata. All the patches after it will get lost.")
                tmpPos=oldPos
                if this.status<=0 {
                    this.status=2
                }
                this.nextAvailablePosition=tmpPos+1
                if needMerge {
                    MergeManager.SubmitTask(this.filename, this.io)
                }
                return nil
            } else {
                if oldPos!=0 {
                    needMerge=true
                }
            }
        }
    }

    return nil
}

// object need not have its Timestamp set, 'cause the function will set it to
// the current systime
// @ Get Normally Grasped
func (this *FD)Submit(object *filetype.Kvmap) error {
    this.updateChainLock.Lock()
    if this.nextAvailablePosition<0 {
        this.updateChainLock.Unlock()
        if err:=this.LoadPointerMap(); err!=nil {
            return nil
        }
        this.updateChainLock.Lock()
    }
    // Insider.LogD(strconv.Itoa(this.nextAvailablePosition))
    var nAP=this.nextAvailablePosition
    this.nextAvailablePosition=nAP+1
    this.updateChainLock.Unlock()

    //Insider.Log(this.filename+".Submit()", "Release updateChainLock")

    var selfName=CONF_FLAG_PREFIX+NODE_SYNC_TIME_PREFIX+strconv.Itoa(NODE_NUMBER)
    var nowTime=GetTimestamp()

    object.CheckOut()

    object.Kvm[selfName]=&filetype.KvmapEntry {
        Key: selfName,
        Val: "",
        Timestamp: nowTime,
    }
    object.TSet(nowTime)
    object.CheckIn()

    //Insider.Log(this.filename+".Submit()", "Start to put")
    var err=this.io.Put(this.GetPatchName(nAP, -1),
                object,
                FileMeta(map[string]string {
                    INTRA_PATCH_METAKEY_NEXT_PATCH: strconv.Itoa(nAP+1),
                    METAKEY_TIMESTAMP: nowTime.String(),
                }))
    if err!=nil {
        Secretary.Warn("distributedvc::FD.Submit()", "Fail in putting file "+this.GetPatchName(nAP, -1))
        // failure rollback
        this.updateChainLock.Lock()
        if nAP+1==this.nextAvailablePosition {
            // up to now, no new patch has been submitted. Just rollback the number.
            this.nextAvailablePosition--
            this.updateChainLock.Unlock()
            return err
        } else {
            Secretary.Error("distributedvc::FD.Submit()", "Submission gap occurs! Trying to fix it: "+this.GetPatchName(nAP, -1)+" TRIAL ")

            //TODO: write in auto fix local log.
        }
        this.updateChainLock.Unlock()
        return err
    }

    //Insider.Log(this.filename+".Submit()", "Put")

    if nAP>0 {
        MergeManager.SubmitTask(this.filename, this.io)
    } else {
        // auto post 'cause it will not trigger WriteBack()
        var err=gsp.GlobalGossiper.PostGossip(&gspdi.GossipEntry{
            Filename: this.filename,
            OutAPI: this.io.GenerateUniqueID(),
            UpdateTime: nowTime,
            NodeNumber: NODE_NUMBER,
        })
        if err!=nil {
            Secretary.Warn("distributedvc::FD.WriteBack", "Fail to post change gossiping to other nodes: "+err.Error())
        }
    }


    //Insider.Log(this.filename+".Submit()", "Submitted Task")
    return nil
}

const CONF_FLAG_PREFIX="/*CONF-FLAG*/"
// NOT for header, so can be camaralized
const NODE_SYNC_TIME_PREFIX="Node-Sync-"

func (this *FD)_checkAndSubmitNumberZero_SansLock() {
    if this.nextAvailablePosition>0 {
        return
    }

    this.nextAvailablePosition++
    // submit a empty patch to number zero

    var selfName=CONF_FLAG_PREFIX+NODE_SYNC_TIME_PREFIX+strconv.Itoa(NODE_NUMBER)
    var object=filetype.NewKvMap()
    object.TSet(0)
    object.CheckOut()
    object.Kvm[selfName]=&filetype.KvmapEntry {
        Key: selfName,
        Val: "",
        Timestamp: 0,
    }
    object.CheckIn()

    var err error
    var unrevocable_io_sleep_time_dur=time.Duration(TRIAL_INTERVAL_IN_UNREVOCABLE_IOERROR)*time.Millisecond
    for {
        err=this.io.Put(this.GetPatchName(0, -1), object,
                    FileMeta(map[string]string {
                        INTRA_PATCH_METAKEY_NEXT_PATCH: strconv.Itoa(1),
                        METAKEY_TIMESTAMP: "0",
                    }))
        if err!=nil {
            Secretary.Error("distributedvc::FD._checkAndSubmitNumberZero_SansLock", "Error trying to write an empty zero patch for "+this.ID()+
                ": "+
                err.Error()+
                ". Subsequent patches may lost. TRYING to resubmit...")
            time.Sleep(unrevocable_io_sleep_time_dur)
        } else {
            break
        }
    }

    this.contentLock.Lock()
    defer this.contentLock.Unlock()

    this.numberZero=object
    this.nextToBeMerge=1
    this.latestReadableVersionTS=0
    this.modified=false
}
// will not update the status
// used by sync() only
// the function will load pointer map as LoadPointerMap() does
// however, if there's no #0 patch. It will create one and move
// nextAvailablePosition to 1
func (this *FD)_LoadPointerMap_SyncUseOnly() error {
    this.updateChainLock.Lock()
    defer this.updateChainLock.Unlock()

    if this.nextAvailablePosition>=0 {
        this._checkAndSubmitNumberZero_SansLock()
        return nil
    }

    var tmpPos=0
    var needMerge=false
    for {
        tMeta, tErr:=this.io.Getinfo(this.GetPatchName(tmpPos, -1))
        if tErr!=nil {
            return tErr
        }
        if tMeta==nil {
            // the file does not exist
            this.nextAvailablePosition=tmpPos
            if needMerge {
                MergeManager.SubmitTask(this.filename, this.io)
            }
            this._checkAndSubmitNumberZero_SansLock()
            return nil
        }
        if tNum, ok:=tMeta[INTRA_PATCH_METAKEY_NEXT_PATCH]; !ok {
            Secretary.WarnD("File "+this.filename+"'s patch #"+strconv.Itoa(tmpPos)+" has broken/invalid metadata. All the patches after it will get lost.")
            this.nextAvailablePosition=tmpPos+1
            if needMerge {
                MergeManager.SubmitTask(this.filename, this.io)
            }
            return nil
        } else {
            var oldPos=tmpPos
            tmpPos, tErr=strconv.Atoi(tNum)
            if tErr!=nil {
                Secretary.WarnD("File "+this.filename+"'s patch #"+strconv.Itoa(tmpPos)+" has broken/invalid metadata. All the patches after it will get lost.")
                tmpPos=oldPos
                this.nextAvailablePosition=tmpPos+1
                if needMerge {
                    MergeManager.SubmitTask(this.filename, this.io)
                }
                return nil
            } else {
                if oldPos!=0 {
                    needMerge=true
                }
            }
        }
    }

    return nil
}
// @ DEPRECATED
func (this *FD)__deprecated__combineNodeX(nodenumber int) error {
    if nodenumber==NODE_NUMBER {
        return nil
    }
    // First, check whether the corresponding version exists or newer than currently
    // merged version.
    var keyStoreName=CONF_FLAG_PREFIX+NODE_SYNC_TIME_PREFIX+strconv.Itoa(nodenumber)

    this.numberZero.CheckOut()

    var lastTime ClxTimestamp
    if elem, ok:=this.numberZero.Kvm[keyStoreName]; ok {
        lastTime=elem.Timestamp
    } else {
        lastTime=0
    }
    if lastTime>0 {
        var meta, err=this.io.Getinfo(this.GetPatchName(0, nodenumber))
        if meta==nil || err!=nil {
            // The file does not exist. Combining ends.
            return nil
        }
        var res, ok=meta[METAKEY_TIMESTAMP]
        if !ok {
            // The file does not exist. Combining ends.
            return nil
        }
        var existRecentTS=String2ClxTimestamp(res)
        if existRecentTS<=lastTime {
            // no need to fetch the file
            return nil
        }
    }

    var file, _, err=readInKvMapfile(this.io, this.GetPatchName(0, nodenumber))
    if err!=nil {
        return err
    }
    if file==nil {
        return nil
    }
    this.numberZero.MergeWith(file)
    this.numberZero.CheckOut()
    var newTS=GetTimestamp()
    var selfName=CONF_FLAG_PREFIX+NODE_SYNC_TIME_PREFIX+strconv.Itoa(NODE_NUMBER)
    this.numberZero.Kvm[selfName]=&filetype.KvmapEntry {
        Key: selfName,
        Val: "",
        Timestamp: newTS,
    }
    this.numberZero.TSet(newTS)
    this.numberZero.CheckIn()
    this.modified=true
    this.needsGossiped=true

    return nil
}
// Read and combine all the version from other nodes, providing the combined version.
// @ Get Reader Grasped
func (this *FD)__deprecated__Sync() error {
    var nowTime=time.Now().Unix()
    if this.lastSyncTime+SINGLE_FILE_SYNC_INTERVAL_MIN>nowTime {
        // interval is too small, abort the sync.
        return nil
    }
    // Submit #0 patch if needed
    this._LoadPointerMap_SyncUseOnly()

    // read patch 0 from container. If just submit, the function will exit immediately
    this.ReadInNumberZero()


    this.contentLock.Lock()
    defer this.contentLock.Unlock()

    if this.lastSyncTime+SINGLE_FILE_SYNC_INTERVAL_MIN>nowTime {
        // interval is too small, abort the sync.
        return nil
    }

    if this.numberZero==nil {
        Insider.Log("distributedvc::FD.Sync()", "Looks like a logical isle: this.numberZero==nil")
        Secretary.Error("distributedvc::FD.Sync()", "Looks like a logical isle: this.numberZero==nil")
        return ex.LOGICAL_ERROR
    }

    // phase1: glean information from different nodes
    // Attentez: the go routines will read numberZero.Kvm but will not write it. So not lock is required.

    this.numberZero.CheckOut()

    var updateChannel=make(chan int, NODE_NUMS_IN_ALL)
    go (func() {
        var wg=sync.WaitGroup{}
        var gleanInfo=func(nodeNumber int) bool {
            defer wg.Done()

            if nodeNumber==NODE_NUMBER {
                return false
            }
            var keyStoreName=CONF_FLAG_PREFIX+NODE_SYNC_TIME_PREFIX+strconv.Itoa(nodeNumber)
            var lastTime ClxTimestamp
            if elem, ok:=this.numberZero.Kvm[keyStoreName]; ok {
                lastTime=elem.Timestamp
            } else {
                lastTime=0
            }

            if lastTime>0 {
                var meta, err=this.io.Getinfo(this.GetPatchName(0, nodeNumber))
                if meta==nil || err!=nil {
                    if err!=nil {
                        Secretary.Warn("distributedvc::FD.Sync()", "Fail to get info from "+this.GetPatchName(0, nodeNumber)+": "+err.Error())
                    }
                    return false
                }
                var result, ok=meta[METAKEY_TIMESTAMP]
                if !ok {
                    return false
                }
                var existRecentTS=String2ClxTimestamp(result)
                if existRecentTS>lastTime {
                    updateChannel<-nodeNumber
                    return true
                }
                return false
            }

            updateChannel<-nodeNumber
            return true
        }
        wg.Add(NODE_NUMS_IN_ALL)
        for i:=0; i<NODE_NUMS_IN_ALL; i++ {
            go gleanInfo(i)
        }
        wg.Wait()
        close(updateChannel)
    })()

    // meanwhile: fetch corresponding patch as need
    var thePatch=filetype.NewKvMap()
    var changed=false
    for i:=range updateChannel {
        var file, _, err=readInKvMapfile_NoWarning(this.io, this.GetPatchName(0, i))
        if err!=nil {
            Secretary.Warn("distributedvc::FD.Sync()", "Fail to read supposed-to-be file "+this.GetPatchName(0, i))
            continue
        }
        if file==nil {
            continue
        }
        thePatch.MergeWith(file)
        changed=true
    }
    // At this time, the channel must be closed and thus, all the reading routines of
    // this.numberZero has terminated safely, on which any write is thread-safe.
    if changed {
        // merging to update modification timestamp
        this.numberZero.MergeWith(filetype.FastMake(CONF_FLAG_PREFIX+NODE_SYNC_TIME_PREFIX+strconv.Itoa(NODE_NUMBER)))
        this.numberZero.MergeWith(thePatch)
        this.modified=true
        this.needsGossiped=true
    }
    this.lastSyncTime=time.Now().Unix()

    return nil
}
// Read the current version. Submit an empty patch if needed
// @ Get Reader Grasped
func (this *FD)Sync() error {
    var nowTime=time.Now().Unix()
    if this.lastSyncTime+SINGLE_FILE_SYNC_INTERVAL_MIN>nowTime {
        // interval is too small, abort the sync.
        return nil
    }
    // Submit #0 patch if needed
    this._LoadPointerMap_SyncUseOnly()

    // read patch 0 from container. If just submit, the function will exit immediately
    this.ReadInNumberZero()

    this.lastSyncTime=time.Now().Unix()

    return nil
}
// @ Grasped Reader
// callback will get async invoked, int={
//      0: nothing posed;
//      1: post original gossip;
//      2: post temporarily nothing. A gossip will be posted when writing back
// }
// It will not writeback any change
func (this *FD)ASYNCMergeWithNodeX(context *gspdi.GossipEntry, callback func(int)) {

    this.contentLock.RLock()
    if this.needsGossiped {
        go callback(2)
    }
    this.contentLock.RUnlock()

    // Submit #0 patch if needed
    this._LoadPointerMap_SyncUseOnly()

    // read patch 0 from container. If just submit, the function will exit immediately
    this.ReadInNumberZero()

    this.contentLock.Lock()

    if this.numberZero==nil {
        Insider.Log("distributedvc::FD.ASYNCMergeWithNodeX", "Looks like a logical isle: this.numberZero==nil")
        Secretary.Error("distributedvc::FD.ASYNCMergeWithNodeX", "Looks like a logical isle: this.numberZero==nil")
        this.contentLock.Unlock()
        go callback(1)
        return
    }

    this.numberZero.CheckOut()


    var keyStoreName=CONF_FLAG_PREFIX+NODE_SYNC_TIME_PREFIX+strconv.Itoa(context.NodeNumber)
    var lastTime ClxTimestamp
    if elem, ok:=this.numberZero.Kvm[keyStoreName]; ok {
        lastTime=elem.Timestamp
    } else {
        lastTime=0
    }
    if lastTime>=context.UpdateTime {
        this.contentLock.Unlock()
        go callback(0)
        return
    }
    go callback(1)
    this.contentLock.Unlock()

    var file, _, err=readInKvMapfile_NoWarning(this.io, this.GetPatchName(0, context.NodeNumber))
    if file==nil {
        if err==nil {
            Secretary.Warn("distributedvc::FD.MergeWithNodeX", "Fail to get gossiped file: nonexist")
        } else {
            Secretary.Warn("distributedvc::FD.MergeWithNodeX", "Fail to get gossiped file: "+err.Error())
        }
        return
    }

    this.contentLock.Lock()
    defer this.contentLock.Unlock()
    if elem, ok:=this.numberZero.Kvm[keyStoreName]; ok {
        lastTime=elem.Timestamp
    } else {
        lastTime=0
    }
    if lastTime>=context.UpdateTime {
        return
    }
    if lastTime>=file.TGet() {
        return
    }
    this.numberZero.MergeWith(file)
    this.numberZero.MergeWith(filetype.FastMake(CONF_FLAG_PREFIX+NODE_SYNC_TIME_PREFIX+strconv.Itoa(NODE_NUMBER)))
    this.modified=true
}

// can be invoked after MergeWith(), Sync() or the moment that the FD goes dormant.
// @ async
func (this *FD)WriteBack() error {
    this.contentLock.Lock()
    defer this.contentLock.Unlock()

    if this.numberZero==nil {
        return nil
    }
    if !this.modified {
        return nil
    }

    var meta4Set=NewMeta()
    meta4Set[METAKEY_TIMESTAMP]=this.numberZero.TGet().String()
    meta4Set[INTRA_PATCH_METAKEY_NEXT_PATCH]=strconv.Itoa(this.nextToBeMerge)
    if err:=this.io.Put(this.GetPatchName(0, -1), this.numberZero, meta4Set); err!=nil {
        return err
    }
    if this.needsGossiped {
        var err=gsp.GlobalGossiper.PostGossip(&gspdi.GossipEntry{
            Filename: this.filename,
            OutAPI: this.io.GenerateUniqueID(),
            UpdateTime: this.numberZero.TGet(),
            NodeNumber: NODE_NUMBER,
        })
        if err!=nil {
            Secretary.Warn("distributedvc::FD.WriteBack", "Fail to post change gossiping to other nodes: "+err.Error())
        }
    }

    this.modified=false
    this.needsGossiped=false
    this.latestReadableVersionTS=this.numberZero.TGet()

    return nil
}
