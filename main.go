package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/alextanhongpin/go-saga-2/saga"
)

var (
	BookFlight      saga.StepName = "book_flight"
	CancelFlight    saga.StepName = "cancel_flight"
	ConfirmFlight   saga.StepName = "confirm_flight"
	TerminateFlight saga.StepName = "terminate_flight"
)

func main() {
	ctx := context.Background()
	sg := NewBookingSaga()
	prettyPrint(sg)

	fmt.Println("onBookFlight", sg.OnBookFlightRequest(ctx, []byte(`{"flight_name": "flight-1"}`)))
	prettyPrint(sg)

	fmt.Println("onBookFlightSuccess", sg.OnFlightBookingSuccess(ctx, []byte(`{"id": "123"}`)))
	prettyPrint(sg)

	//fmt.Println("onFlightConfirmationSuccess", sg.OnFlightConfirmationSuccess(ctx, []byte(`{"id": "123"}`)))
	//prettyPrint(sg)

	fmt.Println("onFlightConfirmationFailed", sg.OnFlightConfirmationFailed(ctx, errors.New("flight full")))
	prettyPrint(sg)

	fmt.Println("onFlightCancelled", sg.OnFlightCancelled(ctx, nil))
	prettyPrint(sg)
}

type BookingSaga struct {
	*saga.Saga
}

func NewBookingSaga() *BookingSaga {
	s := &BookingSaga{
		Saga: saga.NewSaga(),
	}
	s.AddStep(
		saga.NewStep(BookFlight, false, s.BookFlight),
		saga.NewStep(CancelFlight, false, s.CancelFlight),
	)
	s.AddStep(
		saga.NewStep(ConfirmFlight, false, s.ConfirmFlight),
		saga.NewStep(TerminateFlight, false, s.Noop),
	)
	return s
}

type BookFlightRequest struct {
	FlightName string `json:"flight_name"`
}

func (s *BookingSaga) OnBookFlightRequest(ctx context.Context, req []byte) error {
	ok := s.InitStep(BookFlight, req)
	if !ok {
		return fmt.Errorf("book_flight step not initialized")
	}
	return s.Exec(ctx)
}

func (s *BookingSaga) BookFlight(ctx context.Context) error {
	stp := s.GetStep(BookFlight).(*saga.Step)

	var req BookFlightRequest
	if err := json.Unmarshal(stp.Req, &req); err != nil {
		return err
	}
	fmt.Println("booking flight", req)
	return nil
}

func (s *BookingSaga) OnFlightBookingSuccess(ctx context.Context, res []byte) error {
	ok := s.CompleteStep(BookFlight, res)
	if !ok {
		return fmt.Errorf("book_flight step not completed")
	}
	ok = s.InitStep(ConfirmFlight, res)
	if !ok {
		return fmt.Errorf("confirm_flight step not initialized")
	}
	return s.Exec(ctx)
}

func (s *BookingSaga) OnFlightBookingFailed(ctx context.Context, err error) error {
	ok := s.FailStep(BookFlight, err)
	if !ok {
		return fmt.Errorf("book_flight step not failed")
	}
	return s.Exec(ctx)
}

func (s *BookingSaga) CancelFlight(ctx context.Context) error {
	return nil
}

func (s *BookingSaga) OnFlightCancelled(ctx context.Context, res []byte) error {
	fmt.Println("flight cancelled")
	ok := s.CompleteStep(CancelFlight, nil)
	if !ok {
		return fmt.Errorf("flight cancelled step not initialized")
	}
	return s.Exec(ctx)
}

func (s *BookingSaga) ConfirmFlight(ctx context.Context) error {
	fmt.Println("confirming flight")
	return nil
}

func (s *BookingSaga) OnFlightConfirmationSuccess(ctx context.Context, res []byte) error {
	fmt.Println("flight confirmed")
	ok := s.CompleteStep(ConfirmFlight, res)
	if !ok {
		return fmt.Errorf("flight confirmation success step not initialized")
	}
	return s.Exec(ctx)
}

func (s *BookingSaga) OnFlightConfirmationFailed(ctx context.Context, err error) error {
	fmt.Println("failed to confirm flight")
	ok := s.FailStep(ConfirmFlight, err)
	if !ok {
		return fmt.Errorf("flight confirmation failed step not initialized")
	}
	ok = s.InitStep(CancelFlight, nil)
	if !ok {
		return fmt.Errorf("cancel_flight step not initialized")
	}
	return s.Exec(ctx)
}

func (s *BookingSaga) Noop(ctx context.Context) error {
	return nil
}

func prettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
