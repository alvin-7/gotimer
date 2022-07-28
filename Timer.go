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

func (t *SystemTimer) add(timerTask *TimerTask) {
	t.locker.RLock()
	defer t.locker.RUnlock()
	t.addTimerTaskEntry(NewTimerTaskEntry(timerTask, timerTask.delayMs+time.Now().UnixMilli()))
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

func (t *SystemTimer) advanceClock(timeoutMs int64) bool {
	bucket := t.delayQueue.Poll(time.Duration(timeoutMs) * time.Millisecond)
	if bucket != nil {
		t.locker.Lock()
		defer t.locker.Unlock()
		for bucket != nil {
			t.timingWheel.advanceClock(bucket.list.getExpiration())
			bucket.flush(t.addTimerTaskEntry)
			bucket = t.delayQueue.PollE()
			return true
		}
	}
	return false
}

func (t *SystemTimer) size() int64 {
	return t.taskCounter.Load().(int64)
}

func (this *SystemTimer) shutdown() {
	this.group.Wait()
}
