package user_output

import "fmt"

type UserMessenger struct {
}

func (um *UserMessenger) Send(msg string) error {
	fmt.Printf("%s\n", msg)
	return nil
}
