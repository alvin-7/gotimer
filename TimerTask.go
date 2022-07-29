package gotimer

import "sync"

type TimerTask struct {
	delayMs        int64
	timerTaskEntry *TimerTaskEntry
	locker         sync.Mutex
	f              func()
}

func NewTimerTask(delayMs int64, f func()) *TimerTask {
	return &TimerTask{
		delayMs: delayMs,
		f:       f,
	}
}

func (task *TimerTask) cancel() {
	task.locker.Lock()
	defer task.locker.Unlock()
	if task.timerTaskEntry != nil {
		task.timerTaskEntry.remove()
	}
	task.timerTaskEntry = nil
}

func (task *TimerTask) setTimerTaskEntry(entry *TimerTaskEntry) {
	task.locker.Lock()
	defer task.locker.Unlock()
	if task.timerTaskEntry != nil && task.timerTaskEntry != entry {
		task.timerTaskEntry.remove()
	}
	task.timerTaskEntry = entry
}

func (task *TimerTask) getTimerTaskEntry() *TimerTaskEntry {
	return task.timerTaskEntry
}
