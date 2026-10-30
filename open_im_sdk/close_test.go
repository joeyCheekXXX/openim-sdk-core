package open_im_sdk

import (
	"context"
	"testing"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/db_interface"
	"github.com/openimsdk/openim-sdk-core/v3/sdk_struct"
)

type closeDB struct {
	db_interface.DataBase
	calls int
}

func (d *closeDB) Close(context.Context) error { d.calls++; return nil }

// Account removal used to rely on process exit; local Close must not recreate a
// context as Logout does, including when an SDK-triggered logout arrives late.
func TestCloseReleasesOnlyThisInstance(t *testing.T) {
	a, b := NewLoginMgr(), NewLoginMgr()
	for _, mgr := range []*LoginMgr{a, b} {
		if !mgr.InitSDK(sdk_struct.IMConfig{PlatformID: 3}, closeConnListener{}) {
			t.Fatal("init failed")
		}
	}
	db := &closeDB{}
	a.db = db
	a.setLoginStatus(Logged)
	ctxA, ctxB := a.Context(), b.Context()
	if err := a.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ctxA.Err() == nil || ctxB.Err() != nil {
		t.Fatal("Close did not isolate instances")
	}
	if err := a.logout(ctxA, true); err != nil {
		t.Fatal(err)
	}
	if err := a.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if a.Context() != ctxA || db.calls != 1 || a.GetLoginStatus(context.Background()) != 0 {
		t.Fatal("closed instance was revived or closed twice")
	}
	if err := b.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
}

type closeConnListener struct{}

func (closeConnListener) OnConnecting()                 {}
func (closeConnListener) OnConnectSuccess()             {}
func (closeConnListener) OnConnectFailed(int32, string) {}
func (closeConnListener) OnKickedOffline()              {}
func (closeConnListener) OnUserTokenExpired()           {}
func (closeConnListener) OnUserTokenInvalid(string)     {}
