package filemeta

// Used for user-specified metadata. Add a non-replacable prefix.
// The conversion is implemented in InAPI

type UserMeta map[string]string

const USER_META_HEADER="usermeta-"
const HEADER_LENGTH=len(USER_META_HEADER)

func (this UserMeta)ToFileMeta() FileMeta {
    return UserMeta2FileMeta(this)
}

func UserMeta2FileMeta(um UserMeta) FileMeta {
    var ret=FileMeta(map[string]string{})
    for k, v:=range um {
        ret[USER_META_HEADER+k]=v
    }
    return ret
}
