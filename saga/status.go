// You can edit this code!
// Click here and start typing.
package saga

import (
	"encoding/json"
	"fmt"
	"strconv"
)

var statusTextByStatus = map[Status]string{
	StatusPending: "pending",
	StatusSuccess: "success",
	StatusFailed:  "failed",
}

var statusByStatusText = map[string]Status{
	"pending": StatusPending,
	"success": StatusSuccess,
	"failed":  StatusFailed,
}

type Status int

const (
	statusStart Status = iota

	StatusPending
	StatusSuccess
	StatusFailed

	statusEnd
)

func (s Status) Valid() bool {
	return s > statusStart && s < statusEnd
}

func (s Status) Fail() (Status, bool) {
	if s.CanTransition() {
		return StatusFailed, true
	}
	return s, false
}

func (s Status) Complete() (Status, bool) {
	if s.CanTransition() {
		return StatusSuccess, true
	}
	return s, false
}

func (s Status) Retry() (Status, bool) {
	if s.CanRetry() {
		return StatusPending, true
	}
	return s, false
}

func (s Status) CanTransition() bool {
	return s == StatusPending
}

func (s Status) CanRetry() bool {
	return s == StatusFailed
}

func (s Status) String() string {
	return statusTextByStatus[s]
}

func (s Status) MarshalJSON() ([]byte, error) {
	return strconv.AppendQuote(nil, s.String()), nil
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}

	sts := statusByStatusText[text]
	if !sts.Valid() {
		return fmt.Errorf("invalid status %q", text)
	}

	*s = sts
	return nil
}
