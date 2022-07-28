package timer

import (
	"container/heap"
	"fmt"
	"testing"
)

func DelayTest(t *testing.T) {
	q := new(DelayQueueUnit)

	heap.Init(q)

	fmt.Printf("\nAdr: %p\n", &q)
	q.Push(&TimerTaskEntry{})

	for i := 0; i < 5; i++ {
		heap.Push(q, &TimerTaskEntry{})
	}

	for q.Len() > 0 {
		fmt.Println("Item: " + heap.Pop(q).(string))
	}
}
