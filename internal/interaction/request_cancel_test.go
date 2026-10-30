package interaction

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/ccontext"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/protocol/push"
)

// Closing an account cancels requests even when no writePump is left to consume
// the outbound queue. A plain channel send used to keep the caller blocked.
func TestCanceledRequestDoesNotWaitForOutboundQueue(t *testing.T) {
	ctx, cancel := context.WithCancel(ccontext.WithInfo(context.Background(), &ccontext.GlobalConfig{UserID: "test"}))
	mgr := &LongConnMgr{send: make(chan Message)}
	cancel()
	done := make(chan error, 1)
	go func() { done <- mgr.SendReqWaitResp(ctx, &push.DelUserPushTokenReq{}, constant.LogoutMsg, nil) }()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("canceled SDK request remains blocked on outbound queue")
	}
}
