package outapi

import (
    "github.com/ncw/swift"
    "definition/configinfo"
    "fmt"
    . "kernel/distributedvc/filemeta"
    "definition/exception"
    "kernel/filetype"
    "bytes"
    "io"
    . "kernel/distributedvc/constdef"
    "strings"
)

type SwiftConnector struct {
    c *swift.Connection
}

func (this *SwiftConnector)DumpConn() *swift.Connection {
    return this.c
}

func _no__use_1_() {
    fmt.Println("nosue")

}
// If auth failed, return nil
func ConnectbyAuth(username string, passwd string, tenant string) *SwiftConnector {
    swc:=&swift.Connection{
        UserName: username,
        ApiKey: passwd,
        Tenant: tenant,
        AuthUrl: configinfo.SWIFT_AUTH_URL,
        //AuthVersion: 2,
    }
    if err:=swc.Authenticate();err!=nil {
        panic(exception.EX_KEYSTONE_AUTH_ERROR)
        return nil
    }

    return &SwiftConnector{
        c: swc,
    }
}
func ConnectbyPreauth(account string, token string) *SwiftConnector {
    // TODO: set it up!
    panic(exception.EX_KEYSTONE_AUTH_ERROR)
    return nil
}

type Swiftio struct {
    //Implementing outapi.Outapi
    conn *SwiftConnector
    container string
}

// 2 ways to setup a new swift io.
func NewSwiftio(_conn *SwiftConnector, _container string) *Swiftio {
    return &Swiftio{
        conn: _conn,
        container: _container,
    }
}
func DupSwiftio(oldio *Swiftio, _container string) *Swiftio {
    return &Swiftio{
        conn: oldio.conn,
        container: _container,
    }
}

const (
    swift_pfx="outapi.Swiftio: "
)
func (this *Swiftio)GenerateUniqueID() string {
    return swift_pfx+this.container
}
func (_ *Swiftio)RecognizeSelf(name string) Outapi {
    if strings.HasPrefix(name, swift_pfx) {
        // Attentez: return one using DefaultConnector
        return NewSwiftio(DefaultConnector, name[len(swift_pfx):])
    } else {
        return nil
    }
}

func (this *Swiftio)Getinfo(filename string) (FileMeta, error) {
    _, headers, err:=this.conn.c.Object(this.container, filename)
    if err!=nil {
        if err==swift.ObjectNotFound {
            return nil, nil
        }
        return nil, err
    }
    //fmt.Println(headers)
    return FileMeta(headers.ObjectMetadata()), nil
}
func (this *Swiftio)GetinfoX(filename string) (map[string]string, error) {
    _, headers, err:=this.conn.c.Object(this.container, filename)
    if err!=nil {
        if err==swift.ObjectNotFound {
            return nil, nil
        }
        return nil, err
    }
    //fmt.Println(headers)
    return convertToLowerCaseMap(map[string]string(headers)), nil
}

func (this *Swiftio)Putinfo(filename string, info FileMeta) error {
    head4Put:=swift.Metadata(info).ObjectHeaders()
    return this.conn.c.ObjectUpdate(this.container, filename, head4Put)
}

func (this *Swiftio)Delete(filename string) error {
    err:=this.conn.c.ObjectDelete(this.container, filename)
    if err!=nil && err!=swift.ObjectNotFound {
        return err
    }
    return nil
}

func convertToLowerCaseMap(src map[string]string) map[string]string {
    if src==nil {
        return nil
    }
    var ret=make(map[string]string)
    for k, v:=range src {
        ret[strings.ToLower(k)]=v
    }

    return ret
}

// Get file and automatically check the MD5
func (this *Swiftio)Get(filename string) (FileMeta, filetype.Filetype, error) {
    contents:=&bytes.Buffer{}
    header, err:=this.conn.c.ObjectGet(
        this.container, filename, contents,
        configinfo.INDEX_FILE_CHECK_MD5,
        nil)

    if err!=nil {
        if err==swift.ObjectNotFound {
            return nil, nil, nil
        }
        return nil, nil, err
    }
    meta:=header.ObjectMetadata()

    resFile:=filetype.Makefile(meta[METAKEY_TYPE])
    if resFile==nil {
        return nil, nil, exception.EX_UNSUPPORTED_TYPESTAMP
    }
    resFile.LoadIn(contents)

    return FileMeta(meta), resFile, nil
}

func (this *Swiftio)GetX(filename string) (map[string]string, filetype.Filetype, error) {
    contents:=&bytes.Buffer{}
    header, err:=this.conn.c.ObjectGet(
        this.container, filename, contents,
        configinfo.INDEX_FILE_CHECK_MD5,
        nil)

    if err!=nil {
        if err==swift.ObjectNotFound {
            return nil, nil, nil
        }
        return nil, nil, err
    }
    meta:=header.ObjectMetadata()

    resFile:=filetype.Makefile(meta[METAKEY_TYPE])
    if resFile==nil {
        return nil, nil, exception.EX_UNSUPPORTED_TYPESTAMP
    }
    resFile.LoadIn(contents)

    return convertToLowerCaseMap(map[string]string(header)), resFile, nil
}

func (this *Swiftio)Put(filename string, content filetype.Filetype, info FileMeta) error {
    if info==nil {
        info=FileMeta(map[string]string{})
    }
    meta:=swift.Metadata(info.Clone())
    meta[METAKEY_TYPE]=content.GetType()

    buffer:=&bytes.Buffer{}
    content.WriteBack(buffer)

    _, err:=this.conn.c.ObjectPut(this.container, filename, buffer, false, "", "", meta.ObjectHeaders())
    return err
}

func (this *Swiftio)GetStream(filename string) (FileMeta, io.ReadCloser, error) {
    file, header, err:=this.conn.c.ObjectOpen(this.container, filename, false, nil)
    if err!=nil {
        if err==swift.ObjectNotFound {
            return nil, nil, nil
        }
        return nil, nil, err
    }
    meta:=header.ObjectMetadata()

    return FileMeta(meta), file, nil
}
func (this *Swiftio)GetStreamX(filename string) (map[string]string, io.ReadCloser, error) {
    file, header, err:=this.conn.c.ObjectOpen(this.container, filename, false, nil)
    if err!=nil {
        if err==swift.ObjectNotFound {
            return nil, nil, nil
        }
        return nil, nil, err
    }

    return map[string]string(header), file, nil
}

func (this *Swiftio)PutStream(filename string, info FileMeta) (io.WriteCloser, error) {
    if info==nil || !CheckIntegrity(info) {
        return nil, exception.EX_METADATA_NEEDS_TO_BE_SPECIFIED
    }
    meta:=swift.Metadata(info)

    fileW, err:=this.conn.c.ObjectCreate(this.container, filename, false, "", "", meta.ObjectHeaders())
    if err!=nil {
        return nil, err
    }
    return fileW, nil
}

func (this *Swiftio)EnsureSpace() (bool, error) {
    var err=this.conn.c.ContainerCreateX(this.container, nil)
    if err==nil {
        return true, nil
    }
    if err==swift.AlreadyExist {
        return false, nil
    }
    return false, err
    /*
    _, _, err:=this.conn.c.Container(this.container)
    if err==swift.ContainerNotFound {
        err=this.conn.c.ContainerCreate(this.container, nil)
        return true, err
    }
    return false, err
    */
}

func (this *Swiftio)test_Container() (bool, error) {
    var err=this.conn.c.ContainerCreateX(this.container, nil)
    return true, err
}

func (this *Swiftio)Copy(srcname string, desname string, overrideMeta FileMeta) error {
    var _, err=this.conn.c.ObjectCopy(this.container, srcname, this.container, desname, swift.Metadata(overrideMeta).ObjectHeaders())
    return err
}

func (this *Swiftio)CheckExist(filename string) (bool, error) {
    _, _, err:=this.conn.c.Object(this.container, filename)
    if err!=nil {
        if err==swift.ObjectNotFound {
            return false, nil
        }
        return false, err
    }
    return true, nil
}

func (this *Swiftio)ExtractFileMeta(src map[string]string) FileMeta {
    return FileMeta(swift.Headers(src).ObjectMetadata())
}
