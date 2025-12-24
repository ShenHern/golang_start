package main

import (
	"fmt"

	"log"

	"github.com/ShenHern/golang_start/greetings"
)

func main() {
	log.SetPrefix("greetings: ")
	log.SetFlags(0)
	// get a greeting message and print it
	message, err := greetings.Hello("")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(message)
}
