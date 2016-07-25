package distributedvc

import (
    "sync"
)

type fdDLinkedListNode struct {
    carrier *FD
    prev *fdDLinkedListNode
    next *fdDLinkedListNode
}

type fdDLinkedList struct {
    Head *fdDLinkedListNode
    Tail *fdDLinkedListNode
    Length int

    Lock *sync.Mutex
}

func genList(head *fdDLinkedListNode) *fdDLinkedListNode {
    var ret=&fdDLinkedListNode {
        carrier: nil,
        next: nil,
        prev: head,
    }
    head.next=ret
    return ret
}
func NewFSDLinkedList() *fdDLinkedList {
    var head=&fdDLinkedListNode {
        carrier: nil,
        prev: nil,
    }
    var tail=genList(head)

    return &fdDLinkedList {
        Head: head,
        Tail: tail,
        Length: 0,
        Lock: &sync.Mutex{},
    }
}

func (this *fdDLinkedList)Append(element *fdDLinkedListNode) {
    this.Lock.Lock()
    defer this.Lock.Unlock()

    element.prev=this.Tail.prev
    element.next=this.Tail
    element.prev.next=element
    element.next.prev=element

    this.Length++
}
func (this *fdDLinkedList)AppendX(carrier *FD) {
    this.Append(&fdDLinkedListNode {
        carrier: carrier,
    })
}
func (this *fdDLinkedList)AppendWithoutLock(element *fdDLinkedListNode) {
    element.prev=this.Tail.prev
    element.next=this.Tail
    element.prev.next=element
    element.next.prev=element

    this.Length++
}
func (this *fdDLinkedList)Cut(element *fdDLinkedListNode) {
    this.Lock.Lock()
    defer this.Lock.Unlock()

    element.prev.next=element.next
    element.next.prev=element.prev
    this.Length--
}
