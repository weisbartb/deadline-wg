package deadlinewg_test

import (
	"github.com/stretchr/testify/require"
	deadlinewg "github.com/weisbartb/deadline-wg"
	"testing"
	"time"
)

func TestNewWaitGroupDeadline_Timeout(t *testing.T) {
	wg := deadlinewg.NewWaitGroup(time.Millisecond * 100)
	wg.Add(5)
	start := time.Now()
	err := wg.Wait()
	end := time.Now()
	require.ErrorIs(t, err, deadlinewg.ErrTimeout)
	require.WithinDuration(t, end, start, time.Millisecond*200)
	wg.Add(1)
	wg.Done()
	err = wg.Wait()
	require.NoError(t, err)
}

func TestNewWaitGroupDeadline_Complete(t *testing.T) {
	wg := deadlinewg.NewWaitGroup(time.Millisecond * 100)
	wg.Add(1)
	start := time.Now()
	go wg.Done()
	err := wg.Wait()
	end := time.Now()
	require.NoError(t, err)
	require.WithinDuration(t, end, start, time.Millisecond*200)
}
func TestWaitGroup_Labels(t *testing.T) {
	wg := deadlinewg.NewWaitGroup(time.Millisecond * 100)
	wg.LabeledAdd("Test", 1)
	start := time.Now()
	go wg.LabeledDone("Test")
	err := wg.Wait()
	end := time.Now()
	require.NoError(t, err)
	require.WithinDuration(t, end, start, time.Millisecond*200)
}
