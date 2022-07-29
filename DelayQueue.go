package timer

import (
	"container/heap"
	"sync"
	"time"
)

type DelayQueueUnit []*TimerTaskList

func (pq DelayQueueUnit) Len() int {
	return len(pq)
}

func (pq DelayQueueUnit) Less(i, j int) bool {
	return pq[i].compareTo(pq[j])
}

func (pq DelayQueueUnit) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *DelayQueueUnit) Push(x interface{}) {
	*pq = append(*pq, x.(*TimerTaskList))
}

func (pq *DelayQueueUnit) Pop() interface{} {
	old := *pq
	n := len(old)
	itm := old[n-1]
	*pq = old[:n-1]
	return itm
}

type DelayQueue struct {
	queue  DelayQueueUnit
	timer  *time.Timer
	locker sync.Mutex
}

func (delay *DelayQueue) Len() int {
	return delay.queue.Len()
}

func (delay *DelayQueue) Put(list *TimerTaskList) {
	heap.Push(&delay.queue, list)
}

func (delay *DelayQueue) Poll(nanos time.Duration) (bucket *TimerTaskList) {
	delay.locker.Lock()
	defer delay.locker.Unlock()
	for {
		bucket = delay.peek()
		if bucket == nil {
			if nanos <= 0 {
				return nil
			} else {
				delay.timer = time.NewTimer(nanos)
			}
		} else {
			d := bucket.getDelay()
			if d <= 0 {
				return delay.poll()
			}
			if nanos <= 0 {
				return nil
			}
			bucket = nil
			if nanos < d || delay.timer != nil {
				delay.timer.Stop()
				delay.timer = time.NewTimer(nanos)
			} else {
				delay.timer = time.NewTimer(d)
			}
		}
		<-delay.timer.C
	}
}

func (delay *DelayQueue) PollE() *TimerTaskList {
	delay.locker.Lock()
	defer delay.locker.Unlock()
	first := delay.peek()
	if first == nil || first.getDelay().Nanoseconds() > 0 {
		return nil
	} else {
		return delay.poll()
	}
}

func (delay *DelayQueue) peek() *TimerTaskList {
	return delay.queue[0]
}

func (delay *DelayQueue) poll() *TimerTaskList {
	return heap.Pop(&delay.queue).(*TimerTaskList)
}
