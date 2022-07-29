package timer

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
	for i := 0; i < 20; i++ {
		idx := rand.Int63n(100)
		s := time.Now()
		t.Log("idx: ", idx)
		timer.add(NewTimerTask(idx*1000, func() {
			t.Log("idx: ", idx, "offset: ", time.Now().Unix()-s.Unix())
			assert.Equal(t, idx, time.Now().Unix()-s.Unix())
		}))
	}
	go timer.run()
	for timer.size() > 0 {
	}
	timer.shutdown()
}
