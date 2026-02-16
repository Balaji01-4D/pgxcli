package commands

import "fmt"

const escCode = "\033[2J\033[H\033[3J"

func ClearScreen() {
	fmt.Printf(escCode)
}