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

func (s *Saga) GetStep(name StepName) sagaStep {
	for _, step := range s.allSteps() {
		if step.Is(string(name)) {
			return step
		}
	}
	return nil
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

	var failedIdx int
	for idx, step := range s.TxSteps {
		if step.IsCompleted() {
			continue
		}

		if step.IsFailed() {
			failedIdx = idx
			break
		}

		return step.Exec(ctx)
	}

	// Exec in reverse.
	for i := failedIdx - 1; i > -1; i-- {
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
	cx := 0
	failed := false
	for _, step := range s.TxSteps {
		if step.IsCompleted() {
			tx--
			cx++
		}
		if step.IsFailed() {
			failed = true
			break
		}
	}

	if tx == 0 {
		status := StatusSuccess
		s.Status = &status
		return
	}

	if !failed {
		status := StatusPending
		s.Status = &status
		return
	}

	status := StatusPending
	cxn := cx - 1
	for i := cxn; i > -1; i-- {
		step := s.CxSteps[i]
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
	if s.Status == nil {
		return false
	}
	status := *s.Status
	return (status == StatusFailed || status == StatusSuccess)
}

func (s *Saga) InitStep(name StepName, req []byte) bool {
	steps := s.allSteps()
	for i := range steps {
		step := steps[i]
		if step.Is(string(name)) {
			return step.Init(req)
		}
	}
	return false
}

func (s *Saga) CompleteStep(name StepName, res []byte) bool {
	steps := s.allSteps()
	for i := range steps {
		step := steps[i]
		if step.Is(string(name)) {
			return step.Complete(res)
		}
	}
	return false
}

func (s *Saga) FailStep(name StepName, err error) bool {
	steps := s.allSteps()
	for i := range steps {
		step := steps[i]
		if step.Is(string(name)) {
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
