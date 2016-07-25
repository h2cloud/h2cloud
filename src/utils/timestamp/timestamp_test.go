// Unit test for splittree

package timestamp

import (
    "testing"
)

func TestSyncdictAllFuncs(t *testing.T) {
    t.Log(GetTimestamp(0))
    t.Log(GetVersionNumber(GetTimestamp(0)))
    t.Log(GetTimestamp(GetTimestamp(0)))
}
