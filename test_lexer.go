package main

import (
	"bytes"
	"testing"
)

func TestLexer(t *testing.T) {
	var buffer *bytes.Buffer = bytes.NewBufferString("echo {\"hey\":\" null\", \"how far\":      \"i dey\", \"key2\": \"time\"}")
	Lex(buffer)
}
func TestKeyWithSpaceInsideString(t *testing.T)             {}
func TestValueWithSpaceInsideString(t *testing.T)           {}
func TestValueStringWithSpecialCharacters(t *testing.T)     {}
func TestValueStringWithEscapedCharacters(t *testing.T)     {}
func TestValueWithInteger(t *testing.T)                     {}
func TestValueWithFloat(t *testing.T)                       {}
func TestValueWithExponent(t *testing.T)                    {}
func TestValueWithPositiveExponent(t *testing.T)            {}
func TestValueWithNegativeExponent(t *testing.T)            {}
func TestSingleLineNestedObjects(t *testing.T)              {}
func TestMultilineLineNestedObjects(t *testing.T)           {}
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
