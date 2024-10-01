package verification

import (
	"runtime/debug"
	"sync"
	"time"
)

const interval = time.Minute

func (h *ByondVerificationHandler) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		h.runVerification()
		h.Logger.Debugf("restarting in %v...", interval)
		time.Sleep(interval)
	}
}

func (h *ByondVerificationHandler) runVerification() {
	defer func() {
		if err := recover(); err != nil {
			h.Logger.Errorf("handler panicked: %v\nstack trace: %s", err, string(debug.Stack()))
		}
	}()

	h.Logger.Debug("checking verification message and registering handlers...")
	h.Discord.AddHandler(h.handleInteraction)

	for {
		h.updateVerificationMessage()
		time.Sleep(time.Minute)
	}
}
