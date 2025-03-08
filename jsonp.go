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
	scanner := bufio.NewScanner(bytes.NewReader(buf.Bytes()))
	scanner.Split(bufio.ScanLines)
	tokens := [][]string{}
	for scanner.Scan() {
		runeScanner := bufio.NewScanner(bytes.NewReader(scanner.Bytes()))
		runeScanner.Split(bufio.ScanRunes)
		lineTokens := []string{}
		token := ""
		prevChar := rune(0)
		for runeScanner.Scan() {

			fmt.Printf("Lines: %#v \n", lineTokens)
			scannedBytes := runeScanner.Bytes()
			if runeScanner.Text() == "\n" {
				continue
			}
			// return error token if not {, [, or alphanum
			char, _ := utf8.DecodeRune(scannedBytes)
			skipWhitespace(&char, runeScanner)
			if len(token)-2 > 0 {
				prevChar = rune(token[len(token)-2])
			}
			if string(char) == "\"" && prevChar != rune(0) {
				lineTokens = append(lineTokens, string(token))
			}
			token += string(char)

			fmt.Println("token:", token)
			fmt.Println("char:", string(char))
			fmt.Println("Prev char:", string(prevChar))
			switch string(char) {
			case "{", "}", ":", ",", "\"", "[", "]":
				lineTokens = append(lineTokens, string(char))
				token = ""
				continue
			}
			switch token {
			case "true", "false", "null":
				lineTokens = append(lineTokens, token)
				token = ""
				continue
			}

		}
		tokens = append(tokens, lineTokens)
	}
	fmt.Printf("Tokens: %#v \n", tokens)
	return tokens
}
func skipWhitespace(char *rune, scanner *bufio.Scanner) {
	for unicode.IsSpace(*char) {
		scanner.Scan()
		*char, _ = utf8.DecodeRune(scanner.Bytes())
	}
}
