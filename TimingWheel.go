package timer

import (
	"sync"
)

type TimingWheel struct {
	wheelSize     int64
	taskCount     *int32
	interval      int64
	buckets       []*TimerTaskList
	tickMs        int64
	currentTime   int64
	queue         DelayQueueUnit
	overflowWheel *TimingWheel
	locker        *sync.Mutex
}

func NewTimingWheel(tickMs int64, wheelSize int64, startMs int64, taskCounter *int32, queue DelayQueueUnit) *TimingWheel {
	interval := tickMs * wheelSize
	buckets := make([]*TimerTaskList, wheelSize)
	for idx := range buckets {
		buckets[idx] = NewTimerTaskList(taskCounter)
	}
	currentTime := startMs - (startMs % tickMs)
	return &TimingWheel{
		wheelSize:   wheelSize,
		taskCount:   taskCounter,
		interval:    interval,
		buckets:     buckets,
		tickMs:      tickMs,
		currentTime: currentTime,
		queue:       queue,
	}
}

func (wheel *TimingWheel) addOverflowWheel() {
	wheel.locker.Lock()
	defer wheel.locker.Unlock()
	if wheel.overflowWheel == nil {
		wheel.overflowWheel = NewTimingWheel(wheel.interval, wheel.wheelSize, wheel.currentTime, wheel.taskCount, wheel.queue)
	}
}

func (wheel *TimingWheel) add(timerTaskEntry *TimerTaskEntry) bool {
	expiration := timerTaskEntry.expirationMs
	if timerTaskEntry.cancelled() {
		return false
	} else if expiration < wheel.currentTime+wheel.tickMs { // slow than current time
		return false
	} else if expiration < wheel.currentTime+wheel.interval { // in this wheel
		virtualId := expiration / wheel.tickMs
		bucket := wheel.buckets[int(virtualId%wheel.wheelSize)]
		bucket.add(timerTaskEntry)

		if bucket.setExpiration(virtualId * wheel.tickMs) {
			wheel.queue.Push(bucket)
		}
		return true
	} else { // in overflow wheel
		if wheel.overflowWheel == nil {
			wheel.addOverflowWheel()
		}
		return wheel.overflowWheel.add(timerTaskEntry)
	}
}

func (wheel *TimingWheel) advanceClock(timeMs int64) {
	if timeMs >= wheel.currentTime+wheel.tickMs {
		wheel.currentTime = timeMs - (timeMs % wheel.tickMs)
		if wheel.overflowWheel != nil {
			wheel.overflowWheel.advanceClock(wheel.currentTime)
		}
	}
}
