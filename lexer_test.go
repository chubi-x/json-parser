package main

import (
	"bytes"
	"testing"
)

func TestSingleLineObjectWithSpacesBetweenValues(t *testing.T) {
	buffer := bytes.NewBufferString("{\"hey\":\" null\", \"how far\":      \"i dey\", \"key2\": \"time\"}")
	tokens := Lex(buffer)
	numLines := len(tokens)
	numTokens := len(tokens[0])
	if numLines != 1 {
		t.Errorf("Expected one line of tokens. Got %d", numLines)
	}
	if numTokens != 25 {
		t.Errorf("Expected 26 tokens. Got: %d", numTokens)
	}
}
func TestKeyWithSpaceInsideString(t *testing.T) {
	buffer := bytes.NewBufferString(`{"a key": "value"}`)
	tokens := Lex(buffer)
	if len(tokens[0]) != 9 {
		t.Errorf("Expected 9 tokens, Got : %d", len(tokens[0]))
	}
}
func TestValueWithSpaceInsideString(t *testing.T) {

	buffer := bytes.NewBufferString(`{"key": "a value"}`)
	tokens := Lex(buffer)
	if len(tokens[0]) != 9 {
		t.Errorf("Expected 9 tokens, Got : %d", len(tokens[0]))
	}
}
func TestIntegerValueWithSpaceBefore(t *testing.T) {

	buffer := bytes.NewBufferString(`{"key": 4}`)
	tokens := Lex(buffer)
	if len(tokens[0]) != 7 {
		t.Errorf("Expected 7 tokens, Got : %d", len(tokens[0]))
	}
}
func TestIntegerValueWithoutSpaceBefore(t *testing.T) {

	buffer := bytes.NewBufferString(`{"key":4}`)
	tokens := Lex(buffer)
	if len(tokens[0]) != 7 {
		t.Errorf("Expected 7 tokens, Got : %d", len(tokens[0]))
	}
}
func TestFloatValueWithoutSpaceBefore(t *testing.T) {
	buffer := bytes.NewBufferString(`{"key": 4.5}`)
	tokens := Lex(buffer)
	if len(tokens[0]) != 7 {
		t.Errorf("Expected 7 tokens, Got : %d", len(tokens[0]))
	}
}
func TestFloatValueWithSpaceBefore(t *testing.T) {
	buffer := bytes.NewBufferString(`{"key":4.5}`)
	tokens := Lex(buffer)
	if len(tokens[0]) != 7 {
		t.Errorf("Expected 7 tokens, Got : %d", len(tokens[0]))
	}
}
func TestValueStringWithSpecialCharacters(t *testing.T) {

	buffer := bytes.NewBufferString(`{"key": "!£$%^&*()_+{}[,].:@~;'#\\|-+-="`)
	tokens := Lex(buffer)
	if len(tokens[0]) != 8 {
		t.Errorf("Expected 8 tokens, Got : %d", len(tokens[0]))
	}

}
func TestValueStringWithEscapedCharacters(t *testing.T) {

	buffer := bytes.NewBufferString(`{"key": "the boy said to me \" my friend where art thou? \""`)
	tokens := Lex(buffer)
	if len(tokens[0]) != 8 {
		t.Errorf("Expected 8 tokens, Got : %d", len(tokens[0]))
	}
}
func TestValueWithExponent(t *testing.T) {

	buffer := bytes.NewBufferString(`{"key": 10e1}`)
	tokens := Lex(buffer)
	if len(tokens[0]) != 7 {
		t.Errorf("Expected 7 tokens, Got : %d", len(tokens[0]))
	}
}
func TestValueWithPositiveExponent(t *testing.T) {

	buffer := bytes.NewBufferString(`{"key": 10e+10}`)
	tokens := Lex(buffer)
	if len(tokens[0]) != 7 {
		t.Errorf("Expected 7 tokens, Got : %d", len(tokens[0]))
	}

}
func TestValueWithNegativeExponent(t *testing.T) {

	buffer := bytes.NewBufferString(`{"key": 10e-10}`)
	tokens := Lex(buffer)
	if len(tokens[0]) != 7 {
		t.Errorf("Expected 7 tokens, Got : %d", len(tokens[0]))
	}

}

func TestValueWithUpperCaseExponent(t *testing.T) {

	buffer := bytes.NewBufferString(`{"key": 10E66}`)
	tokens := Lex(buffer)
	if len(tokens[0]) != 7 {
		t.Errorf("Expected 7 tokens, Got : %d", len(tokens[0]))
	}

}
func TestSingleLineNestedObjects(t *testing.T) {

	buffer := bytes.NewBufferString(`{"key": {"nested key": {"nested key 2": "value"}}}`)
	tokens := Lex(buffer)
	if len(tokens[0]) != 21 {
		t.Errorf("Expected 21 tokens, Got : %d", len(tokens[0]))
	}

}
func TestMultilineLineNestedObjects(t *testing.T) {

	buffer := bytes.NewBufferString(`
    {
    "key": {
      "nested key": {
          "nested key 2": "value"
          }
      }
    }
    `)
	tokens := Lex(buffer)
	if len(tokens) != 7 {
		t.Errorf("Expected 7 lines. Got: %d", len(tokens))
	}
	if len(tokens[0]) != 1 {
		t.Errorf("Expected 1 token on Line 1, Got : %d", len(tokens[0]))
	}
	if len(tokens[1]) != 5 {
		t.Errorf("Expected 5 token on Line 2, Got : %d", len(tokens[1]))
	}
	if len(tokens[2]) != 5 {
		t.Errorf("Expected 5 token on Line 3, Got : %d", len(tokens[2]))
	}
	if len(tokens[3]) != 7 {
		t.Errorf("Expected 7 tokens on Line 4, Got : %d", len(tokens[3]))
	}
	if len(tokens[4]) != 1 {
		t.Errorf("Expected 1 token on Line 5, Got : %d", len(tokens[4]))
	}
	if len(tokens[5]) != 1 {
		t.Errorf("Expected 1 token on Line 6, Got : %d", len(tokens[5]))
	}
	if len(tokens[0]) != 1 {
		t.Errorf("Expected 1 token on Line 7, Got : %d", len(tokens[6]))
	}
}
func TestArrayContainingFloats(t *testing.T)                {}
func TestArrayContainingIntegers(t *testing.T)              {}
func TestArrayContainingFloatsAndIntegers(t *testing.T)     {}
func TestArrayContainingStrings(t *testing.T)               {}
func TestArrayContainingStringsAndFloats(t *testing.T)      {}
func TestArrayContainingStringsAndIntegers(t *testing.T)    {}
func TestArrayContainingSingleObject(t *testing.T)          {}
func TestArrayContainingMultipleObjects(t *testing.T)       {}
func TestMultilineObjectWithMultipleElements(t *testing.T)  {}
func TestMultilineObjectWithArrayElement(t *testing.T)      {}
func TestSinglelineObjectWithArrayElement(t *testing.T)     {}
func TestSinglelineObjectWithMultipleElements(t *testing.T) {}
func TestMultilineAarrayWithMultipleElements(t *testing.T)  {}
func TestSinglelineAarrayWithMultipleElements(t *testing.T) {}
