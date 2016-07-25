package logger

import (
    . "definition"
)

var Secretary Logger=&consoleLogger{
    doLog: true,
    doWarn: true,
    doErr: true,

    next: nil,
}

type Logger interface {
    LogD(c Tout)
    WarnD(c Tout)
    ErrorD(c Tout)
    Log(pos string, c Tout)
    Warn(pos string, c Tout)
    Error(pos string, c Tout)

    // 000 stands for log, warn, error
    // the larger, the more verbose
    SetLevel(level int)

    // chain anther logger to the last pos of the chain
    Chain(obj Logger) Logger
}
