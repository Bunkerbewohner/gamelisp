package main

import "fmt"

const VERSION = "0.1"

func main() {
	InitRuntime()

	go REPL()

	RunGamehost()

	ShutdownRuntime()
	fmt.Printf("GAME OVER.\n")
}
