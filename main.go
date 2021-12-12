package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/alextanhongpin/go-saga-2/saga"
)

func main() {
	sg := saga.NewSaga()
	sg.AddStep(NewBookFlightStep(), NewCancelFlightStep())
	sg.AddStep(NewConfirmFlightStep(), NewNoopStep())
	sg.InitStep("book_flight", []byte(`{"flight_name": "boieng"}`))
	log.Println(sg)
	if err := sg.Exec(context.Background()); err != nil {
		panic(err)
	}

	//log.Println(sg.CompleteStep("book_flight", []byte(`{"id": "1234"}`)))

	// When received event booking_flight_failed
	// 1. Fail the step
	// 2. Initialize the compensation step.
	// 3. Execute.
	log.Println(sg.FailStep("book_flight", errors.New("failed to book flight")))
	log.Println(sg.InitStep("cancel_flight", nil))
	if err := sg.Exec(context.Background()); err != nil {
		panic(err)
	}
	prettyPrint(sg)

	// When received event flight_cancelled
	// 1. Update the step
	// 2. Execute.

	log.Println(sg.CompleteStep("cancel_flight", nil))
	if err := sg.Exec(context.Background()); err != nil {
		panic(err)
	}
	prettyPrint(sg)
}

func prettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

type BookFlightStep struct {
	*saga.Step
}

func NewBookFlightStep() *BookFlightStep {
	s := &BookFlightStep{}
	s.Step = saga.NewStep("book_flight", true, s.Exec)
	return s
}

type BookFlightRequest struct {
	FlightName        string
	FlightDestination string
}

type BookFlightResponse struct {
	ID string
}

func (s *BookFlightStep) Exec(ctx context.Context) error {
	var req BookFlightRequest
	if err := json.Unmarshal(s.Step.Req, &req); err != nil {
		return err
	}
	fmt.Println("booking flight")
	return nil
}

type CancelFlightStep struct {
	*saga.Step
}

func NewCancelFlightStep() *CancelFlightStep {
	s := &CancelFlightStep{}
	s.Step = saga.NewStep("cancel_flight", false, s.Exec)
	return s
}

func (s *CancelFlightStep) Exec(ctx context.Context) error {
	fmt.Println("cancelling flight")
	return nil
}

type ConfirmFlightStep struct {
	*saga.Step
}

func NewConfirmFlightStep() *ConfirmFlightStep {
	s := &ConfirmFlightStep{}
	s.Step = saga.NewStep("confirm_flight", false, s.Exec)
	return s
}

func (s *ConfirmFlightStep) Exec(ctx context.Context) error {
	fmt.Println("confirming flight")
	return nil
}

type NoopStep struct {
	*saga.Step
}

func NewNoopStep() *NoopStep {
	s := &NoopStep{}
	s.Step = saga.NewStep("noop", false, s.Exec)
	return s
}

func (s *NoopStep) Exec(ctx context.Context) error {
	return nil
}
