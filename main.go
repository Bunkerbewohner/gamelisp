package main

import "fmt"
import "flag"

const VERSION = "0.1"

func main() {
	flag.Parse()
	InitRuntime()
	fmt.Printf("Apollo %s\n", VERSION)

	go REPL()

	if flag.NArg() == 1 {
		var scriptfile = flag.Arg(0)
		_, err := EvaluateString("(import "+scriptfile+")", MainContext)
		if err != nil {
			panic(err)
		}
	}

	RunGamehost()

	ShutdownRuntime()
	fmt.Printf("GAME OVER.\n")
}
