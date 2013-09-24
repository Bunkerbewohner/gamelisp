package main

import "fmt"
import "regexp"
import "strconv"
import "strings"
import "errors"

var whitespaceRegex = regexp.MustCompile("^\\s+")
var stringRegex = regexp.MustCompile("^\"(?:\\.|[^\\\"]|\"\")*\"")
var intRegex = regexp.MustCompile("^-?[\\d,]+")
var floatRegex = regexp.MustCompile("^-?[\\d,]+[.]\\d*")
var symbolRegex = regexp.MustCompile("^[^0-9\\s(\\[{}\\])][^\\s(\\[{}\\])]*")
var keywordRegex = regexp.MustCompile("^:[^0-9\\s(\\[{}\\])][^\\s(\\[{}\\])]*")

// A parser function receives an input string and a read offset
// to parse string representations of data or code. it returns
// the parsed data and the next reading position
type ParserFunc func(input string, offset int) (Data, int)

func Parse(input string) (Data, error) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("Parser Error: %v", e)
		}
	}()

	data, _ := ParseAny(input, 0)

	if data == nil {
		return nil, errors.New("Failed to parse string")
	}

	return data, nil
}

func ParseList(input string, offset int) (Data, int) {
	list := CreateList()
	start, end, readPos := getDelimeters(input, offset)

	if start == '[' {
		// [x1 x2 ... xn] denotes the datatype list (not to be executed)
		list.PushBack(Symbol{"list"})
	} else if start == '{' {
		// {} denotes dictionaries
		list.PushBack(Symbol{"dict"})
	}

	for readPos < len(input)-1 && input[readPos] != end {
		item, endPos := ParseAny(input, readPos)

		if item == nil {
			return nil, readPos + 1
		}

		readPos = endPos
		list.PushBack(item)
	}

	return list, readPos + 1
}

func ParseDict(input string, offset int) (Data, int) {
	dict := CreateDict()
	_, end, readPos := getDelimeters(input, offset)

	for readPos < len(input)-1 && input[readPos] != end {
		if input[readPos] == ',' {
			readPos++
		}
		key, keyEnd := ParseAny(input, readPos)
		value, valueEnd := ParseAny(input, keyEnd)

		dict.entries[key] = value
		readPos = valueEnd
	}

	return dict, readPos + 1
}

func ParseString(input string, offset int) (Data, int) {
	str := stringRegex.FindString(input[offset:])
	length := len(str)

	if length == 0 {
		panic(fmt.Sprintf("Failed to parse string in input \"%s\"", input[offset:]))
	}

	return String{str[1 : length-1]}, offset + length
}

func ParseSymbol(input string, offset int) (Data, int) {
	str := symbolRegex.FindString(input[offset:])
	length := len(str)

	if length == 0 {
		panic(fmt.Sprintf("Failed to parse symbol in input \"%s\"", input[offset:]))
	}

	return Symbol{str}, offset + length
}

func ParseKeyword(input string, offset int) (Data, int) {
	str := symbolRegex.FindString(input[offset:])
	length := len(str)

	if length == 0 {
		panic(fmt.Sprintf("Failed to parse keyword in input \"%s\"", input[offset:]))
	}

	return Keyword{str}, offset + length
}
func ParseNumber(input string, offset int) (Data, int) {
	// try float first
	floatStr := floatRegex.FindString(input[offset:])
	if floatStr != "" {
		floatVal, err := strconv.ParseFloat(strings.Replace(floatStr, ",", "", -1), 64)
		if err == nil {
			return Float{floatVal}, offset + len(floatStr)
		}
	}

	// integer second
	intStr := intRegex.FindString(input[offset:])
	if intStr != "" {
		intVal, err := strconv.ParseInt(strings.Replace(intStr, ",", "", -1), 10, 32)
		if err == nil {
			return Int{int(intVal)}, offset + len(intStr)
		}
	}

	panic(fmt.Sprintf("Invalid number format \"%s\"", input[offset:]))
}

// Parses any data value
func ParseAny(input string, offset int) (Data, int) {
	// ignore whitespace
	for input[offset] == ' ' || input[offset] == '\t' || input[offset] == '\r' || input[offset] == 'n' {
		offset++
	}

	switch input[offset] {
	case '(', '[', '{':
		return ParseList(input, offset)
	case '"', '\'':
		return ParseString(input, offset)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return ParseNumber(input, offset)
	case '-':
		if input[offset+1] >= '0' && input[offset+1] <= '9' {
			return ParseNumber(input, offset)
		} else {
			return ParseSymbol(input, offset)
		}
	case ':':
		return ParseKeyword(input, offset)
	}

	return ParseSymbol(input, offset)
}

func getDelimeters(input string, offset int) (startDelim byte, endDelim byte, readPos int) {
	startDelim = input[offset]

	switch startDelim {
	case '(':
		endDelim = ')'
	case '[':
		endDelim = ']'
	case '{':
		endDelim = '}'
	case '"':
		endDelim = '"'
	case '\'':
		endDelim = '\''
	default:
		panic("Unexpected start delimeter encountered: " + string(startDelim))
	}

	readPos = offset + 1
	return
}
