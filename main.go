package main

import "fmt"

func main() {
	data, _ := ParseAny("(times 10 (print \"Hello World!\")", 0)
	fmt.Printf("%v", data)
}
