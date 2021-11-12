package user_input

import (
	"activity_log/api/constructs"
	"fmt"
	"time"
)

type service interface {
	GetUserInput(timeout time.Duration) (*constructs.UserInput, error)
}

type UserListener struct {
	listeningService service
}

func New(listeningService service) *UserListener {
	return &UserListener{
		listeningService: listeningService,
	}

}

// Try to get the input {maxRetries} times with a {timeoutLimit} on
// each individual retry and check response satisfies {invariants}.
func (ul *UserListener) GetUserInput(
	timeoutLimit time.Duration,
	maxRetries int,
	invariants func(*constructs.UserInput) error,
) (*constructs.UserInput, error) {
	retriesLeft := maxRetries

	var ui *constructs.UserInput
	var err error
	for retriesLeft >= 0 {
		ui, err = ul.listeningService.GetUserInput(timeoutLimit)
		if err != nil {
			retriesLeft--
			continue
		}

		if err = invariants(ui); err == nil {
			return ui, nil
		}

		// TODO(luca): use messenger
		fmt.Printf("Invalid input: %v\n", err)

		retriesLeft--
	}

	return nil, fmt.Errorf("failed after %d attempts. User input: %+v. Last err: %w", maxRetries, ui, err)
}
