package timer

import (
	"sync"
	"sync/atomic"
	"time"
)

type TimerTaskList struct {
	root        *TimerTaskEntry
	list        *TimerTaskList
	next        *TimerTaskEntry
	prev        *TimerTaskEntry
	expiration  int64
	taskCounter *int32
	locker      sync.Mutex
}

func NewTimerTaskList(taskCounter *int32) *TimerTaskList {
	root := NewTimerTaskEntry(nil, -1)
	root.prev = root
	root.next = root
	expiration := int64(-1)
	return &TimerTaskList{
		root:        root,
		list:        nil,
		prev:        nil,
		expiration:  expiration,
		taskCounter: taskCounter,
	}
}

func (this *TimerTaskList) setExpiration(expirationMs int64) bool {
	return atomic.SwapInt64(&this.expiration, expirationMs) != expirationMs
}

func (this *TimerTaskList) getExpiration() int64 {
	return atomic.LoadInt64(&this.expiration)
}

func (this *TimerTaskList) foreach(f func(*TimerTask)) {
	this.locker.Lock()
	defer this.locker.Unlock()
	entry := this.root.next
	for entry != this.root {
		nextEntry := entry.next
		if !entry.cancelled() {
			f(entry.timerTask)
		}
		entry = nextEntry
	}
}

func (this *TimerTaskList) add(entry *TimerTaskEntry) {
	var done = false
	for !done {
		entry.remove()
		this.locker.Lock()
		defer this.locker.Unlock()
		{
			entry.locker.Lock()
			defer entry.locker.Unlock()
			if entry.list == nil {
				tail := this.root.prev
				entry.next = this.root
				entry.prev = tail
				entry.list = this
				tail.next = entry
				this.root.prev = entry
				atomic.AddInt32(this.taskCounter, 1)
				done = true
			}
		}
	}
}

func (this *TimerTaskList) remove(entry *TimerTaskEntry) {
	this.locker.Lock()
	defer this.locker.Unlock()

	if entry.list == this {
		entry.next.prev = entry.prev
		entry.prev.next = entry.next
		entry.next = nil
		entry.prev = nil
		entry.list = nil
		atomic.AddInt32(this.taskCounter, -1)
	}
}

func (this *TimerTaskList) flush(f func(*TimerTaskEntry)) {
	this.locker.Lock()
	defer this.locker.Unlock()
	head := this.root.next
	for head != this.root {
		this.remove(head)
		f(head)
		head = this.root.next
	}
	atomic.StoreInt64(&this.expiration, -1)
}

func (this *TimerTaskList) getDelay() time.Duration {
	delay := this.getExpiration() - time.Now().UnixMilli()
	if delay > 0 {
		return time.Duration(delay)
	} else {
		return time.Duration(0)
	}
}

func (this *TimerTaskList) compareTo(l *TimerTaskList) bool {
	return this.getExpiration() < l.getExpiration()
}
