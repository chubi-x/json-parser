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

var fileName string

const (
	LEFTCURLYBRACE   = "{"
	RIGHTCURLYBRACE  = "}"
	LEFTSQUAREBRACE  = "["
	RIGHTSQUAREBRACE = "]"
	QUOTE            = "\""
	TRUE             = "true"
	FALSE            = "false"
	NULL             = "null"
	COLON            = ":"
	COMMA            = ","
)

func readJson() *bytes.Buffer {

	var buf *bytes.Buffer = bytes.NewBuffer(make([]byte, 0))
	flag.StringVar(&fileName, "file", "", "Path to JSON file")
	flag.Parse()
	if fileName != "" {

		openFile, err := os.Open(fileName)

		_, copyErr := io.Copy(buf, openFile)
		handleFileReadError("Unable to read file ", err)
		handleFileReadError("Error opening file "+fileName, copyErr)
		defer openFile.Close()
	} else if jsonString := flag.Arg(0); jsonString == "" && fileName == "" {
		_, err := io.Copy(buf, os.Stdin)
		handleFileReadError("Unable to read from Stdin", err)
	} else {
		buf = bytes.NewBufferString(jsonString)
	}
	return buf
}
func main() {
	// use json mckenna format
	ParseJson()
}
func ParseJson() (bool, error) {
	json := readJson()
	tokens := Lex(json)
	return Parse(slices.Concat(tokens...))
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
	return tokens
}

func Parse(tokens []string) (bool, error) {

	// in recursive descent parsers we write a method to match each "entity " in the string
	// we also have methods that implement a production rule in the grammar, so basically we need function to match:
	// keyword tokens, numbers, strings, objects, and arrays
	pos := -1
	if len(tokens) == 0 {
		return false, fmt.Errorf("Expected tokens but found nil")
	}
	lastToken := tokens[len(tokens)-1]
	switch tokens[pos+1] {
	case LEFTCURLYBRACE:
		if lastToken != RIGHTCURLYBRACE {
			return false, parserError(len(tokens)-1, "}", lastToken)
		}
		if _, err := parseObject(tokens, &pos); err != nil {
			return false, err
		}
		return true, nil
	case LEFTSQUAREBRACE:

		if lastToken != RIGHTSQUAREBRACE {
			return false, parserError(len(tokens)-1, "}", lastToken)
		}
		if _, err := parseArray(tokens, &pos); err != nil {
			return false, err
		}
		return true, nil

	}
	nextToken(&pos)
	return false, fmt.Errorf("Invalid JSON string. Expected { or [, got %s", tokens[pos])
}

func parseObject(tokens []string, pos *int) (bool, error) {

	for {
		nextToken(pos)
		switch tokens[*pos+1] {

		case RIGHTCURLYBRACE:
			if matchComma(tokens[*pos]) {
				return false, parserError(*pos, "token", "}")
			}
			return true, nil
		case QUOTE:
			if _, err := parseString(tokens, pos); err != nil {
				return false, err
			}
			nextToken(pos)
		default:
			return false, parserError(*pos, "\"", tokens[*pos+1])
		}
		if tokens[*pos+1] != COLON {
			return false, parserError(*pos, ":", tokens[*pos+1])
		}
		nextToken(pos)
		if _, err := parseValues(tokens, pos); err != nil {
			return false, err
		}
		if *pos == len(tokens)-1 {
			return false, parserError(*pos, ", or }", "EOF")
		}
		if !matchComma(tokens[*pos+1]) {

			if !matchRightCurlyBrace(tokens[*pos+1]) {
				return false, parserError(*pos, "}", tokens[*pos+1])
			}
			return true, nil

		}
	}

}

func parseArray(tokens []string, pos *int) (bool, error) {

	for {
		nextToken(pos)

		if matchRightSquareBrace(tokens[*pos+1]) {
			if matchComma(tokens[*pos]) {
				return false, parserError(*pos, "token", "]")
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
	}
}

// Parse out a string,object,number, or array
func parseValues(tokens []string, pos *int) (bool, error) {
	switch tokens[*pos+1] {
	case LEFTCURLYBRACE:
		if _, err := parseObject(tokens, pos); err != nil {
			return false, err
		}
	case LEFTSQUAREBRACE:
		if _, err := parseArray(tokens, pos); err != nil {
			return false, err
		}
	case QUOTE:
		if _, err := parseString(tokens, pos); err != nil {
			return false, err
		}
	case TRUE, FALSE, NULL:
	default:
		if !matchNumber(tokens[*pos+1]) {
			return false, parserError(*pos, "token", tokens[*pos+1])
		}
	}
	nextToken(pos)
	return true, nil
}
func parseString(tokens []string, pos *int) (bool, error) {
	nextToken(pos)
	if !matchQuote(tokens[*pos+1]) {
		nextToken(pos)
		if !matchQuote(tokens[*pos+1]) {
			return false, parserError(*pos, "\"", tokens[*pos])
		}
	}
	return true, nil
}
func nextToken(pos *int) {
	*pos += 1
}
func matchRightSquareBrace(token string) bool {
	return matchKeyword(token, RIGHTSQUAREBRACE)
}
func matchRightCurlyBrace(token string) bool {
	return matchKeyword(token, RIGHTCURLYBRACE)
}
func matchComma(token string) bool {
	return matchKeyword(token, COMMA)
}
func matchQuote(token string) bool {
	return matchKeyword(token, QUOTE)
}
func matchNumber(token string) bool {
	_, intErr := strconv.ParseInt(token, 10, 64)
	_, floatErr := strconv.ParseFloat(token, 64)
	if intErr != nil && floatErr != nil {
		return false
	}
	return true
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
func handleFileReadError(errMsg string, err error) {
	if err != nil {
		io.WriteString(os.Stderr, fmt.Sprintf("%s: %s \n", errMsg, err))
		os.Exit(1)
	}

}
