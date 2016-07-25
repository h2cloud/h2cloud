package logger

import (
    . "definition"
)

type voidDebugger struct {
}

func (_ *voidDebugger)LogD(c Tout) {
}
func (_ *voidDebugger)Log(pos string, c Tout) {
}
