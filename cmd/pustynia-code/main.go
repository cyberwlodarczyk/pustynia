package main

import (
	"fmt"
	"log"

	"github.com/cyberwlodarczyk/pustynia"
)

func main() {
	c, err := pustynia.NewCode(nil)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%s\n", c)
}
