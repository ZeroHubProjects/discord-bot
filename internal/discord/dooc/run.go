package dooc

import (
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

const interval = time.Minute

func (h *DOOCHandler) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		h.runDOOC()
		h.Logger.Debugf("restarting in %v...", interval)
		time.Sleep(interval)
	}
}

func (h *DOOCHandler) runDOOC() {
	defer func() {
		if err := recover(); err != nil {
			h.Logger.Errorf("handler panicked: %v\nstack trace: %s", err, string(debug.Stack()))
		}
	}()

	h.Logger.Debug("registering handler and listening for messages...")
	h.Discord.AddHandler(h.handleDOOCMessage)

	// keep processing unsent messages if any
	for {
		msgs, err := h.Discord.ChannelMessages(h.OOCChannelID, 50, "", "", "")
		if err != nil {
			h.Logger.Errorf("failed to get messages from the channel: %w", err)
		}
		for _, msg := range msgs {
			if strings.HasPrefix(msg.Content, retryMarker) {
				h.retryMessage(msg)
			}
		}
		time.Sleep(time.Minute)
	}
}
