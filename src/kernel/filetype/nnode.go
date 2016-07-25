package filetype

import (
    "io/ioutil"
    "io"
)

type Nnode struct {
    DesName string
}

func NewNnode(desname string) *Nnode {
    return &Nnode {
        DesName: desname,
    }
}
func (this *Nnode)LoadIn(dtSource io.Reader) error {
    var res, err=ioutil.ReadAll(dtSource)
    if err!=nil {
        return err
    }
    this.DesName=string(res)
    return nil
}
func (this *Nnode)GetType() string {
    return "namenode file"
}
func (this *Nnode)WriteBack(dtDes io.Writer) error {
    _, err:=dtDes.Write([]byte(this.DesName))
    return err
}
