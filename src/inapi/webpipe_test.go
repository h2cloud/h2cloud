package inapi

import (
    "net/http"
    "testing"
    "io"
)

func TestMakePipedRequests(t *testing.T) {
    r, w:=io.Pipe()
    req, _:=http.NewRequest("PUT", "http://127.0.0.1:9144/upload", r)
    go func() {
        w.Write([]byte("Huahua aichi baomihua!"))
        w.Close()
    } ()
    _, err:=(&http.Client{}).Do(req)
    t.Log(err)

    return
}
