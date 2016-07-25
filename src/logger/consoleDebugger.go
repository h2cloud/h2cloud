package logger

import (
    "fmt"
    . "definition"
    "time"
)

type consoleDebugger struct {
}

func (_ *consoleDebugger)LogD(c Tout) {
    fmt.Println("#Debugger  #", time.Now().Format(time.StampMilli)+"#", c)
}
func (_ *consoleDebugger)Log(pos string, c Tout) {
    fmt.Println("#Debugger  #", time.Now().Format(time.StampMilli)+", "+pos+"#", c)
}
