package timer

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDelayQueue(t *testing.T) {
	delay := &DelayQueue{}
	taskCounter := int32(0)
	itemL := 10
	for i := 0; i < itemL; i++ {
		list := NewTimerTaskList(&taskCounter)
		exp := rand.Int63n(10000)
		list.setExpiration(exp)
		delay.Put(list)
		t.Log("push", exp)
	}
	t.Log("------------------------")

	queue := make(DelayQueueUnit, len(delay.queue))
	copy(queue, delay.queue)
	sort.Sort(queue)

	assert.Equal(t, itemL, delay.Len())

	i := 0
	for delay.Len() > 0 {
		item := delay.PollE().getExpiration()
		assert.Equal(t, queue[i].getExpiration(), item)
		i += 1
	}
}
