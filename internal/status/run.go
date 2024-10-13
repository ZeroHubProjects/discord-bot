package status

import (
	"runtime/debug"
	"sync"
	"time"
)

const interval = time.Minute

func (s *StatusUpdater) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		s.runUpdates()
		s.Logger.Debugf("restarting in %v...", interval)
		time.Sleep(interval)
	}
}

func (s *StatusUpdater) runUpdates() {
	defer func() {
		if err := recover(); err != nil {
			s.Logger.Errorf("panicked: %v\nstack trace: %s", err, string(debug.Stack()))
		}
	}()

	s.Logger.Debugf("updating with %v interval...", interval)

	for {
		err := s.update()
		if err != nil {
			s.Logger.Errorf("failed to update: %v", err)
		}
		time.Sleep(interval)
	}
}
