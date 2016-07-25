package fmapcomposition

/*
** Fast make KvMaps which has corresponding fmapMeta, Similar to filetype.FastMake
*/

import (
    . "kernel/filetype"
    . "utils/timestamp"
    . "logger"
)

var folderInitialMeta=FMapMeta(map[string]string {
    FMAP_META_TYPE: FMAP_META_TYPE_DIRECTORY,
}).Stringify()
func FastMakeFolderPatch(stringList ...string) *Kvmap {
    var ret=NewKvMap()
    var nowTime=GetTimestamp()
    ret.CheckOut()
    for _, elem:=range stringList {
        ret.Kvm[elem]=&KvmapEntry {
            Key: elem,
            Val: folderInitialMeta,
            Timestamp: nowTime,
        }
    }
    ret.CheckIn()
    ret.TSet(nowTime)

    return ret
}


var fileInitialMeta=FMapMeta(map[string]string {
    FMAP_META_TYPE: FMAP_META_TYPE_FILE,
}).Stringify()
func FastMakeFilePatch(stringList ...string) *Kvmap {
    var ret=NewKvMap()
    var nowTime=GetTimestamp()
    ret.CheckOut()
    for _, elem:=range stringList {
        ret.Kvm[elem]=&KvmapEntry {
            Key: elem,
            Val: fileInitialMeta,
            Timestamp: nowTime,
        }
    }
    ret.CheckIn()
    ret.TSet(nowTime)

    return ret
}


// initMap could be nil
func FastMakeSingleFilePatch(filename string, initMap FMapMeta) *Kvmap {
    if initMap==nil {
        initMap=make(map[string]string)
    }
    initMap[FMAP_META_TYPE]=FMAP_META_TYPE_FILE

    var ret=NewKvMap()
    var nowTime=GetTimestamp()
    ret.CheckOut()

    ret.Kvm[filename]=&KvmapEntry {
        Key: filename,
        Val: initMap.Stringify(),
        Timestamp: nowTime,
    }

    ret.CheckIn()
    ret.TSet(nowTime)

    return ret
}

// initMap CANNOT be nil
func FastMakeWithMeta(fn string, initMap FMapMeta) *Kvmap {
    if initMap==nil {
        Insider.Log("fmapcomposition::FastMakeWithMeta()", "Invoked with nil initMap.")
        Secretary.Error("fmapcomposition::FastMakeWithMeta()", "Invoked with nil initMap.")
        return nil
    }

    var ret=NewKvMap()
    var nowTime=GetTimestamp()
    ret.CheckOut()
    ret.Kvm[fn]=&KvmapEntry {
        Key: fn,
        Val: initMap.Stringify(),
        Timestamp: nowTime,
    }
    ret.CheckIn()
    ret.TSet(nowTime)

    return ret
}
