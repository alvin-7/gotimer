package timer

import (
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
	taskCounter atomic.Value
	timingWheel *TimingWheel
	locker      *sync.RWMutex
	group       *sync.WaitGroup
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
			}(this.group)
		}
	}
}

func (this *SystemTimer) advanceClock(timeoutMs int64) bool {
	if timeoutMs == 0 {
		timeoutMs = this.timingWheel.tickMs
	}
	bucket := this.delayQueue.Poll(time.Duration(timeoutMs) * time.Millisecond)
	if bucket != nil {
		this.locker.Lock()
		defer this.locker.Unlock()
		for bucket != nil {
			this.timingWheel.advanceClock(bucket.list.getExpiration())
			bucket.flush(this.addTimerTaskEntry)
			bucket = this.delayQueue.PollE()
			return true
		}
	}
	return false
}

func (this *SystemTimer) run() {
	for {
		this.advanceClock(0)
	}
}

func (this *SystemTimer) size() int64 {
	return this.taskCounter.Load().(int64)
}

func (this *SystemTimer) shutdown() {
	this.group.Wait()
}
