package timer

import "sync"

type TimerTaskEntry struct {
	list         *TimerTaskList
	prev         *TimerTaskEntry
	next         *TimerTaskEntry
	timerTask    *TimerTask
	locker       *sync.Mutex
	expirationMs int64
}

func NewTimerTaskEntry(timeTask *TimerTask, expirationMs int64) *TimerTaskEntry {
	entry := &TimerTaskEntry{
		timerTask:    timeTask,
		expirationMs: expirationMs,
	}
	if entry.timerTask != nil {
		entry.timerTask.setTimerTaskEntry(entry)
	}
	return entry
}

func (entry *TimerTaskEntry) cancelled() bool {
	return entry.timerTask.getTimerTaskEntry() != entry
}

func (entry *TimerTaskEntry) compare(that *TimerTaskEntry) bool {
	return entry.expirationMs > that.expirationMs
}

func (entry *TimerTaskEntry) remove() {
	currentList := entry.list
	for currentList != nil {
		currentList.remove(entry)
		currentList = entry.list
	}
}
