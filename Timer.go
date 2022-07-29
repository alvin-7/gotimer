package timer

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Timer interface {
	add()
	addTimerTaskEntry()
	size()
	shutdown()
}

type SystemTimer struct {
	delayQueue  *DelayQueue
	taskCounter *int32
	timingWheel *TimingWheel
	locker      sync.RWMutex
	group       sync.WaitGroup
}

func NewSystemTimer(tickMs int64, wheelSize int64) *SystemTimer {
	delay := &DelayQueue{}
	counter := int32(0)
	startMs := time.Now().UnixMilli()
	timingWheel := NewTimingWheel(tickMs, wheelSize, startMs, &counter, delay)
	return &SystemTimer{
		delayQueue:  delay,
		taskCounter: &counter,
		timingWheel: timingWheel,
	}
}

func (this *SystemTimer) add(timerTask *TimerTask) {
	this.locker.RLock()
	defer this.locker.RUnlock()
	this.addTimerTaskEntry(NewTimerTaskEntry(timerTask, timerTask.delayMs+time.Now().UnixMilli()))
}

func (this *SystemTimer) addTimerTaskEntry(entry *TimerTaskEntry) {
	if !this.timingWheel.add(entry) {
		if !entry.cancelled() {
			this.group.Add(1)
			go func(ctx *sync.WaitGroup) {
				defer ctx.Done()
				entry.timerTask.f()
			}(&this.group)
		}
	}
}

func (this *SystemTimer) advanceClock(timeoutMs int64) bool {
	if timeoutMs == 0 {
		timeoutMs = this.timingWheel.tickMs
	}
	var bucket *TimerTaskList
	bucket = this.delayQueue.Poll(time.Duration(timeoutMs) * time.Millisecond)
	if bucket != nil {
		this.locker.Lock()
		defer this.locker.Unlock()
		for bucket != nil {
			this.timingWheel.advanceClock(bucket.getExpiration())
			bucket.flush(this.addTimerTaskEntry)
			bucket = this.delayQueue.PollE()
		}
		return true
	}
	return false
}

func (this *SystemTimer) run() {
	for {
		this.advanceClock(0)
	}
}

func (this *SystemTimer) size() int32 {
	return atomic.LoadInt32(this.taskCounter)
}

func (this *SystemTimer) shutdown() {
	this.group.Wait()
}
