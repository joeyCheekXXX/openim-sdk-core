package conversation_msg

import (
	"context"
	"errors"
	"testing"
	"time"
)

// In-process account removal must release idle sender workers; range(queue)
// previously kept them alive until the host process exited.
func TestSenderWorkersExitWithAccount(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	sender := newMessageSender(&Conversation{ctx: ctx})
	cancel()
	done := make(chan struct{})
	go func() { sender.wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("sender workers remain after account cancellation")
	}
	if err := sender.submit(&sendTask{}); !errors.Is(err, context.Canceled) {
		t.Fatal("closed account accepted a send", err)
	}
}
