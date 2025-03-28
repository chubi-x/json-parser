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

func handleLexerError(errMsg string, err error) {
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
		handleLexerError("Unable to read file ", err)
		handleLexerError("Error opening file "+fileName, copyErr)
		defer openFile.Close()
	} else if jsonString := flag.Arg(0); jsonString == "" && fileName == "" {
		_, err := io.Copy(buf, os.Stdin)
		handleLexerError("Unable to read from Stdin", err)
	} else {
		buf = bytes.NewBufferString(jsonString)
	}
	tokens := Lex(buf)
	fmt.Println(Parse(slices.Concat(tokens...)))
}

// Function to extract JSON tokens from a buffer
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

// Parse a given slice of already lexed tokens
func Parse(tokens []string) (bool, error) {
	pos := -1
	if len(tokens) == 0 {
		return false, fmt.Errorf("Expected tokens but found nil")
	}
	if matchLeftCurlyBrace(tokens[pos+1]) {
		if _, err := parseObject(tokens, &pos, true); err != nil {
			return false, err
		}
		return true, nil
	} else if matchLeftSquareBrace(tokens[pos+1]) {
		if _, err := parseArray(tokens, &pos, true); err != nil {
			return false, err
		}
		return true, nil
	}
	nextToken(&pos)
	return false, fmt.Errorf("Invalid JSON string. Expected { or [, got %s", tokens[pos])
}

func parseObject(tokens []string, pos *int, args ...bool) (bool, error) {
	isOuterObject := false
	if len(args) > 0 {
		isOuterObject = args[0]
	}
	nextToken(pos)

	if matchRightCurlyBrace(tokens[*pos+1]) {
		if matchComma(tokens[*pos]) {
			return false, parserError(*pos, "token", "}")
		}
		if isOuterObject {
			if *pos+1 != len(tokens)-1 {
				return false, parserError(*pos, "EOF", tokens[*pos+1])
			}
		}
		return true, nil
	} else if !matchQuote(tokens[*pos+1]) { // parse key
		return false, parserError(*pos, "\"", tokens[*pos+1])
	} else {
		if _, err := parseString(tokens, pos); err != nil {
			return false, err
		}
	}
	if !matchColon(tokens[*pos+1]) {
		return false, parserError(*pos, ":", tokens[*pos+1])
	}
	nextToken(pos)
	if _, err := parseValues(tokens, pos); err != nil {
		return false, err
	}
	if *pos == len(tokens)-1 {
		return false, parserError(*pos, ", or ]", "EOF")
	}
	if !matchComma(tokens[*pos+1]) {

		if !matchRightCurlyBrace(tokens[*pos+1]) {
			return false, parserError(*pos, "}", tokens[*pos+1])
		}
		return true, nil
	}
	if _, err := parseObject(tokens, pos); err != nil {
		return false, err
	}
	return true, nil

}

func parseArray(tokens []string, pos *int, args ...bool) (bool, error) {

	isOuterArray := false
	if len(args) > 0 {
		isOuterArray = args[0]
	}
	nextToken(pos)

	if matchRightSquareBrace(tokens[*pos+1]) {
		if matchComma(tokens[*pos]) {
			return false, parserError(*pos, "token", "]")
		}
		if isOuterArray {
			if *pos+1 != len(tokens)-1 {
				return false, parserError(*pos, "EOF", tokens[*pos+1])
			}
		}
		return true, nil
	}

	if _, err := parseValues(tokens, pos); err != nil {
		return false, err
	}
	if *pos == len(tokens)-1 {
		return false, parserError(*pos, ", or ]", "EOF")
	}
	if !matchComma(tokens[*pos+1]) {

		if !matchRightSquareBrace(tokens[*pos+1]) {
			return false, parserError(*pos, "]", tokens[*pos+1])
		}
		return true, nil
	}
	if _, err := parseArray(tokens, pos); err != nil {
		return false, err
	}
	return true, nil
}

// Parse out a string,object,number, or array
func parseValues(tokens []string, pos *int) (bool, error) {
	if matchNumber(tokens[*pos+1]) {
		nextToken(pos)
	} else if matchLeftCurlyBrace(tokens[*pos+1]) {

		if _, err := parseObject(tokens, pos); err != nil {
			return false, err
		}
		nextToken(pos)
	} else if matchLeftSquareBrace(tokens[*pos+1]) {
		if _, err := parseArray(tokens, pos); err != nil {
			return false, err
		}
		nextToken(pos)
	} else if matchQuote(tokens[*pos+1]) {
		if _, err := parseString(tokens, pos); err != nil {
			return false, err
		}
	} else if matchBool(tokens[*pos+1]) {
		nextToken(pos)
	} else {
		return false, parserError(*pos, "token", tokens[*pos+1])
	}

	return true, nil
}
func parseString(tokens []string, pos *int) (bool, error) {
	nextToken(pos)
	if !matchQuote(tokens[*pos+1]) {
		nextToken(pos)
		if !matchQuote(tokens[*pos+1]) {
			return false, parserError(*pos, "\"", tokens[*pos])
		}
		nextToken(pos)
	} else {
		nextToken(pos)
	}
	return true, nil
}
func nextToken(pos *int) {
	*pos += 1
}
func matchColon(token string) bool {
	return matchKeyword(token, ":")
}
func matchLeftCurlyBrace(token string) bool {
	return matchKeyword(token, "{")
}
func matchRightCurlyBrace(token string) bool {
	return matchKeyword(token, "}")
}
func matchLeftSquareBrace(token string) bool {
	return matchKeyword(token, "[")
}
func parseRightSquareBrace(token string) bool {
	return matchKeyword(token, "]")
}
func matchComma(token string) bool {
	return matchKeyword(token, ",")
}
func matchQuote(token string) bool {
	return matchKeyword(token, "\"")
}
func matchNumber(token string) bool {
	_, intErr := strconv.ParseInt(token, 10, 64)
	_, floatErr := strconv.ParseFloat(token, 64)
	if intErr != nil && floatErr != nil {
		return false
	}
	return true

}
func matchBool(token string) bool {
	return matchKeyword(token, "true") || matchKeyword(token, "false") || matchKeyword(token, "null")
}

func matchKeyword(token string, keyword string) bool {
	if token == keyword {
		return true
	}
	return false
}

func parserError(pos int, expected string, got string) error {
	return fmt.Errorf("Error Parsing JSON at %d. Expected %s but got %s", pos, expected, got)
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
