package uniqueid

import (
    "testing"
)

func TestGenGlobalUniqueName(t *testing.T) {
    t.Log(GenGlobalUniqueName())
    t.Log(GenGlobalUniqueName())
    t.Log(GenGlobalUniqueName())
    t.Log(GenGlobalUniqueName())
    t.Log(GenGlobalUniqueName())
}
