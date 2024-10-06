package ahelp

import (
	"runtime/debug"
	"sync"
	"time"
)

const interval = time.Minute

func (h *AhelpHandler) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		h.runAhelp()
		h.Logger.Debugf("restarting in %v...", interval)
		time.Sleep(interval)
	}
}

func (h *AhelpHandler) runAhelp() {
	defer func() {
		if err := recover(); err != nil {
			h.Logger.Errorf("handler panicked: %v\nstack trace: %s", err, string(debug.Stack()))
		}
	}()

	h.Logger.Debug("registering handler and listening for messages...")
	h.Discord.AddHandler(h.handleAhelpMessage)

	for {
		time.Sleep(time.Minute)
	}
}
