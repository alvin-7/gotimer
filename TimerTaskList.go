package timer

import (
	"sync"
	"sync/atomic"
	"time"
)

type TimerTaskList struct {
	root        *TimerTaskEntry
	expiration  *int64
	taskCounter *int32
	locker      sync.Mutex
	rLocker     sync.Mutex
}

func NewTimerTaskList(taskCounter *int32) *TimerTaskList {
	lis := new(TimerTaskList)
	root := NewTimerTaskEntry(nil, -1)
	root.prev = root
	root.next = root
	root.list = lis
	expiration := int64(-1)

	lis.root = root
	lis.expiration = &expiration
	lis.taskCounter = taskCounter
	return lis
}

func (l *TimerTaskList) setExpiration(expirationMs int64) bool {
	return atomic.SwapInt64(l.expiration, expirationMs) != expirationMs
}

func (l *TimerTaskList) getExpiration() int64 {
	return *l.expiration
}

func (l *TimerTaskList) foreach(f func(*TimerTask)) {
	l.locker.Lock()
	defer l.locker.Unlock()
	entry := l.root.next
	for entry != l.root {
		nextEntry := entry.next
		if !entry.cancelled() {
			f(entry.timerTask)
		}
		entry = nextEntry
	}
}

func (l *TimerTaskList) add(entry *TimerTaskEntry) {
	var done = false
	for !done {
		entry.remove()
		l.locker.Lock()
		defer l.locker.Unlock()
		{
			entry.locker.Lock()
			defer entry.locker.Unlock()
			if entry.list == nil {
				tail := l.root.prev
				entry.next = l.root
				entry.prev = tail
				entry.list = l
				tail.next = entry
				l.root.prev = entry
				atomic.AddInt32(l.taskCounter, 1)
				done = true
			}
		}
	}
}

func (l *TimerTaskList) remove(entry *TimerTaskEntry) {
	l.rLocker.Lock()
	defer l.rLocker.Unlock()

	if entry.list == l {
		entry.next.prev = entry.prev
		entry.prev.next = entry.next
		entry.next = nil
		entry.prev = nil
		entry.list = nil
		atomic.AddInt32(l.taskCounter, -1)
	}
}

func (l *TimerTaskList) flush(f func(*TimerTaskEntry)) {
	l.locker.Lock()
	defer l.locker.Unlock()
	head := l.root.next
	for head != l.root {
		l.remove(head)
		f(head)
		head = l.root.next
	}
	atomic.StoreInt64(l.expiration, -1)
}

func (l *TimerTaskList) getDelay() time.Duration {
	delay := l.getExpiration() - time.Now().UnixMilli()
	if delay > 0 {
		return time.Duration(delay) * time.Millisecond
	} else {
		return 0
	}
}

func (l *TimerTaskList) compareTo(list *TimerTaskList) bool {
	return l.getExpiration() < list.getExpiration()
}
