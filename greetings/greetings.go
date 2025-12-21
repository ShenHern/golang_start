package greetings

import "fmt"

// Hello returns a greeting for the named person
func Hello(name string) string {
	// return a greeting that embeds a name in the string
	message := fmt.Sprintf("Hi, %v. Welcome!", name)
	return message
}
