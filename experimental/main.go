package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Printf("%d\n", int(time.Since(time.Date(2021, 11, 0, 0, 0, 0, 0, time.UTC)).Minutes()))
}
