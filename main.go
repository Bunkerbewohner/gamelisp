package main

import "fmt"

func main() {
	data, _ := ParseAny("(times 10 (print \"Hello World!\") False { :name \"Mathias\", :age 25, :flag true }", 0)
	fmt.Printf("%v", data)
}
