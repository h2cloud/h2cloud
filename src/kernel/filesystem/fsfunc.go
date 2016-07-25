package filesystem
// Most impotant implementation of pseudo-filesystem layer.

import (
    "outapi"
    "definition/exception"
    egg "definition/errorgroup"
    dvc "kernel/distributedvc"
    "kernel/filetype"
    "utils/uniqueid"
    "strings"
    "strconv"
    . "kernel/distributedvc/filemeta"
    . "kernel/distributedvc/constdef"
    . "utils/timestamp"
    . "logger"
    fc "kernel/filesystem/fmapcomposition"
    "io"
    "sync"
    "fmt"
)

// ## META_PARENT_INODE & META_INODE_TYPE are set to inode

// Parent inode meta changes faster than inode::.. file, and when moved, it is
// synchronously updated. So it can be used for more reliable and faster check
const META_PARENT_INODE="parent-inode"

// can be file or a folder
const META_INODE_TYPE="inode-type"
const META_INODE_TYPE_FOLDER="DIR"
const META_INODE_TYPE_FILE="FILE"

const FMAP_METAKEY_FILE_PREFIX="fmap-file-"
func InferFMapMetaFromNNodeMeta(obj FileMeta) fc.FMapMeta {
    if obj[META_INODE_TYPE]==META_INODE_TYPE_FILE {
        var ret=fc.FMapMeta(map[string]string {
            fc.FMAP_META_TYPE: fc.FMAP_META_TYPE_FILE,
        })
        for k, v:=range obj {
            if strings.HasPrefix(k, FMAP_METAKEY_FILE_PREFIX) {
                ret[k]=v
            }
        }
        return ret
    } else {
        return fc.FMapMeta(map[string]string {
            fc.FMAP_META_TYPE: fc.FMAP_META_TYPE_DIRECTORY,
        })
    }
}

func ConvertFileHeaderToNNodeMeta(header map[string]string) FileMeta {
    if header==nil {
        return nil
    }
    var ret=make(map[string]string)
    for k, v:=range header {
        ret[FMAP_METAKEY_FILE_PREFIX+k]=v
    }

    return ret
}

// ## META_ORIGINAL_NAME is set to object node
// Identical info which can be inferred by its namenode in its parent inode.
// It is synchronously updated in move operation
const META_ORIGINAL_NAME="file-original-name"

const ROOT_INODE_NAME="rootNode"

type Fs struct {
    io outapi.Outapi
    rootName string

    cLock *sync.RWMutex
    trashInode string
}
func __nouse__() {
    fmt.Println("123")
}

// for internal use only
func newFs(_io outapi.Outapi) *Fs {
    return &Fs{
        io: _io,
        rootName: ROOT_INODE_NAME,

        cLock: &sync.RWMutex{},
        trashInode: "",
    }
}

const FOLDER_MAP="/$Folder-Map/"
const TRASH_BOX=".trash"

func (this *Fs)GetTrashInode() string {
    this.cLock.RLock()
    if t:=this.trashInode; t!="" {
        this.cLock.RUnlock()
        return t
    }
    this.cLock.RUnlock()
    this.cLock.Lock()
    defer this.cLock.Unlock()
    if this.trashInode!="" {
        return this.trashInode
    }
    // fetch it from storage
    var _, file, _=this.io.Get(GenFileName(this.rootName, TRASH_BOX))
    var filen, _=file.(*filetype.Nnode)
    if filen==nil {
        return ""
    } else {
        return filen.DesName
    }
}

//==============================================================================
// Followings are filesystem functions:

// path is a unix-like path string. If path starts with "/", search begins at
// root node. Otherwise in the frominode folder, when the frominode must exist.
// For any error, a blank string and error will be returned.
func (this *Fs)Locate(path string, frominode string/*=""*/) (string, error) {
    if strings.HasPrefix(path, "/") || frominode=="" {
        frominode=this.rootName
    }
    var rawResult=strings.Split(path, "/")
    for _, e:=range rawResult {
        if e!="" {
            frominode, _=lookUp(frominode, e, this.io)
            if frominode=="" {
                // It is correct only to check result without referring to error.
                return "", exception.EX_FAIL_TO_LOOKUP
            }
        }
    }

    return frominode, nil
}

// If the file exist and forceMake==false, an error EX_FOLDER_ALREADY_EXIST will be returned
func (this *Fs)Mkdir(foldername string, frominode string, forceMake bool) error {
    if !CheckValidFilename(foldername) {
        return exception.EX_INVALID_FILENAME
    }

    // nnodeName: parentInode::foldername
    var nnodeName=GenFileName(frominode, foldername)
    if !forceMake {
        if tmeta, _:=this.io.Getinfo(nnodeName); tmeta!=nil {
            return exception.EX_FOLDER_ALREADY_EXIST
        }
    }

    // newDomainname: <GENERATED>
    var newDomainname=uniqueid.GenGlobalUniqueName()
    var newNnode=filetype.NewNnode(newDomainname)
    var initMeta=FileMeta(map[string]string {
        META_INODE_TYPE:    META_INODE_TYPE_FOLDER,
        META_PARENT_INODE:  frominode,
    })
    if err:=this.io.Put(nnodeName, newNnode, initMeta); err!=nil {
        return err
    }
    // initialize two basic element
    var initMetaSelf=FileMeta(map[string]string {
        META_INODE_TYPE:    META_INODE_TYPE_FOLDER,
        META_PARENT_INODE:  newDomainname,
    })
    if err:=this.io.Put(GenFileName(newDomainname, ".."), filetype.NewNnode(frominode), initMetaSelf); err!=nil {
        Secretary.Error("kernel.filesystem::Mkdir", "Fail to create .. link for new folder "+nnodeName+".")
        return err
    }

    if err:=this.io.Put(GenFileName(newDomainname, "."), filetype.NewNnode(newDomainname), initMetaSelf); err!=nil {
        Secretary.Error("kernel.filesystem::Mkdir", "Fail to create . link for new folder "+nnodeName+".")
        return err
    }

    // write new folder's map
    {
        var newFolderMapFD=dvc.GetFD(GenFileName(newDomainname, FOLDER_MAP), this.io)
        if newFolderMapFD==nil {
            Secretary.Error("kernel.filesystem::Mkdir", "Fail to create foldermap fd for new folder "+nnodeName+".")
            return exception.EX_IO_ERROR
        }
        if err:=newFolderMapFD.Submit(fc.FastMakeFolderPatch(".", "..")); err!=nil {
            Secretary.Error("kernel.filesystem::Mkdir", "Fail to init foldermap for new folder "+nnodeName+".")
            newFolderMapFD.Release()
            return err
        }
        newFolderMapFD.Release()
    }

    // submit patch to parent folder's map
    {
        var parentFolderMapFD=dvc.GetFD(GenFileName(frominode, FOLDER_MAP), this.io)
        if parentFolderMapFD==nil {
            Secretary.Error("kernel.filesystem::Mkdir", "Fail to create foldermap fd for new folder "+nnodeName+"'s parent map'")
            return exception.EX_IO_ERROR
        }
        if err:=parentFolderMapFD.Submit(fc.FastMakeFolderPatch(foldername)); err!=nil {
            Secretary.Error("kernel.filesystem::Mkdir", "Fail to submit patch to foldermap for new folder "+nnodeName+"'s parent map'")
            parentFolderMapFD.Release()
            return err
        }
        parentFolderMapFD.Release()
    }

    return nil
}

// Format the filesystem.
// TODO: Setup clear old fs info? Up to now set up will not clear old data and will not
// remove the old folder map
func (this *Fs)FormatFS() error {
    var initMetaSelf=FileMeta(map[string]string {
        META_INODE_TYPE:    META_INODE_TYPE_FOLDER,
        META_PARENT_INODE:  this.rootName,
    })
    if err:=this.io.Put(GenFileName(this.rootName, ".."), filetype.NewNnode(this.rootName), initMetaSelf); err!=nil {
        Secretary.Error("kernel.filesystem::FormatFS", "Fail to create .. link for Root.")
        return err
    }

    if err:=this.io.Put(GenFileName(this.rootName, "."), filetype.NewNnode(this.rootName), initMetaSelf); err!=nil {
        Secretary.Error("kernel.filesystem::FormatFS", "Fail to create . link for Root.")
        return err
    }

    {
        var rootFD=dvc.GetFD(GenFileName(this.rootName, FOLDER_MAP), this.io)
        if rootFD==nil {
            Secretary.Error("kernel.filesystem::FormatFS", "Fail to get FD for Root.")
            return exception.EX_IO_ERROR
        }
        if err:=rootFD.Submit(fc.FastMakeFolderPatch(".", "..")); err!=nil {
            Secretary.Error("kernel.filesystem::FormatFS", "Fail to submit format patch for Root.")
            rootFD.Release()
            return nil
        }
        rootFD.Release()
    }
    // setup Trash for users
    return this.Mkdir(TRASH_BOX, this.rootName, true)
}

// Only returns file name list of one inode. Innername excluded.
func (this *Fs)List(frominode string) ([]string, error) {
    var fd=dvc.GetFD(GenFileName(frominode, FOLDER_MAP), this.io)
    if fd==nil {
        Secretary.Error("kernel.filesystem::List", "Fail to get FD for "+frominode)
        return nil, exception.EX_IO_ERROR
    }
    defer fd.Release()
    fd.GraspReader()
    defer fd.ReleaseReader()

    if err:=fd.Sync(); err!=nil {
        Secretary.Error("kernel.filesystem::List", "SYNC error for "+frominode+": "+err.Error())
    }
    if content, err:=fd.Read(); err!=nil {
        Secretary.Error("kernel.filesystem::List", "Read error for "+frominode+": "+err.Error())
        return nil, err
    } else {
        var ret=[]string{}
        for k, _:=range content {
            if CheckValidFilename(k) {
                ret=append(ret, k)
            }
        }
        return ret, nil
    }
}

func (this *Fs)ListX(frominode string) ([]*filetype.KvmapEntry, error) {
    var fd=dvc.GetFD(GenFileName(frominode, FOLDER_MAP), this.io)
    if fd==nil {
        Secretary.Error("kernel.filesystem::List", "Fail to get FD for "+frominode)
        return nil, exception.EX_IO_ERROR
    }
    defer fd.Release()
    fd.GraspReader()
    defer fd.ReleaseReader()

    if err:=fd.Sync(); err!=nil {
        Secretary.Error("kernel.filesystem::List", "SYNC error for "+frominode+": "+err.Error())
    }
    if content, err:=fd.Read(); err!=nil {
        Secretary.Error("kernel.filesystem::List", "Read error for "+frominode+": "+err.Error())
        return nil, err
    } else {
        var ret=[]*filetype.KvmapEntry{}
        for k, v:=range content {
            if CheckValidFilename(k) {
                ret=append(ret, v)
            }
        }
        return ret, nil
    }
}

// list all including non-file ones
func (this *Fs)ListXPP(frominode string) ([]*filetype.KvmapEntry, error) {
    var fd=dvc.GetFD(GenFileName(frominode, FOLDER_MAP), this.io)
    if fd==nil {
        Secretary.Error("kernel.filesystem::List", "Fail to get FD for "+frominode)
        return nil, exception.EX_IO_ERROR
    }
    defer fd.Release()
    fd.GraspReader()
    defer fd.ReleaseReader()

    if err:=fd.Sync(); err!=nil {
        Secretary.Error("kernel.filesystem::List", "SYNC error for "+frominode+": "+err.Error())
    }
    if content, err:=fd.Read(); err!=nil {
        Secretary.Error("kernel.filesystem::List", "Read error for "+frominode+": "+err.Error())
        return nil, err
    } else {
        var ret=[]*filetype.KvmapEntry{}
        for _, v:=range content {
            ret=append(ret, v)
        }
        return ret, nil
    }
}

// All the folder will be removed. No matter if it is empty or not.
// Move it to the trash
func (this *Fs)Rm(foldername string, frominode string) error {
    if tsinode:=this.GetTrashInode(); tsinode=="" {
        Secretary.ErrorD("IO: "+this.io.GenerateUniqueID()+" has an invalid trashbox, which leads to removing failure.")
        return exception.EX_TRASHBOX_NOT_INITED
    } else {
        return this.MvX(foldername, frominode, uniqueid.GenGlobalUniqueNameWithTag("removed"), tsinode, true)
        // TODO: logging the original position for recovery
    }
}

// Attentez: It is not atomic
// If byForce set to false and the destination file exists, an EX_FOLDER_ALREADY_EXIST will be returned
func (this *Fs)MvX(srcName, srcInode, desName, desInode string, byForce bool) error {
    // Create a mirror at destination position.
    // Then, remove the old one.
    // Third, modify the .. pointer.

    if !CheckValidFilename(srcName) || !CheckValidFilename(desName) {
        return exception.EX_INVALID_FILENAME
    }
    if !byForce && outapi.ForceCheckExist(this.io.CheckExist(GenFileName(desInode, desName))) {
        return exception.EX_FOLDER_ALREADY_EXIST
    }

    var modifiedMeta=FileMeta(map[string]string {
        META_PARENT_INODE: desInode,
    })
    if err:=this.io.Copy(GenFileName(srcInode, srcName), GenFileName(desInode, desName), modifiedMeta); err!=nil {
        return exception.EX_FILE_NOT_EXIST
    }

    // remove the old one.
    this.io.Delete(GenFileName(srcInode, srcName))

    {
        var srcParentMap=dvc.GetFD(GenFileName(srcInode, FOLDER_MAP), this.io)
        if srcParentMap==nil {
            Secretary.Error("kernel.filesystem::MvX", "Fail to get foldermap fd for folder "+srcInode)
            return exception.EX_IO_ERROR
        }
        if err:=srcParentMap.Submit(filetype.FastAntiMake(srcName)); err!=nil {
            Secretary.Error("kernel.filesystem::MvX", "Fail to submit foldermap patch for folder "+srcInode)
            srcParentMap.Release()
            return err
        }
        srcParentMap.Release()
    }

    // modify the .. pointer
    var dstMeta, dstFileNnodeOriginal, _=this.io.Get(GenFileName(desInode, desName))
    var dstFileNnode, _=dstFileNnodeOriginal.(*filetype.Nnode)
    if dstFileNnode==nil {
        Secretary.Error("kernel.filesystem::MvX", "Fail to read nnode "+GenFileName(desInode, desName)+".")
        return exception.EX_IO_ERROR
    }

    {
        var desParentMap=dvc.GetFD(GenFileName(desInode, FOLDER_MAP), this.io)
        if desParentMap==nil {
            Secretary.Error("kernel.filesystem::MvX", "Fail to get foldermap fd for folder "+desInode)
            return exception.EX_IO_ERROR
        }
        if err:=desParentMap.Submit(fc.FastMakeWithMeta(desName, InferFMapMetaFromNNodeMeta(dstMeta))); err!=nil {
            Secretary.Error("kernel.filesystem::MvX", "Fail to submit foldermap patch for folder "+desInode)
            desParentMap.Release()
            return err
        }
        desParentMap.Release()
    }

    var target=GenFileName(dstFileNnode.DesName, "..")
    if err:=this.io.Put(target, filetype.NewNnode(desInode), nil); err!=nil {
        Secretary.Error("kernel.filesystem::MvX", "Fail to modify .. link for "+dstFileNnode.DesName+".")
        return err
    } else {
        //Secretary.Log("kernel.filesystem::MvX", "Update file "+target)
    }

    // ALL DONE!
    return nil

}

// To put a large file and modify its corresponding index.
// Note that the function is synchronous, which means that it
// will block until data are fully written.
// It will try to put a file at destination, no matter whether
// there's already one file, which will be replaced then.

// if filename!="", a new filename will be assigned and frominode::filename will be set (create mode)
// otherwise, frominode indicates the target fileinode and the target file will override it (override mode)

// the second value indicates the manipulated file inode. It may be some valid value
// or just empty no matter whether error==nil, however, when error==nil, the second value
// must be valid

// Attentez: in override mode, parent folder's foldermap will not get updated.
const STREAM_TYPE="stream type file"
func (this *Fs)Put(filename string, frominode string, meta FileMeta/*=nil*/, dataSource io.Reader) (error, string) {
    var targetFileinode string
    var oldOriName string
    if filename!="" {
        // CREATE MODE
        if !CheckValidFilename(filename) {
            return exception.EX_INVALID_FILENAME, ""
        }
        // set inode
        targetFileinode=uniqueid.GenGlobalUniqueNameWithTag("Stream")
    } else {
        // OVERRIDE MODE
        var oldMeta, err=this.io.Getinfo(frominode)
        if oldMeta==nil {
            if err!=nil {
                return err, ""
            }
            return exception.EX_FILE_NOT_EXIST, ""
        }
        oldOriName=oldMeta[META_ORIGINAL_NAME]
        targetFileinode=frominode
    }

    // set object node
    if meta==nil {
        meta=NewMeta()
    }
    meta=meta.Clone()
    meta[METAKEY_TYPE]=STREAM_TYPE
    if filename!="" {
        meta[META_ORIGINAL_NAME]=filename
    } else {
        meta[META_ORIGINAL_NAME]=oldOriName
    }

    if wc, err:=this.io.PutStream(targetFileinode, meta); err!=nil {
        Secretary.Error("kernel.filesystem::Put", "Put stream for new file "+GenFileName(frominode, filename)+" failed.")
        return err, targetFileinode
    } else {
        if _, err2:=io.Copy(wc, dataSource); err2!=nil {
            wc.Close()
            Secretary.Error("kernel.filesystem::Put", "Piping stream for new file "+GenFileName(frominode, filename)+" failed.")
            return err2, targetFileinode
        }
        if err2:=wc.Close(); err2!=nil {
            Secretary.Error("kernel.filesystem::Put", "Close writer for new file "+GenFileName(frominode, filename)+" failed.")
            return err2, targetFileinode
        }
    }

    if filename!="" {
        // CREATE MODE. Set its parent node's foldermap and write the nnode concurrently
        var currentHeader, terr=this.io.GetinfoX(targetFileinode)
        if currentHeader==nil {
            if terr!=nil {
                Secretary.Warn("kernel.filesystem::Put", "Fail to refetch supposed-to-be file meta: "+targetFileinode+". Error is "+terr.Error())
            } else {
                Secretary.Warn("kernel.filesystem::Put", "Fail to refetch supposed-to-be file meta: "+targetFileinode+". The file seems to be non-existence.")
            }

            return exception.EX_CONCURRENT_CHAOS, targetFileinode
        }
        //fmt.Println(currentHeader)
        var pointedMeta=ConvertFileHeaderToNNodeMeta(currentHeader)
        var metaToSet=pointedMeta
        metaToSet[META_PARENT_INODE]=frominode
        metaToSet[META_INODE_TYPE]=META_INODE_TYPE_FILE

        var wg=sync.WaitGroup{}
        var globalError *egg.ErrorAssembly=nil
        var geLock sync.Mutex

        wg.Add(2)
        go (func() {
            defer wg.Done()

            // Write the nnode

            if err:=this.io.Put(GenFileName(frominode, filename), filetype.NewNnode(targetFileinode), metaToSet); err!=nil {
                Secretary.Warn("kernel.filesystem::Put", "Put nnode for new file "+GenFileName(frominode, filename)+" failed.")
                geLock.Lock()
                globalError=egg.AddIn(globalError, err)
                geLock.Unlock()
                return
            }
        })()

        go (func() {
            // update parentNode's foldermap
            defer wg.Done()

            var parentFD=dvc.GetFD(GenFileName(frominode, FOLDER_MAP), this.io)
            if parentFD==nil {
                Secretary.Error("kernel.filesystem::Put", "Get FD for "+GenFileName(frominode, FOLDER_MAP)+" failed.")
                geLock.Lock()
                globalError=egg.AddIn(globalError, exception.EX_INDEX_ERROR)
                geLock.Unlock()
                return
            }
            if err:=parentFD.Submit(fc.FastMakeWithMeta(filename, InferFMapMetaFromNNodeMeta(metaToSet))); err!=nil {
                Secretary.Error("kernel.filesystem::Put", "Submit patch for "+GenFileName(frominode, filename)+" failed: "+err.Error())
                parentFD.Release()
                geLock.Lock()
                globalError=egg.AddIn(globalError, exception.EX_INDEX_ERROR)
                geLock.Unlock()
                return
            }
            parentFD.Release()
        })()
        wg.Wait()

        if globalError!=nil {
            return globalError, targetFileinode
        }
    }

    return nil, targetFileinode
}

// If the file does not exist, an EX_FILE_NOT_EXIST will be returned.
// the callback parameters are (objectNode, originalName, fileMeta)
func (this *Fs)Get(filename string, frominode string, beforePipe func(string, string, map[string]string) io.Writer) error {
    var targetFileinode string
    if filename!="" {
        if !CheckValidFilename(filename) {
            return exception.EX_INVALID_FILENAME
        }
        var meta, file, _=this.io.Get(GenFileName(frominode, filename))
        var filen, _=file.(*filetype.Nnode)
        if filen==nil {
            return exception.EX_FILE_NOT_EXIST
        }
        if meta==nil || meta[META_INODE_TYPE]!=META_INODE_TYPE_FILE {
            return exception.EX_WRONG_FILEFORMAT
        }
        targetFileinode=filen.DesName
    } else {
        targetFileinode=frominode
    }

    var oriMeta, rc, _=this.io.GetStreamX(targetFileinode)
    if oriMeta==nil || rc==nil {
        return exception.EX_FILE_NOT_EXIST
    }
    var meta=this.io.ExtractFileMeta(oriMeta)
    if val, ok:=meta[METAKEY_TYPE]; !ok || val!=STREAM_TYPE {
        rc.Close()
        return exception.EX_WRONG_FILEFORMAT
    }

    if beforePipe==nil {
        rc.Close()
        return nil
    }
    var w=beforePipe(targetFileinode, meta[META_ORIGINAL_NAME], oriMeta)
    if _, copyErr:=io.Copy(w, rc); copyErr!=nil {
        rc.Close()
        return copyErr
    }
    if err2:=rc.Close(); err2!=nil {
        return err2
    }
    return nil
}

func (this *Fs)BatchPutDir(filenameprefix string, frominode string, fromn int, ton int, content string) error {
    var kvmp=filetype.NewKvMap()
    var nowTime=GetTimestamp()
    kvmp.CheckOut()

    for i:=fromn; i<ton; i++ {
        var filename=filenameprefix+strconv.Itoa(i)
        var nnodeName=GenFileName(frominode, filename)
        var newNnode=filetype.NewNnode(content)
        var initMeta=FileMeta(map[string]string {
            META_INODE_TYPE:    META_INODE_TYPE_FOLDER,
            META_PARENT_INODE:  frominode,
        })
        if err:=this.io.Put(nnodeName, newNnode, initMeta); err!=nil {
            Secretary.Error("kernel/filesystem::Fs::BatchPutDir", "Error when putting "+filename+": "+err.Error())
            return err
        }
        kvmp.Kvm[filename]=&filetype.KvmapEntry {
            Key: filename,
            Val: "",
            Timestamp: nowTime,
        }
    }
    kvmp.CheckIn()
    kvmp.TSet(nowTime)

    {
        var parentFolderMapFD=dvc.GetFD(GenFileName(frominode, FOLDER_MAP), this.io)
        if parentFolderMapFD==nil {
            Secretary.Error("kernel.filesystem::BatchPutDir", "Fail to create foldermap")
            return exception.EX_IO_ERROR
        }
        if err:=parentFolderMapFD.Submit(kvmp); err!=nil {
            Secretary.Error("kernel.filesystem::BatchPutDir", "Fail to submit patch to foldermap")
            parentFolderMapFD.Release()
            return err
        }
        parentFolderMapFD.Release()
    }

    return nil
}
