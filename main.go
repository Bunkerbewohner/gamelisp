package main

import "fmt"
import "os"
import "bufio"

func main() {
	fmt.Printf("GameLISP v0.1\n")
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("\n> ")
		line, err := reader.ReadString('\n')

		if err != nil || (len(line) >= 4 && line[0:4] == "exit") {
			break // EOF
		}

		data, _ := ParseAny(line, 0)
		evaled := Evaluate(data, MainContext)

		fmt.Printf(evaled.String())
	}

	fmt.Printf("GAME OVER.\n")
}
