package cli

import (
	"activity_log/api/constructs"
	"fmt"
	"strings"
	"time"
)

type CLIListener struct {
}

// TODO(Luca409): utilize timeout
func (clil *CLIListener) GetUserInput(timeout time.Duration) (*constructs.UserInput, error) {
	var inputString string
	fmt.Scanln(&inputString)

	return &constructs.UserInput{
		Text: strings.Trim(inputString, "\n"),
	}, nil
	// input := make(chan string, 1)
	// quitChan := make(chan struct{}, 1)
	// defer func() { quitChan <- struct{}{} }()
	// go getInput(input, quitChan)

	// for {
	// 	select {
	// 	case i := <-input:
	// 		return &constructs.UserInput{
	// 			Text: strings.Trim(i, "\n"),
	// 		}, nil
	// 	case <-time.After(timeout):
	// 		return nil, apperror.NewTimeoutError(fmt.Errorf("timed out"), timeout)
	// 	}
	// }
}

func getInput(input chan string, quitChan chan struct{}) {
	for {
		select {
		case <-quitChan:
			return
		default:
			var inputString string
			fmt.Scanln(&inputString)

			input <- inputString
		}
	}
}
