// Define the output interface to contect with Openstack Swift.
// Also, it can be rewrite to connect to local disk or other storage media.
// Pay attention that the input/output data is assembled, for streaming version
// Please refer to put/getStream

package outapi

import (
    "kernel/filetype"
    . "kernel/distributedvc/filemeta"
    "io"
)

type Outapi interface {

    GenerateUniqueID() string

    // Recognize the unique id of this kind. If match, generate one instance.
    // else, returns nil
    // @ Static
    RecognizeSelf(name string) Outapi

    // Need not have typestamp in FileMeta. It will be set according to content's record automatically.
    // Filemeta could be nil.
    Put(filename string, content filetype.Filetype, info FileMeta) error

    // If file does not exist, a nil will be returned. No error occurs.
    Get(filename string) (FileMeta, filetype.Filetype, error)
    // Differs from Get() that it returns the whole HTTP header, not only the
    // object meta. ALSO, all the keys are in lower case.
    GetX(filename string) (map[string]string, filetype.Filetype, error)

    Putinfo(filename string, info FileMeta) error

    // If file does not exist, a nil will be returned. No error occurs.
    Getinfo(filename string) (FileMeta, error)
    // Differs from Getinfo() that it returns the whole HTTP header, not only the
    // object meta. ALSO, all the keys are in lower case.
    GetinfoX(filename string) (map[string]string, error)

    Delete(filename string) error

    // If file does not exist, a nil will be returned. No error occurs.
    // Pay attention that io.ReadCloser should be closed.
    GetStream(filename string) (FileMeta, io.ReadCloser, error)
    GetStreamX(filename string) (map[string]string, io.ReadCloser, error)

    PutStream(filename string, info FileMeta) (io.WriteCloser, error)

    // If the space is not available, create it and return (TRUE, nil);
    // If the space is already available, return (FALSE, nil);
    // Otherwise, (space is not available and fail to create), return a non-nil error.
    EnsureSpace() (bool, error)

    // Copy a file on the server side. With Keys set in FileMeta overriding the
    // original ones.
    // If the file does not exist, an error will be returned.
    Copy(srcname string, desname string, overrideMeta FileMeta) error

    // if error!=nil, bool is always false
    CheckExist(filename string) (bool, error)

    ExtractFileMeta(src map[string]string) FileMeta

}

func ForceCheckExist(ex bool, err error) bool {
    if err!=nil {
        return false
    }
    return ex
}

var enumTypes=[]Outapi{&Swiftio{}, &EmptyIO{}}
func DeSerializeID(name string) Outapi {
    for _, e:=range enumTypes {
        if t:=e.RecognizeSelf(name); t!=nil {
            return t
        }
    }
    return nil
}
