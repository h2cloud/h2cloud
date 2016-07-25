package interactive

import (
    . "utils/timestamp"
    "strconv"
    . "definition"
    "strings"
    "errors"
)

const OUTAPI_PLACEHOLDER_PING_FLAG="$SYS.PING"

// implementation for the format to gossip
// if OutAPI==OUTAPI_PLACEHOLDER_PING_FLAG, the entry is a sync heartbeat, in which
// the UpdateTime indicate the sending time, NodeNumber indicates the initial node and,
// Filename is the senders API address(NOT IMPLEMENTED YET)

type GossipEntry struct {
    Filename string
    OutAPI string
    UpdateTime ClxTimestamp
    NodeNumber int
}

// Just segment them with \n
func (this *GossipEntry)Stringify() string {
    return  this.Filename+"\n"+
            this.OutAPI+"\n"+
            this.UpdateTime.String()+"\n"+
            strconv.Itoa(this.NodeNumber)
}

func BatchStringify(src []Tout) (string, error) {
    var ret=""
    for i, e:=range src {
        if p, ok:=e.(*GossipEntry); !ok {
            return "", errors.New("Format error")
        } else {
            ret+=p.Stringify()
            if i!=len(src)-1 {
                ret+="\n"
            }
        }
    }

    return ret, nil
}

// For errors returns nil
func ParseOne(src string) *GossipEntry {
    var res=strings.SplitN(src, "\n", 4)
    if len(res)!=4 {
        return nil
    }
    var pInt, err=strconv.Atoi(res[3])
    if err!=nil {
        return nil
    }

    return &GossipEntry {
        Filename: res[0],
        OutAPI: res[1],
        UpdateTime: String2ClxTimestamp(res[2]),
        NodeNumber: pInt,
    }
}

// For errors returns nil
func ParseAll(src string) []*GossipEntry {
    var res=strings.Split(src, "\n")
    if len(res)%4!=0 {
        return nil
    }
    var result=[]*GossipEntry{}
    for i:=0; i<len(res); i+=4 {
        var pInt, err=strconv.Atoi(res[i+3])
        if err!=nil {
            return nil
        }
        result=append(result, &GossipEntry {
            Filename: res[i],
            OutAPI: res[i+1],
            UpdateTime: String2ClxTimestamp(res[i+2]),
            NodeNumber: pInt,
        })
    }

    return result
}
