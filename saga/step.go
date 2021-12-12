package saga

import (
	"context"
	"encoding/json"
	"errors"
)

type StepFn func(ctx context.Context) error

type Step struct {
	Name   string `json:"name"`
	Async  bool   `json:"async"`
	StepFn `json:"-"`
	Req    json.RawMessage `json:"req"`
	Res    json.RawMessage `json:"res"`
	Err    string          `json:"err"`
	*StepStatus
}

func NewStep(name string, async bool, stepFn StepFn) *Step {
	return &Step{
		Name:   name,
		Async:  async,
		StepFn: stepFn,
	}
}

func (s *Step) Is(name string) bool {
	return s.Name == name
}

func (s *Step) IsAsync() bool {
	return s.Async
}

func (s *Step) Exec(ctx context.Context) error {
	if s.StepStatus == nil {
		return errors.New("step not initialized")
	}
	if !s.Status.CanTransition() {
		return errors.New("cannot transition from " + s.Status.String())
	}
	return s.StepFn(ctx)
}

func (s *Step) Fail(err error) bool {
	ok := s.StepStatus.Fail()
	if !ok {
		return false
	}
	s.Err = err.Error()
	return ok
}

func (s *Step) Complete(res []byte) bool {
	ok := s.StepStatus.Complete()
	if !ok {
		return false
	}
	s.Res = res
	return ok
}

func (s *Step) Init(req []byte) bool {
	if s.StepStatus != nil {
		return false
	}
	s.StepStatus = NewStepStatus()
	s.Req = req
	return true
}