package gotimer

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestTimer(t *testing.T) {
	timer := NewSystemTimer(1000, 6)
	cnt := 20
	count := int32(cnt)
	for i := 0; i < cnt; i++ {
		idx := rand.Int63n(100)
		s := time.Now()
		t.Log("idx: ", idx)
		task := NewTimerTask(idx*1000, func() {
			t.Log("idx: ", idx, "offset: ", time.Now().Unix()-s.Unix())
			assert.Equal(t, idx, time.Now().Unix()-s.Unix())
		})
		timer.Add(task)
		if idx*1000 == 0 {
			count -= 1
		} else if rand.Intn(100) < 20 {
			count -= 1
			task.cancel()
			t.Log("idx: ", idx, "cancel")
		}
	}
	assert.Equal(t, count, timer.Size())
	go timer.Run()
	for timer.Size() > 0 {
	}
	timer.Shutdown()
}
