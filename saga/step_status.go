package saga

import "time"

type StatusHistory struct {
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type StepStatus struct {
	Status  Status          `json:"status"`
	History []StatusHistory `json:"status_history"`
}

func NewStepStatus() *StepStatus {
	s := &StepStatus{
		Status:  StatusPending,
		History: []StatusHistory{},
	}
	s.snapshot()
	return s
}

func (s *StepStatus) IsCompleted() bool {
	if s == nil {
		return false
	}
	return s.Status == StatusSuccess
}

func (s *StepStatus) IsFailed() bool {
	if s == nil {
		return false
	}
	return s.Status == StatusFailed
}

func (s *StepStatus) IsPending() bool {
	if s == nil {
		return false
	}
	return s.Status == StatusPending
}

func (s *StepStatus) Complete() bool {
	sts, ok := s.Status.Complete()
	if !ok {
		return false
	}
	s.Status = sts
	s.snapshot()
	return true
}

func (s *StepStatus) Fail() bool {
	sts, ok := s.Status.Fail()
	if !ok {
		return false
	}
	s.Status = sts
	s.snapshot()
	return true
}

func (s *StepStatus) snapshot() {
	s.History = append(s.History, StatusHistory{
		Status:    s.Status,
		CreatedAt: time.Now(),
	})
}
