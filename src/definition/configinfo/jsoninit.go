package configinfo

import (
    "io/ioutil"
    . "definition"
    "encoding/json"
    "bytes"
    _ "fmt"
    "errors"
)

// Removes the /**/ line from the string
func RemoveSlashCommentLine(str []byte) []byte {
    var i=0
    var j, k int
    lim:=len(str)
    res:=[]byte{}
    sep:=[]byte("/*")
    ed:=[]byte("*/")
    for i<lim {
        j=bytes.Index(str[i:], sep)
        if j==-1 {
            res=append(res, str[i:]...)
            break
        }
        res=append(res, str[i:i+j]...)
        k=bytes.Index(str[i+j:], ed)
        //fmt.Println(i,",",j,",",k)
        if k==-1 {
            break
        }
        i=i+j+k+len(ed)
    }

    return res
}

//Pay attention that filename is a relative path
func ReadFileToJSON(filename string) (map[string]Tout, error) {
    var err error
    var res []byte
    filename, err=GetABSPath(filename)
    if err!=nil {
        return nil, err
    }
    res, err=ioutil.ReadFile(filename)
    if err!=nil {
        return nil, err
    }

    //fmt.Println(string(res))
    res=RemoveSlashCommentLine(res)
    //fmt.Println(string(res))

    var ret map[string]Tout
    err=json.Unmarshal(res, &ret)
    if err!=nil {
        return nil, err
    }

    return ret, nil
}

func AppendFileToJSON(filename string, tomodify map[string]Tout) error {
    if tomodify==nil {
        return errors.New("Empty config pattern.")
    }
    var tMap, err=ReadFileToJSON(filename)
    if err!=nil {
        return err
    }
    for k, v:=range tMap {
        if _, ok:=tomodify[k]; !ok {
            tomodify[k]=v
        }
    }
    return nil
}
