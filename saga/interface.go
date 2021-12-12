package saga

import "context"

type identifiable interface {
	Is(name string) bool
}

type executable interface {
	Exec(ctx context.Context) error
}

type completable interface {
	Complete(res []byte) bool
	IsCompleted() bool
}

type failable interface {
	Fail(err error) bool
	IsFailed() bool
}

type initializable interface {
	Init(req []byte) bool
	IsPending() bool
}

type async interface {
	IsAsync() bool
}
