// Wrapper for all the filetypes, offering a quick way to setup one class.
// The id of different types is GetType(). Keep it unique.
package filetype

import (
    "reflect"
    //"fmt"
)
//============================================================
// Modify this to add new filetype.
var prototypeList=[]Filetype{&Kvmap{}, &Nnode{}}
//============================================================

var typeMap=makeTypeMap()

func makeTypeMap() map[string]reflect.Type {
    ret:=make(map[string]reflect.Type)
    for _, elem:=range prototypeList {
        ret[elem.GetType()]=reflect.TypeOf(elem).Elem()
    }
    return ret
}

func Makefile(typeid string) Filetype {
    if res, ok:=typeMap[typeid]; ok {
        return reflect.New(res).Interface().(Filetype)
    }
    return nil
}
