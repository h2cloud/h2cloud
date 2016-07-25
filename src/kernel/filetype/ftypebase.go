package filetype

import (
    "io"
)

type Filetype interface {
    LoadIn(dtSource io.Reader) error
    WriteBack(dtDes io.Writer) error

    GetType() string
}
