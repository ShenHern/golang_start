package main

import (
	"fmt"

	"github.com/ShenHern/golang_start/greetings"
)

func main() {
	// get a greeting message and print it
	message := greetings.Hello("Shen Hern")
	fmt.Println(message)
}
