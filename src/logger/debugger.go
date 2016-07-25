package logger

import (
    . "definition"
)

var Insider Dubugger=&consoleDebugger{}
//var Insider Dubugger=&voidDebugger{}

type Dubugger interface {
    LogD(c Tout)
    Log(pos string, c Tout)
}
