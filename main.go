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

		if err != nil {
			break // EOF
		}

		data, _ := ParseAny(line, 0)
		fmt.Printf(data.String())
	}

	fmt.Printf("GAME OVER.\n")
}
