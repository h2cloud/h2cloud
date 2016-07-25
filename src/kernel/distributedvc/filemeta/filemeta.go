package filemeta

import (
    . "kernel/distributedvc/constdef"
    "strings"
)

type FileMeta map[string]string

func NewMeta() FileMeta {
    return FileMeta(map[string]string{})
}

func CheckIntegrity(obj FileMeta) bool {
    if _, ok:=obj[METAKEY_TYPE]; !ok {
        return false
    }
    return true
}
func (this FileMeta)ToUserMeta() UserMeta {
    return FileMeta2UserMeta(this)
}

func FileMeta2UserMeta(fm FileMeta) UserMeta {
    var ret=UserMeta(map[string]string{})
    for k, v:=range fm {
        if strings.HasPrefix(k, USER_META_HEADER) {
            ret[k[HEADER_LENGTH:]]=v
        }
    }
    return ret
}

func (this FileMeta)Clone() FileMeta {
    var ret=NewMeta()
    for k, v:=range this {
        ret[k]=v
    }

    return ret
}
