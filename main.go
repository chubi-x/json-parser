package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"unicode"
	"unicode/utf8"
)

func handleError(errMsg string, err error) {
	if err != nil {
		io.WriteString(os.Stderr, fmt.Sprintf("%s: %s \n", errMsg, err))
		os.Exit(1)
	}

}

var fileName string

func main() {
	// use json mckenna format
	var buf *bytes.Buffer = bytes.NewBuffer(make([]byte, 0))
	flag.StringVar(&fileName, "file", "", "Path to JSON file")
	flag.Parse()
	if fileName != "" {

		openFile, err := os.Open(fileName)

		_, copyErr := io.Copy(buf, openFile)
		handleError("Unable to read file ", err)
		handleError("Error opening file "+fileName, copyErr)
		defer openFile.Close()
	} else if jsonString := flag.Arg(0); jsonString == "" && fileName == "" {
		_, err := io.Copy(buf, os.Stdin)
		handleError("Unable to read from Stdin", err)
	} else {
		buf = bytes.NewBufferString(jsonString)
	}
	Lex(buf)
}
func Lex(buf *bytes.Buffer) [][]string {

	lineScanner := bufio.NewScanner(bytes.NewReader(buf.Bytes()))
	lineScanner.Split(bufio.ScanLines)
	tokens := [][]string{}
	staticTokens := []string{"{", "}", ":", ",", "\"", "[", "]", "true", "false", "null"}
	for lineScanner.Scan() {
		runeScanner := bufio.NewScanner(bytes.NewReader(lineScanner.Bytes()))
		runeScanner.Split(bufio.ScanRunes)
		lineTokens := []string{}
		token := ""
		prevToken := ""
		prevChar := rune(0)
		isLexingNumber := false
		isLexingString := false
		for runeScanner.Scan() {
			scannedBytes := runeScanner.Bytes()
			char, _ := utf8.DecodeRune(scannedBytes)
			if char == '"' && !isLexingString {
				isLexingString = true
			}
			//skip spaces that do not exist within a string
			if unicode.IsSpace(char) && !isLexingString {
				prevChar = char
				saveToken(&token, &lineTokens, &prevToken)
				continue
			}
			prevTokenIsQuote := prevToken == "\""
			//stop lexing key string after reaching end quote
			if char == ':' && prevTokenIsQuote {
				isLexingString = false
			}
			// save value string token when we reach closing quote. handles escaped quotes
			// TODO: This a naive implementation as it doesn't cover the edge case where the string "\\" is read as one token instead of 3
			if isLexingString && char == '"' && prevChar != rune(0) && prevChar != '\\' && prevTokenIsQuote {
				isLexingString = false
				saveToken(&token, &lineTokens, &prevToken)
			}
			isNegative := prevChar == '-' && unicode.IsNumber(char)
			if !isLexingString && (isNegative || unicode.IsNumber(char)) {
				isLexingNumber = true
			}
			isLexingFloat := char == '.'
			isLexingExponent := isExponent(char)
			isLexingExponentSign := isExponent(prevChar) && (char == '+' || char == '-')
			if isLexingNumber && !isLexingFloat && !isLexingExponent && !isLexingExponentSign && !unicode.IsNumber(char) {
				isLexingNumber = false
				saveToken(&token, &lineTokens, &prevToken)
			}
			// only save static tokens that are not part of a string
			// De Morgan's Law to the rescue. second condition was previously !(token !="\"" && isLexingString && prevTokenIsQuote)
			if slices.Contains(staticTokens, token) && (token == "\"" || !isLexingString || !prevTokenIsQuote) {
				saveToken(&token, &lineTokens, &prevToken)
			}
			token += string(char)
			prevChar = char

		}
		// save the last token that was accumulated
		saveToken(&token, &lineTokens, &prevToken)
		tokens = append(tokens, lineTokens)
	}
	for i := 0; i < len(tokens); i++ {
		fmt.Printf("Line: %#v \n", tokens[i])
	}
	return tokens
}

// in recursive descent parsers we write a method to match each "entity " in the string
// we also have methods that implement a production rule in the grammar, so basically we need function to match:
// keyword tokens, numbers, strings, objects, and arrays
func Parse(tokens [][]string) (bool, error) {
	// flatten tokens
	// need to use lookahead method to keep track of our current postiion,
	// so use a pointer instead of looping.
	// track current position, look at next position. call every single match function that we have.
	// if any of them match increase your pointer and start the process all over again
	// if not throw an error. what if there are multiple errors? then keep track of all errors.
	newTokens := slices.Concat(tokens...)
	pos := -1
	switch newTokens[0] {
	case "{":
		parseObject(newTokens[1:], &pos)
	case "[":
		parseArray(newTokens[1:], &pos)
	}
	return false, fmt.Errorf("Invalid JSON string. Expected { or [, got %s", newTokens[0])
}

func parseObject(tokens []string, pos *int) bool {

	// parse
	*pos += 1
	parseLeftCurlyBrace(tokens[*pos], pos)
	parseQuote(tokens[*pos], pos)
	parseString()
	parseQuote(tokens[*pos], pos)
	parseColon(tokens[*pos], pos)
	// is string or number or array

	parseRightCurlyBrace(tokens[*pos], pos)
	return false
	// parse colon
	// parse value
}
func parseArray(tokens []string, pos *int) {
	*pos += 1
	parseLeftSquareBrace(tokens[*pos], pos)
	// could be a string, number, object, or array
	parseRightSquareBrace(tokens[*pos], pos)
}
func parseColon(token string, pos *int) bool {
	return matchKeyword(token, ":", pos)
}
func parseLeftCurlyBrace(token string, pos *int) bool {
	return matchKeyword(token, "{", pos)
}
func parseRightCurlyBrace(token string, pos *int) bool {
	return matchKeyword(token, "}", pos)
}
func parseLeftSquareBrace(token string, pos *int) bool {
	return matchKeyword(token, "[", pos)
}
func parseRightSquareBrace(token string, pos *int) bool {
	return matchKeyword(token, "]", pos)
}
func parseComma(token string, pos *int) bool {
	return matchKeyword(token, ",", pos)
}
func parseQuote(token string, pos *int) bool {
	return matchKeyword(token, "\"", pos)
}
func parseString() {}
func parseNumber(token string, pos *int) bool {
	_, intErr := strconv.ParseInt(token, 10, 64)
	_, floatErr := strconv.ParseFloat(token, 64)
	if intErr != nil && floatErr != nil {
		return false
	}
	*pos += 1
	return true

}
func parseBool(token string, pos *int) bool {
	return matchKeyword(token, "true", pos) || matchKeyword(token, "false", pos) || matchKeyword(token, "null", pos)
}

func matchKeyword(token string, keyword string, pos *int) bool {
	if token == keyword {
		*pos += 1
		return true
	}
	return false
}

// lexing functions
func isExponent(char rune) bool {
	return (char == 'e' || char == 'E')
}
func saveToken(token *string, lineTokens *[]string, prevToken *string) {
	if *token != "" {
		*lineTokens = append(*lineTokens, *token)
		*prevToken = *token
		*token = ""
	}

}
