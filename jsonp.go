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

	// scan file line by line
	// scan line for json tokens
	// using ScanRune. scan each character and match to a token.
	// use json mckenna format
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
	staticTokens := []string{"{", "}", ":", ",", "\"", "[", "]", "true", "false", "null"}
	for lineScanner.Scan() {
		runeScanner := bufio.NewScanner(bytes.NewReader(lineScanner.Bytes()))
		runeScanner.Split(bufio.ScanRunes)
		lineTokens := []string{}
		token := ""
		prevChar := rune(0)
		isLexingNumber := false
		isLexingString := false
		for runeScanner.Scan() {
			scannedBytes := runeScanner.Bytes()
			if runeScanner.Text() == "\n" {
				continue
			}
			char, _ := utf8.DecodeRune(scannedBytes)
			if string(char) == "\"" && token == "" {
				isLexingString = true
			}
			//skip spaces that do not exist within a string
			if unicode.IsSpace(char) && !isLexingString {
				continue
			}
			// save string token when we reach closing quote
			if string(char) == "\"" && prevChar != rune(0) && token != "" {
				isLexingString = false
				saveToken(&token, &lineTokens)
			}
			// TODO: handle escaped characters
			// handle integer digits
			// when we reach the first number, check if previousChar is : accumulate token until we hit a ,
			if string(prevChar) == ":" && string(char) != "," && unicode.IsNumber(char) || string(char) == "-" {
				isLexingNumber = true
			}
			// if we're lexing an integer accumulate the token until we hit a comma
			if isLexingNumber && unicode.IsNumber(prevChar) && string(char) == "," {
				isLexingNumber = false
				saveToken(&token, &lineTokens)
			}
			token += string(char)
			if slices.Contains(staticTokens, token) {
				saveToken(&token, &lineTokens)
			}
			prevChar = char

		}
		tokens = append(tokens, lineTokens)
	}
	for i := 0; i < len(tokens); i++ {

		fmt.Printf("Line: %#v \n", tokens[i])
	}
	return tokens
}
func saveToken(token *string, lineTokens *[]string) {
	*lineTokens = append(*lineTokens, *token)
	*token = ""

}
