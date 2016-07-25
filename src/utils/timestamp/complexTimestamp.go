// OLD version is deprecated. It is currently only serve for absTimestamp

package timestamp

import (
    "time"
    "strconv"
)

type ClxTimestamp uint64


func processRawTime(unixTime uint64) ClxTimestamp {
    unixTime=unixTime&0xfffffffff
    unixTime=0xfffffffff-unixTime
    return ClxTimestamp(unixTime)
}

func GetTimestamp() ClxTimestamp {
    return ClxTimestamp(time.Now().UnixNano())
}

func String2ClxTimestamp(val string) ClxTimestamp {
    res, err:=strconv.ParseUint(val, 10, 64)
    if err!=nil {
        return 0
    }
    return ClxTimestamp(res)
}
func ClxTimestamp2String(val ClxTimestamp) string {
    return strconv.FormatUint(uint64(val), 10)
}

// identical to ClxTimestamp2String
func (this ClxTimestamp)String() string {
    return ClxTimestamp2String(this)
}
func (this ClxTimestamp)Val() uint64 {
    return uint64(this)
}

func MergeTimestamp(ts1, ts2 ClxTimestamp) ClxTimestamp {
    if ts1>ts2 {
        return ts1
    } else {
        return ts2
    }
}
