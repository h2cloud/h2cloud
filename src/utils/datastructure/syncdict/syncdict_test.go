// Unit test for splittree

package syncdict

import (
    "testing"
)

func TestSyncdictAllFuncs(t *testing.T) {
	var f=NewSyncdict()
    f.Declare("x","huahua")
    f.Declare("x","asdua")
    f.Set("x",1.23)
    t.Log(f.Get("x"))
}
