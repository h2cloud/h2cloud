package fsmanage

// APIs for managing pseudo-filesystem, not for managing files
// Attentions:
// - Move/Remove items are multi-invocation-unsafe. Concurrent moves on a file may
//   lead to multi-existence of one single nnode.

import (
    . "github.com/levythu/gurgling"
    "outapi"
    "kernel/filesystem"
    "definition/exception"
    "fmt"
    "utils/pathman"
    egg "definition/errorgroup"
    "time"
    "strconv"
    "kernel/filetype"
    "sync"
    //"logger"
)


func __fsmanagego_nouse() {
    fmt.Println("nouse")
}

func FMRouter() Router {
    var rootRouter=ARegexpRouter()
    rootRouter.Use(`/([^/]+)/\[\[SC\](.+)\]/?(.*)`, handlingShortcut)

    rootRouter.Get(`/([^/]+)/(.*)`, lsDirectory)
    rootRouter.Put(`/([^/]+)/(.*)`, mkDirectoryByForce)
    rootRouter.Post(`/([^/]+)/(.*)`, mkDirectory)
    rootRouter.Delete(`/([^/]+)/(.*)`, rmDirectory)
    rootRouter.UseSpecified(`/([^/]+)/(.*)`, "PATCH", mvDirectory, true)

    return rootRouter
}

const LAST_PARENT_NODE="Parent-Node"

// Handling shortcut retrieve. It's applied to all the api in the field
// format: /fs/{contianer}/[[SC]{rootnode}]/{followingpath}
func handlingShortcut(req Request, res Response) bool {
    // After the midware,
    // req.F()["HandledRR"][1]=={container},
    // req.F()["HandledRR"][2]=={followingpath},
    // req.F()["HandledRR"][3]=={rootnode},
    // note that if no shortcut is specified, there should not be [3]
    var matchRes=req.F()["RR"].([]string)
    var t=matchRes[2]
    matchRes[2]=matchRes[3]
    matchRes[3]=t
    req.F()["HandledRR"]=matchRes

    return true
}

// ==========================API DOCS=======================================
// API Name: List all the object in the directory
// Action: Return all the file in the format of JSON
// API URL: /fs/{contianer}/{followingpath}
// REQUEST: GET
// Parameters:
//      - contianer(in URL): the container name
//      - followingpath(in URL): the path to be listed
//      - Show-All(in Header): TRUE to show all in the kvfile
// Returns:
//      - HTTP 200: No error and the result will be returned in JSON in the body.
//              When success, 'Parent-Node' will indicate the listed directory.
//      - HTTP 404: Either the container or the filepath does not exist.
//      - HTTP 500: Error. The body is supposed to return error info.
// ==========================API DOCS END===================================
func lsDirectory(req Request, res Response) {

    //===============measurement========================
    var startTime int64=0
    if req.Get("Enable-Measure")=="True" {
        startTime=time.Now().UnixNano()
    }
    //==================================================

    var pathDetail, _=req.F()["HandledRR"].([]string)
    if pathDetail==nil {
        pathDetail=req.F()["RR"].([]string)
        pathDetail=append(pathDetail, filesystem.ROOT_INODE_NAME)
    }

    var fs=filesystem.GetFs(outapi.NewSwiftio(outapi.DefaultConnector, pathDetail[1]))
    if fs==nil {
        res.Status("Internal Error: the FS pool is full.", 500)
    }
    defer fs.Release()

    var nodeName, err=fs.Locate(pathDetail[2], pathDetail[3])
    if err!=nil {
        res.Status("Nonexist container or path. "+err.Error(), 404)
        return
    }
    var resultList []*filetype.KvmapEntry
    if req.Get("Show-All")=="TRUE" {
        resultList, err=fs.ListXPP(nodeName)
    } else {
        resultList, err=fs.ListX(nodeName)
    }
    if err!=nil {
        res.Status("Reading error: "+err.Error(), 404)
        return
    }

    res.Set(LAST_PARENT_NODE, nodeName)
    //=============================================
    if startTime>0 {
        res.Set("Time-Consumed", strconv.FormatInt(time.Now().UnixNano()-startTime, 10))
    }
    //=============================================
    res.JSON(resultList)
}

const HEADER_DISTABLE_PARALLEL="Disable-Parallel"

// ==========================API DOCS=======================================
// API Name: Make one directory
// Action: make the directory only if it does not exist and its parent path exists
// API URL: /fs/{contianer}/{followingpath}
// REQUEST: POST
// Parameters:
//      - contianer(in URL): the container name
//      - followingpath(in URL): the path to be create. Please guarantee its parent node exists.
//      - Disable-Parallel(in Header): if set to TRUE, a non-parallelized mkdir will be
//              operated. Default to FALSE, it is only for test and not recommend to set.
// Returns:
//      - HTTP 201: No error and the directory creation application has been submitted.
//        to ensure created, another list operation should be carried.
//              When success, 'Parent-Node' will indicate the parent of created directory.
//      - HTTP 202: No error but the directory has existed before.
//              When success, 'Parent-Node' will indicate the parent of already exist directory.
//      - HTTP 404: Either the container or the parent filepath does not exist.
//      - HTTP 500: Error. The body is supposed to return error info.
// ==========================API DOCS END===================================
// ==========================API DOCS=======================================
// API Name: Make one directory by force
// Action: make the directory by force if its parent path exists
// API URL: /fs/{contianer}/{followingpath}
// REQUEST: PUT
// Parameters:
//      - contianer(in URL): the container name
//      - followingpath(in URL): the path to be create. Please guarantee its parent node exists.
//      - Disable-Parallel(in Header): if set to TRUE, a non-parallelized mkdir will be
//              operated. Default to FALSE, it is only for test and not recommend to set.
// Returns:
//      - HTTP 201: No error and the directory creation application has been submitted.
//        to ensure created, another list operation should be carried.
//              When success, 'Parent-Node' will indicate the parent of created directory.
//      - HTTP 404: Either the container or the parent filepath does not exist.
//      - HTTP 500: Error. The body is supposed to return error info.
// ==========================API DOCS END===================================
func mkDirectoryX(req Request, res Response, byforce bool) {

    //===============measurement========================
    var startTime int64=0
    if req.Get("Enable-Measure")=="True" {
        startTime=time.Now().UnixNano()
    }
    //==================================================

    var pathDetail, _=req.F()["HandledRR"].([]string)
    if pathDetail==nil {
        pathDetail=req.F()["RR"].([]string)
        pathDetail=append(pathDetail, filesystem.ROOT_INODE_NAME)
    }

    var trimer=pathDetail[2]
    var i int
    for i=len(trimer)-1; i>=0; i-- {
        if trimer[i]!='/' {
            break
        }
    }
    if i<0 {
        res.Status("The directory to create should be specified.", 404)
        return
    }
    trimer=trimer[:i+1]

    // Now trimmer eleminates all the trailing slashes

    var j int
    for j=i; j>=0; j-- {
        if trimer[j]=='/' {
            break
        }
    }
    var base=trimer[:j+1]
    trimer=trimer[j+1:]
    // now trimer holds the last foldername
    // base holds the parent folder path

    var fs=filesystem.GetFs(outapi.NewSwiftio(outapi.DefaultConnector, pathDetail[1]))
    if fs==nil {
        res.Status("Internal Error: the FS pool is full.", 500)
    }
    defer fs.Release()

    var nodeName, err=fs.Locate(base, pathDetail[3])
    if err!=nil {
        res.Status("Nonexist container or path. "+err.Error(), 404)
        return
    }

    res.Set(LAST_PARENT_NODE, nodeName)
    if req.Get(HEADER_DISTABLE_PARALLEL)=="TRUE" {
        err=fs.Mkdir(trimer, nodeName, byforce)
    } else {
        err=fs.MkdirParalleled(trimer, nodeName, byforce)
    }
    if !egg.Nil(err) {
        if egg.In(err, exception.EX_INODE_NONEXIST) {
            res.Status("Nonexist container or path.", 404)
            return
        }
        if egg.In(err, exception.EX_FOLDER_ALREADY_EXIST) {
            res.SendCode(202)
            return
        }
        res.Status("Internal Error: "+err.Error(), 500)
        return
    }

    //=============================================
    if startTime>0 {
        res.Set("Time-Consumed", strconv.FormatInt(time.Now().UnixNano()-startTime, 10))
    }
    //=============================================
    res.SendCode(201)
}
func mkDirectoryByForce(req Request, res Response) {
    mkDirectoryX(req, res, true)
}
func mkDirectory(req Request, res Response) {
    mkDirectoryX(req, res, false)
}


// ==========================API DOCS=======================================
// API Name: Remove one directory
// Action: remove the directory only if it exists and its parent path exists
// API URL: /fs/{contianer}/{followingpath}
// REQUEST: DELETE
// Parameters:
//      - contianer(in URL): the container name
//      - followingpath(in URL): the path to be removed. Please guarantee its parent node exists.
//      - Disable-Parallel(in Header): if set to TRUE, a non-parallelized mkdir will be
//              operated. Default to FALSE, it is only for test and not recommend to set.
// Returns:
//      - HTTP 204: The deletion succeeds but it is only a patch. to ensure created, another list
//        operation should be carried.
//              When success, 'Parent-Node' will indicate the parent of removed directory.
//      - HTTP 404: Either the container or the parent filepath or the file itself does not exist.
//      - HTTP 500: Error. The body is supposed to return error info.
// ==========================API DOCS END===================================
func rmDirectory(req Request, res Response) {
    //===============measurement========================
    var startTime int64=0
    if req.Get("Enable-Measure")=="True" {
        startTime=time.Now().UnixNano()
    }
    //==================================================

    var pathDetail, _=req.F()["HandledRR"].([]string)
    if pathDetail==nil {
        pathDetail=req.F()["RR"].([]string)
        pathDetail=append(pathDetail, filesystem.ROOT_INODE_NAME)
    }

    var trimer=pathDetail[2]
    var i int
    for i=len(trimer)-1; i>=0; i-- {
        if trimer[i]!='/' {
            break
        }
    }
    if i<0 {
        res.Status("The directory to remove should be specified.", 404)
        return
    }
    trimer=trimer[:i+1]
    var j int
    for j=i; j>=0; j-- {
        if trimer[j]=='/' {
            break
        }
    }
    var base=trimer[:j+1]
    trimer=trimer[j+1:]
    // now trimer holds the last foldername
    // base holds the parent folder path

    var fs=filesystem.GetFs(outapi.NewSwiftio(outapi.DefaultConnector, pathDetail[1]))
    if fs==nil {
        res.Status("Internal Error: the FS pool is full.", 500)
    }
    defer fs.Release()

    var nodeName, err=fs.Locate(base, pathDetail[3])
    if err!=nil {
        res.Status("Nonexist container or path. "+err.Error(), 404)
        return
    }

    res.Set(LAST_PARENT_NODE, nodeName)
    // TODO: what if the src file does not exist?
    if req.Get(HEADER_DISTABLE_PARALLEL)=="TRUE" {
        err=fs.Rm(trimer, nodeName)
    } else {
        err=fs.RmParalleled(trimer, nodeName)
    }
    if !egg.Nil(err) {
        if egg.In(err, exception.EX_INODE_NONEXIST) {
            res.Status("Nonexist container or path.", 404)
            return
        }
        if egg.In(err, exception.EX_FILE_NOT_EXIST) {
            res.Status("Nonexist container or path.", 404)
            return
        }
        res.Status("Internal Error: "+err.Error(), 500)
        return
    }

    //=============================================
    if startTime>0 {
        res.Set("Time-Consumed", strconv.FormatInt(time.Now().UnixNano()-startTime, 10))
    }
    //=============================================
    res.SendCode(204)
}

// ==========================API DOCS=======================================
// API Name: Move one directory
// Action: move the directory only if it exists and its parent path exists
// API URL: /fs/{contianer}/{followingpath}
// REQUEST: PATCH
// Parameters:
//      - contianer(in URL): the container name
//      - followingpath(in URL): the path to be removed. Please guarantee its parent node exists.
//      - Disable-Parallel(in Header): if set to TRUE, a non-parallelized mkdir will be
//              operated. Default to FALSE, it is only for test and not recommend to set.
//      - C-Destination(in Header): the destination path, must be in the same container.
//                               shortcut is allowed.
//      - C-By-Force(in Header): if set to "TRUE", a force move will override any existing
//                               destination file. Default value is FALSE
// Returns:
//      - HTTP 201: The moving has been successfully carried.
//      - HTTP 202: Only returns when C-By-Force is not TRUE and the destination
//                  has already existed.
//      - HTTP 403: Destination should be specified in the header.
//      - HTTP 404: the file or direcory does not exist.
//      - HTTP 500: Internal error. The body is supposed to return error info.
// ==========================API DOCS END===================================
const HEADER_DESTINATION="C-Destination"
const HEADER_MOVE_BY_FORCE="C-By-Force"
func mvDirectory(req Request, res Response) {
    //===============measurement========================
    var startTime int64=0
    if req.Get("Enable-Measure")=="True" {
        startTime=time.Now().UnixNano()
    }
    //==================================================

    var pathDetail, _=req.F()["HandledRR"].([]string)
    if pathDetail==nil {
        pathDetail=req.F()["RR"].([]string)
        pathDetail=append(pathDetail, filesystem.ROOT_INODE_NAME)
    }

    var base, filename=pathman.SplitPath(pathDetail[2])
    if filename=="" {
        res.Status("The directory/file to move should be specified.", 404)
        return
    }
    var destinationALL=req.Get(HEADER_DESTINATION)
    if destinationALL=="" {
        res.Status("Destination path should be specified in the Header "+HEADER_DESTINATION, 403)
        return
    }
    var destinationSC, destinationPath=pathman.ShortcutResolver(destinationALL)
    if destinationSC=="" {
        destinationSC=filesystem.ROOT_INODE_NAME
    } else {
        if len(destinationPath)>0 {
            destinationPath=destinationPath[1:]
        }
    }
    var desBase, desFilename=pathman.SplitPath(destinationPath)
    if desFilename=="" {
        res.Status("The destination directory/file should be specified.", 404)
        return
    }

    var fs=filesystem.GetFs(outapi.NewSwiftio(outapi.DefaultConnector, pathDetail[1]))
    if fs==nil {
        res.Status("Internal Error: the FS pool is full.", 500)
    }
    defer fs.Release()

    var srcNodeNames, desNodeNames string
    var err error

    if req.Get(HEADER_DISTABLE_PARALLEL)=="TRUE" {
        srcNodeNames, err=fs.Locate(base, pathDetail[3])
        if err!=nil {
            res.Status("Nonexist container or path. "+err.Error(), 404)
            return
        }
        desNodeNames, err=fs.Locate(desBase, destinationSC)
    } else {
        var wg sync.WaitGroup
        var lock sync.Mutex
        wg.Add(2)
        go (func() {
            defer wg.Done()

            var err2 error
            srcNodeNames, err2=fs.Locate(base, pathDetail[3])
            if err2!=nil {
                lock.Lock()
                if err==nil {
                    err=err2
                }
                lock.Unlock()
            }
        })()
        go (func() {
            defer wg.Done()

            var err2 error
            desNodeNames, err=fs.Locate(desBase, destinationSC)
            if err2!=nil {
                lock.Lock()
                if err==nil {
                    err=err2
                }
                lock.Unlock()
            }
        })()
        wg.Wait()
    }
    if err!=nil {
        res.Status("Nonexist container or path. "+err.Error(), 404)
        return
    }

    var byForce=req.Get(HEADER_MOVE_BY_FORCE)=="TRUE"
    if req.Get(HEADER_DISTABLE_PARALLEL)=="TRUE" {
        err=fs.MvX(filename, srcNodeNames, desFilename, desNodeNames, byForce)
    } else {
        err=fs.MvXParalleled(filename, srcNodeNames, desFilename, desNodeNames, byForce)
    }
    if !egg.Nil(err) {
        if egg.In(err, exception.EX_FILE_NOT_EXIST) {
            res.Status("Not Found", 404)
            return
        }
        if egg.In(err, exception.EX_FOLDER_ALREADY_EXIST) {
            res.Status("The destination has already existed.", 202)
            return
        }
        res.Status("Internal error: "+err.Error(), 500)
        return
    }

    //=============================================
    if startTime>0 {
        res.Set("Time-Consumed", strconv.FormatInt(time.Now().UnixNano()-startTime, 10))
    }
    //=============================================
    res.SendCode(201)
}
