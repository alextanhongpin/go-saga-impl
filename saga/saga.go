package saga

import (
	"context"
)

type sagaStep interface {
	executable
	identifiable
	completable
	failable
	initializable
	async
}

type Saga struct {
	Status  *Status    `json:"status"`
	TxSteps []sagaStep `json:"transactions"`
	CxSteps []sagaStep `json:"compensations"`
}

func NewSaga() *Saga {
	return &Saga{}
}

func (s *Saga) AddStep(tx, cx sagaStep) {
	s.TxSteps = append(s.TxSteps, tx)
	s.CxSteps = append(s.CxSteps, cx)
}

func (s *Saga) Exec(ctx context.Context) error {
	s.UpdateStatus()

	if s.Done() {
		return nil
	}

	var failed bool
	for _, step := range s.TxSteps {
		if step.IsCompleted() {
			continue
		}

		if step.IsFailed() {
			failed = true
			break
		}

		if err := step.Exec(ctx); err != nil {
			return err
		}
		if step.IsAsync() {
			return nil
		}
	}

	if !failed {
		return nil
	}

	// Exec in reverse.
	for i := len(s.CxSteps) - 1; i > -1; i-- {
		step := s.CxSteps[i]
		if step.IsCompleted() {
			continue
		}
		// For compensations, all steps MUST succeed.
		//if step.IsFailed() {
		//continue
		//}
		if err := step.Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *Saga) UpdateStatus() {
	tx := len(s.TxSteps)
	for _, step := range s.TxSteps {
		if step.IsCompleted() {
			tx--
		}
	}

	status := StatusPending
	if tx == 0 {
		status = StatusSuccess
	}

	cx := len(s.CxSteps)
	for _, step := range s.CxSteps {
		if step.IsCompleted() {
			cx--
		}
	}
	if cx == 0 {
		status = StatusFailed
	}
	s.Status = &status
}

func (s *Saga) Done() bool {
	return s.Status != nil && (*s.Status == StatusFailed || *s.Status == StatusSuccess)
}

func (s *Saga) InitStep(name string, req []byte) bool {
	for _, step := range s.allSteps() {
		if step.Is(name) {
			return step.Init(req)
		}
	}
	return false
}

func (s *Saga) CompleteStep(name string, res []byte) bool {
	for _, step := range s.allSteps() {
		if step.Is(name) {
			return step.Complete(res)
		}
	}
	return false
}

func (s *Saga) FailStep(name string, err error) bool {
	for _, step := range s.allSteps() {
		if step.Is(name) {
			return step.Fail(err)
		}
	}
	return false
}

func (s *Saga) allSteps() []sagaStep {
	var result []sagaStep
	result = append(result, s.TxSteps...)
	result = append(result, s.CxSteps...)
	return result
}
