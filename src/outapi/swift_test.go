package outapi

import (
    "testing"
    "kernel/filetype"
    "fmt"
)

func TestPut(t *testing.T) {
    // May returns 404 for unexist container
    var io=NewSwiftio(DefaultConnector, "1@levy.at")
    fmt.Println(io.Put("1234", filetype.FastMake("12"), nil))
}
