package main

import (
	"fmt"
	"os"

	"github.com/cyberwlodarczyk/pustynia/code"
)

func main() {
	c, err := code.New()
	if err != nil {
		fmt.Printf("error generating new code: %v", err)
		os.Exit(1)
	}
	fmt.Printf("%s\n", c)
}
