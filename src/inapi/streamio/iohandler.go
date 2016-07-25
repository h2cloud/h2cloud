package streamio

// the router implementing restAPIs that support streaming upload/download

import (
    . "github.com/levythu/gurgling"
    "outapi"
    "kernel/filesystem"
    //"kernel/filetype"
    "definition/exception"
    "io"
    "fmt"
    "utils/pathman"
    egg "definition/errorgroup"
    //"logger"
)

func __iohandlergo_nouse() {
    fmt.Println("nouse")
}

func IORouter() Router {
    var rootRouter=ARegexpRouter()
    rootRouter.Use(`/([^/]+)/\[\[SC\](.+)\]/?(.*)`, handlingShortcut)

    rootRouter.Get(`/([^/]+)/(.*)`, downloader)
    rootRouter.Put(`/([^/]+)/(.*)`, uploader)

    return rootRouter
}

// Handling shortcut retrieve. It's applied to all the api in the field
// format: /fs/{contianer}/[[SC]{rootnode}]/{followingpath}

// Note: In Shortcut Access([SC[inode]]/...), the trailing path can be empty, so that
// the operation will be directly conducted on the inode itself.
// If the inode is a folder and a PUT is conducted, the folder will be3 both a
// directory and a file
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

const PARENT_NODE="Parent-Node"
const FILE_NODE="File-Node"

// ==========================API DOCS=======================================
// API Name: Stream data from specified path
// Action: Read the destination data and return it as a file by streaming
// API URL: /io/{contianer}/{followingpath}
// REQUEST: GET
// Parameters:
//      - contianer(in URL): the container name
//      - followingpath(in URL): the path to be listed
// Returns:
//      - HTTP 200: No error and the result will be returned in raw-data streaming.
//                  if successfully, the returned header Parent-Node(if accessed) will
//                  contain its parent inode and File-Node will indicate the file itself.
//      - HTTP 404: Either the container or the filepath does not exist.
//      - HTTP 500: Error. The body is supposed to return error info.
// ==========================API DOCS END===================================

const HEADER_CONTENT_DISPOSE="Content-Disposition"
const ORIGINAL_HEADER="Ori-"
func downloader(req Request, res Response) {
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

    var hasSent bool=false
    if base, filename:=pathman.SplitPath(pathDetail[2]); filename=="" {
        var err=fs.Get("", pathDetail[3], func(fileInode string, oriName string, oriHeader map[string]string) io.Writer {
            for k, v:=range oriHeader {
                res.Set(ORIGINAL_HEADER+k, v)
            }
            res.Set(FILE_NODE, fileInode)
            res.Set(HEADER_CONTENT_DISPOSE, "inline; filename=\""+oriName+"\"")
            res.SendCode(200)
            hasSent=true
            return res.R()
        })
        if err!=nil && !hasSent {
            if err==exception.EX_FILE_NOT_EXIST {
                res.Status("File Not Found.", 404)
            } else {
                res.Status("Internal Error: "+err.Error(), 500)
            }
        }
    } else {
        var nodeName, err=fs.Locate(base, pathDetail[3])
        if err!=nil {
            res.Status("Nonexist container or path. "+err.Error(), 404)
            return
        }
        err=fs.Get(filename, nodeName, func(fileInode string, oriName string, oriHeader map[string]string) io.Writer {
            for k, v:=range oriHeader {
                res.Set(ORIGINAL_HEADER+k, v)
            }
            res.Set(PARENT_NODE, nodeName)
            res.Set(HEADER_CONTENT_DISPOSE, "inline; filename=\""+oriName+"\"")
            res.Set(FILE_NODE, fileInode)
            res.SendCode(200)
            hasSent=true
            return res.R()
        })
        if err!=nil && !hasSent {
            if err==exception.EX_FILE_NOT_EXIST {
                res.Status("File Not Found.", 404)
            } else {
                res.Status("Internal Error: "+err.Error(), 500)
            }
        }
    }
}

// ==========================API DOCS=======================================
// API Name: Stream data to specified file
// Action: Write the data from http to the file streamingly
// API URL: /io/{contianer}/{followingpath}
// REQUEST: PUT
// Parameters:
//      - contianer(in URL): the container name
// Returns:
//      - HTTP 200: No error, the file is written by force
//              When success, the returned header Parent-Node(if accessed) will
//              contain its parent inode and File-Node will indicate the file itself.
//      - HTTP 404: Either the container or the filepath does not exist.
//      - HTTP 500: Error. The body is supposed to return error info.
// ==========================API DOCS END===================================
func uploader(req Request, res Response) {
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

    var putErr error
    if base, filename:=pathman.SplitPath(pathDetail[2]); filename=="" {
        // TODO: glean user meta
        putErr, _=fs.Put("", pathDetail[3], nil, req.R().Body)
        res.Set(FILE_NODE, pathDetail[3])
    } else {
        var nodeName, err=fs.Locate(base, pathDetail[3])
        if err!=nil {
            res.Status("Nonexist container or path. "+err.Error(), 404)
            return
        }
        // TODO: glean user meta
        var targetNode string
        putErr, targetNode=fs.Put(filename, nodeName, nil, req.R().Body)
        if targetNode!="" {
            res.Set(FILE_NODE, targetNode)
        }
        res.Set(PARENT_NODE, nodeName)
    }

    if  !egg.Nil(putErr) {
        if egg.In(putErr, exception.EX_FILE_NOT_EXIST) {
            res.Status("Nonexist container or path. Or you cannot refer to a non-existing inode in ovveride mode.", 404)
            return
        }
        res.Status("Internal Error: "+putErr.Error(), 500)
        return
    }
    res.SendCode(200)
}
