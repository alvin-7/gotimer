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

func (t *SystemTimer) add(timerTask *TimerTask) {
	t.locker.RLock()
	defer t.locker.RUnlock()
	t.addTimerTaskEntry(NewTimerTaskEntry(timerTask, timerTask.delayMs+time.Now().UnixMilli()))
}

func (t *SystemTimer) addTimerTaskEntry(entry *TimerTaskEntry) {
	if !t.timingWheel.add(entry) {
		if !entry.cancelled() {
			t.group.Add(1)
			go func(ctx *sync.WaitGroup) {
				defer ctx.Done()
				entry.timerTask.f()
			}(&t.group)
		}
	}
}

func (t *SystemTimer) advanceClock(timeoutMs int64) bool {
	if timeoutMs == 0 {
		timeoutMs = t.timingWheel.tickMs
	}
	var bucket *TimerTaskList
	bucket = t.delayQueue.Poll(time.Duration(timeoutMs) * time.Millisecond)
	if bucket != nil {
		t.locker.Lock()
		defer t.locker.Unlock()
		for bucket != nil {
			t.timingWheel.advanceClock(bucket.getExpiration())
			bucket.flush(t.addTimerTaskEntry)
			bucket = t.delayQueue.PollE()
		}
		return true
	}
	return false
}

func (t *SystemTimer) run() {
	for {
		t.advanceClock(0)
	}
}

func (t *SystemTimer) size() int32 {
	return atomic.LoadInt32(t.taskCounter)
}

func (t *SystemTimer) shutdown() {
	t.group.Wait()
}
