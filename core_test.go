package main

import "testing"
import "fmt"

// Ensure that the intrinsic types are available
func TestTypes(t *testing.T) {
	types := []string{"Int", "Float", "String", "Bool", "List", "Dict", "Keyword",
		"Symbol", "NativeFunction", "NativeFunctionB"}

	for _, name := range types {
		result, err := EvaluateString(fmt.Sprintf("(type %s)", name), MainContext)
		if err != nil {
			t.Error(err.Error())
			continue
		}

		_, ok := result.(DataType)
		if !ok {
			t.Errorf("(type %s) did not yield a type objet: %s", name, result.String())
			continue
		}
	}
}

// Check some instances of the instrinsic types
func TestTypeInstances(t *testing.T) {
	instances := []string{"8", "0.5", "\"Hallo Welt\"", "true", "{:version 800}", ":keyword", "(symbol \"x\")", "map", "def"}

	for _, inst := range instances {
		result, err := EvaluateString(inst, MainContext)
		if err != nil {
			t.Error(err.Error())
			continue
		}

		expected := inst

		switch inst {
		case "(symbol \"x\")":
			expected = "x"
		case "map":
			expected = "native function"
		case "def":
			expected = "native function"
		}

		if result.String() != expected {
			t.Errorf("Expected %s, found %s", inst, result.String())
		}
	}
}
