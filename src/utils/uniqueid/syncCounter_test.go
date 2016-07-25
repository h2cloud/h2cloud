package uniqueid

import (
    "testing"
)

func TestSyncdictAllFuncs(t *testing.T) {
	var f=SyncCounter{}
    t.Log(f.Get())
    t.Log(f.Inc())
    t.Log(f.Inc())
    t.Log(f.Dec())
    f.Set(-5)
    t.Log(f.Inc())
}
