package gotimer

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
	mux         sync.RWMutex
	group       sync.WaitGroup
}

func NewSystemTimer(tickMs int64, wheelSize int64) *SystemTimer {
	delay := &DelayQueue{}
	counter := int32(0)
	startMs := time.Now().UnixMilli()
	timingWheel := newTimingWheel(tickMs, wheelSize, startMs, &counter, delay)
	return &SystemTimer{
		delayQueue:  delay,
		taskCounter: &counter,
		timingWheel: timingWheel,
	}
}

func (t *SystemTimer) Add(timerTask *TimerTask) {
	t.mux.RLock()
	defer t.mux.RUnlock()
	t.addTimerTaskEntry(newTimerTaskEntry(timerTask, timerTask.delayMs+time.Now().UnixMilli()))
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

func (t *SystemTimer) AdvanceClock(timeoutMs int64) bool {
	if timeoutMs == 0 {
		timeoutMs = t.timingWheel.tickMs
	}
	var bucket *TimerTaskList
	bucket = t.delayQueue.poll(time.Duration(timeoutMs) * time.Millisecond)
	if bucket != nil {
		t.mux.Lock()
		defer t.mux.Unlock()
		for bucket != nil {
			t.timingWheel.advanceClock(bucket.getExpiration())
			bucket.flush(t.addTimerTaskEntry)
			bucket = t.delayQueue.pollE()
		}
		return true
	}
	return false
}

func (t *SystemTimer) RunAndShutDown() {
	for t.Size() > 0 {
		t.AdvanceClock(0)
	}
	t.Shutdown()
}

func (t *SystemTimer) Size() int32 {
	return atomic.LoadInt32(t.taskCounter)
}

func (t *SystemTimer) Shutdown() {
	t.group.Wait()
}
