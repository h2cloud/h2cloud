package logger

import (
    . "definition"
)

type voidLogger struct {
    next Logger
}

func (this *voidLogger)LogD(c Tout) {
    if this.next!=nil {
        this.next.LogD(c)
    }
}
func (this *voidLogger)WarnD(c Tout) {
    if this.next!=nil {
        this.next.WarnD(c)
    }
}
func (this *voidLogger)ErrorD(c Tout) {
    if this.next!=nil {
        this.next.ErrorD(c)
    }
}
func (this *voidLogger)Log(pos string, c Tout) {
    if this.next!=nil {
        this.next.Log(pos, c)
    }
}
func (this *voidLogger)Warn(pos string, c Tout) {
    if this.next!=nil {
        this.next.Warn(pos, c)
    }
}
func (this *voidLogger)Error(pos string, c Tout) {
    if this.next!=nil {
        this.next.Error(pos, c)
    }
}

func (this *voidLogger)SetLevel(level int) {
    if this.next!=nil {
        this.next.SetLevel(level)
    }
}

func (this *voidLogger)Chain(obj Logger) Logger {
    if this.next==nil {
        this.next=obj
    } else {
        this.next.Chain(obj)
    }
    return this
}
