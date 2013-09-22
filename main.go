package main

import "fmt"
import "os"
import "bufio"

const VERSION = "0.1"

func main() {
	fmt.Printf("GameLISP %s\n", VERSION)
	reader := bufio.NewReader(os.Stdin)

	for {
		// read commands from stdin
		fmt.Printf("\n> ")
		line, err := reader.ReadString('\n')

		// check for exit command
		if err != nil || (len(line) >= 4 && line[0:4] == "exit") {
			break // EOF
		}

		// parse code into AST
		data, err := Parse(line)
		if err != nil {
			fmt.Println(err.Error())
			continue
		} else if data == nil {
			continue
		}

		// evaluate the expressions
		if result, err := Evaluate(data, MainContext); err == nil {
			if result != nil {
				fmt.Printf(result.String())
			}
		} else {
			fmt.Println(err.Error())
		}
	}

	fmt.Printf("GAME OVER.\n")
}
