package timer

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestTimer(t *testing.T) {
	timer := NewSystemTimer(1000, 60)
	timer.add(NewTimerTask(1000, func() { t.Log(1000) }))
	// timer.run()
	timer.advanceClock(0)
}
