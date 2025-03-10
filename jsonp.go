package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
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

	// scan file line by line
	// scan line for json tokens
	// using ScanRune. scan each character and match to a token.
	// use json mckenna format
	// validate elements first, where element  = ws value ws
	var buf *bytes.Buffer = bytes.NewBuffer(make([]byte, 0))

	flag.Parse()

	if file_arg := flag.Arg(0); file_arg == "" {
		_, err := io.Copy(buf, os.Stdin)
		handleError("Unable to read from Stdin", err)
	} else {
		fileName = file_arg
		open_file, err := os.Open(fileName)

		_, copyErr := io.Copy(buf, open_file)
		handleError("Unable to read file ", err)
		handleError("Error opening file "+fileName, copyErr)
		defer open_file.Close()
	}
	Lex(buf)
}
func Lex(buf *bytes.Buffer) [][]string {
	lineScanner := bufio.NewScanner(bytes.NewReader(buf.Bytes()))
	lineScanner.Split(bufio.ScanLines)
	tokens := [][]string{}
	for lineScanner.Scan() {
		runeScanner := bufio.NewScanner(bytes.NewReader(lineScanner.Bytes()))
		runeScanner.Split(bufio.ScanRunes)
		lineTokens := []string{}
		token := ""
		prevChar := rune(0)
		for runeScanner.Scan() {
			scannedBytes := runeScanner.Bytes()
			if runeScanner.Text() == "\n" {
				continue
			}
			char, _ := utf8.DecodeRune(scannedBytes)
			if unicode.IsSpace(char) && !unicode.IsLetter(prevChar) && string(prevChar) != "\"" {
				continue
			}
			// save string token when we reach closing quote
			if string(char) == "\"" && prevChar != rune(0) && token != "" {
				lineTokens = append(lineTokens, string(token))
			}
			token += string(char)
			processStaticTokensAndContinue(char, &token, &lineTokens)
			prevChar = char

		}
		tokens = append(tokens, lineTokens)
	}
	for i := 0; i < len(tokens); i++ {

		fmt.Printf("Line: %#v \n", tokens[i])
	}
	return tokens
}
func processStaticTokensAndContinue(char rune, token *string, lineTokens *[]string) {
	goToNextToken := false
	switch string(char) {
	case "{", "}", ":", ",", "\"", "[", "]":
		*lineTokens = append(*lineTokens, string(char))
		goToNextToken = true
	}
	switch *token {
	case "true", "false", "null":
		*lineTokens = append(*lineTokens, *token)
		goToNextToken = true
	}
	if goToNextToken {
		*token = ""
	}
}
