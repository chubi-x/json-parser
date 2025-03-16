package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
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
			if isLexingString && char == '"' && prevChar != rune(0) && prevTokenIsQuote {
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
			if slices.Contains(staticTokens, token) {
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
