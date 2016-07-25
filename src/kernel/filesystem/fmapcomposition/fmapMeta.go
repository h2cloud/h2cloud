package fmapcomposition

/*
** This package is used for defining the meta involved presenting in folder map.
** All the value is presented as a stringified JSON in the value field in kvmap
** file, and identical to the foldermap itself, asychronously updated.
** Attentez: all the meta field is in lower case.
**
** TYPE: DIR/FILE
** For file, any meta of the file may be included.
*/

import (
    "encoding/json"
    . "logger"
)

const (
    FMAP_META_TYPE="type"
    FMAP_META_TYPE_DIRECTORY="dir"
    FMAP_META_TYPE_FILE="file"
)

type FMapMeta map[string]string

func Stringify(obj FMapMeta) string {
    if obj==nil {
        return ""
    }
    var result, err=json.Marshal(obj)
    if err!=nil {
        Secretary.Error("fmapcomposition::Stringify()", "Failed to stringify a map[string]string")
        Insider.Log("fmapcomposition::Stringify()", "Failed to stringify a map[string]string")
        return ""
    }

    return string(result)
}

func Parse(src string) (FMapMeta, error) {
    var ret map[string]string
    var err=json.Unmarshal([]byte(src), &ret)
    if err!=nil {
        Secretary.Error("fmapcomposition::Parse()", "Failed to parse JSON: "+err.Error())
        return nil, err
    }
	return ret, err
}

func (this FMapMeta)Stringify() string {
    return Stringify(this)
}
