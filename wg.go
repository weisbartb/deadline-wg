package deadlinewg

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type WaitGroup struct {
	sync.WaitGroup
	m       map[string]int
	mu      sync.Mutex
	ct      int
	maxWait time.Duration
}

var ErrTimeout = errors.New("timeout")

func (dwg *WaitGroup) LabeledAdd(label string, count int) {
	dwg.Add(count)
	dwg.mu.Lock()
	defer dwg.mu.Unlock()
	dwg.ct += count
	if _, found := dwg.m[label]; !found {
		dwg.m[label] = 0
	}
	dwg.m[label]++
}

func (dwg *WaitGroup) LabeledDone(label string) {
	dwg.mu.Lock()
	defer dwg.mu.Unlock()
	if ct, found := dwg.m[label]; found {
		if ct <= 0 {
			panic(errors.Errorf("%v has underflowed", label))
		}
		dwg.m[label] = ct - 1
		dwg.ct--
	} else {
		panic(errors.Errorf("%v is an unregistered label", label))
	}
	dwg.Done()
}

func (dwg *WaitGroup) Add(count int) {
	dwg.mu.Lock()
	defer dwg.mu.Unlock()
	dwg.ct += count
	dwg.WaitGroup.Add(count)
}
func (dwg *WaitGroup) Done() {
	dwg.mu.Lock()
	defer dwg.mu.Unlock()
	dwg.ct--
	dwg.WaitGroup.Done()
}

func (dwg *WaitGroup) flush() {
	dwg.mu.Lock()
	defer dwg.mu.Unlock()
	for dwg.ct > 0 {
		dwg.ct--
		dwg.WaitGroup.Done()
	}
}

func NewWaitGroup(maxWait time.Duration) *WaitGroup {
	return &WaitGroup{
		maxWait: maxWait,
		m:       make(map[string]int),
	}
}
func (dwg *WaitGroup) Wait() error {
	t := time.NewTimer(dwg.maxWait)
	done := make(chan struct{})
	go func() {
		dwg.WaitGroup.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-t.C:
		dwg.mu.Lock()
		if len(dwg.m) > 0 {
			buf := strings.Builder{}
			ct := 0
			for k, v := range dwg.m {
				if v > 0 {
					if ct > 0 {
						buf.WriteByte('\n')
					}
					buf.WriteString(fmt.Sprintf("expected to see label [%v] %v more times", k, v))
					ct++
				}
			}
			dwg.mu.Unlock()
			dwg.flush()
			return errors.New(buf.String())
		}
		dwg.mu.Unlock()
		dwg.flush()
		return ErrTimeout
	}
}
