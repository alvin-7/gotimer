package gotimer

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
	queue DelayQueueUnit
	timer *time.Timer
	mux   sync.Mutex
}

func (delay *DelayQueue) Len() int {
	return delay.queue.Len()
}

func (delay *DelayQueue) Put(list *TimerTaskList) {
	heap.Push(&delay.queue, list)
}

func (delay *DelayQueue) poll(millis time.Duration) (bucket *TimerTaskList) {
	delay.mux.Lock()
	defer delay.mux.Unlock()
	for {
		bucket = delay.peek()
		if bucket == nil {
			if millis <= 0 {
				return nil
			} else {
				delay.timer = time.NewTimer(millis)
			}
		} else {
			bucketDelay := bucket.getDelay()
			if bucketDelay <= 0 {
				return delay.pop()
			}
			if millis <= 0 {
				return nil
			}
			bucket = nil
			if delay.timer != nil {
				delay.timer.Stop()
			}
			if millis < bucketDelay {
				delay.timer = time.NewTimer(millis)
			} else {
				delay.timer = time.NewTimer(bucketDelay)
			}
		}
		<-delay.timer.C
	}
}

func (delay *DelayQueue) pollE() *TimerTaskList {
	delay.mux.Lock()
	defer delay.mux.Unlock()
	first := delay.peek()
	if first == nil || first.getDelay().Milliseconds() > 0 {
		return nil
	} else {
		return delay.pop()
	}
}

func (delay *DelayQueue) peek() *TimerTaskList {
	if delay.queue.Len() > 0 {
		return delay.queue[0]
	} else {
		return nil
	}
}

func (delay *DelayQueue) pop() *TimerTaskList {
	return heap.Pop(&delay.queue).(*TimerTaskList)
}
