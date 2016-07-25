package timestamp

import (
    "time"
    "strconv"
)

func GetABSTimestamp() uint64 {
    return uint64(time.Now().UnixNano())
}

func ABSTimestamp2String(val uint64) string {
    return strconv.FormatUint(val, 10)
}

func String2ABSTimestamp(val string) uint64 {
    res, err:=strconv.ParseUint(val, 10, 64)
    if err!=nil {
        return 0
    }
    return res
}
