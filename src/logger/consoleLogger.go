package logger

import (
    "fmt"
    . "definition"
    "strconv"
    "time"
)

type consoleLogger struct {
    doLog bool
    doWarn bool
    doErr bool

    next Logger
}

func (this *consoleLogger)LogD(c Tout) {
    if this.doLog {
        fmt.Println("<Log  @", time.Now().Format(time.RFC3339)+">", c)
    }
    if this.next!=nil {
        this.next.LogD(c)
    }
}
func (this *consoleLogger)WarnD(c Tout) {
    if this.doWarn {
        fmt.Println("<Warn @", time.Now().Format(time.RFC3339)+">", c)
    }
    if this.next!=nil {
        this.next.WarnD(c)
    }
}
func (this *consoleLogger)ErrorD(c Tout) {
    if this.doErr {
        fmt.Println("<Err  @", time.Now().Format(time.RFC3339)+">", c)
    }
    if this.next!=nil {
        this.next.ErrorD(c)
    }
}
func (this *consoleLogger)Log(pos string, c Tout) {
    if this.doLog {
        fmt.Println("<Log  @", time.Now().Format(time.RFC3339)+", "+pos+">", c)
    }
    if this.next!=nil {
        this.next.Log(pos, c)
    }
}
func (this *consoleLogger)Warn(pos string, c Tout) {
    if this.doWarn {
        fmt.Println("<Warn @", time.Now().Format(time.RFC3339)+", "+pos+">", c)
    }
    if this.next!=nil {
        this.next.Warn(pos, c)
    }
}
func (this *consoleLogger)Error(pos string, c Tout) {
    if this.doErr {
        fmt.Println("<Err  @", time.Now().Format(time.RFC3339)+", "+pos+">", c)
    }
    if this.next!=nil {
        this.next.Error(pos, c)
    }
}

func (this *consoleLogger)SetLevel(level int) {
    this.doErr  =(level & 1!=0)
    this.doWarn =(level & 2!=0)
    this.doLog  =(level & 4!=0)

    Secretary.Log("logger.consoleLogger::SetLevel", "Log Level is set to "+strconv.Itoa(level & 7))

    if this.next!=nil {
        this.next.SetLevel(level)
    }
}
func (this *consoleLogger)Chain(obj Logger) Logger {
    if this.next==nil {
        this.next=obj
    } else {
        this.next.Chain(obj)
    }
    return this
}
