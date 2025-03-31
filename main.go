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

type lexnexttokenparams struct {
	lineTokens       *[]string
	token            *string
	prevToken        *string
	prevChar         *rune
	char             *rune
	isLexingString   *bool
	prevTokenIsQuote *bool
}

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

var staticTokens = []string{LEFTCURLYBRACE, RIGHTCURLYBRACE, LEFTSQUAREBRACE, RIGHTSQUAREBRACE, QUOTE, COLON, TRUE, FALSE, NULL, COMMA}

func readJson() *bytes.Buffer {

	var fileName string
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
	fmt.Println(ParseJson())
}
func ParseJson() (bool, error) {
	json := readJson()
	tokens := slices.Concat(Lex(json)...)
	return Parse(tokens)
}

// Function to extract JSON tokens from a buffer
func Lex(buf *bytes.Buffer) [][]string {

	lineScanner := bufio.NewScanner(bytes.NewReader(buf.Bytes()))
	lineScanner.Split(bufio.ScanLines)
	tokens := [][]string{}
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
			prevTokenIsQuote := prevToken == "\""
			if char == '"' && !isLexingString {
				isLexingString = true
			}
			//skip spaces that do not exist within a string
			if unicode.IsSpace(char) && !isLexingString {
				prevChar = char
				saveToken(&token, &lineTokens, &prevToken)
				continue
			}
			// save value string token when we reach closing quote. handles escaped quotes
			// TODO: This a naive implementation as it doesn't cover the edge case where the string "\\" is read as one token instead of 3
			if isLexingString && char == '"' && prevChar != rune(0) && prevChar != '\\' && prevTokenIsQuote {
				isLexingString = false
				lexNextToken(true, lexnexttokenparams{&lineTokens, &token, &prevToken, &prevChar, &char, &isLexingString, &prevTokenIsQuote})
				continue
			}
			//stop lexing key string after reaching end quote
			if char == ':' && prevTokenIsQuote {
				isLexingString = false
				lexNextToken(false, lexnexttokenparams{&lineTokens, &token, &prevToken, &prevChar, &char, &isLexingString, &prevTokenIsQuote})
				continue
			}
			isNegative := prevChar == '-' && unicode.IsNumber(char)
			if !isLexingString && (isNegative || unicode.IsNumber(char)) {
				isLexingNumber = true
				lexNextToken(false, lexnexttokenparams{&lineTokens, &token, &prevToken, &prevChar, &char, &isLexingString, &prevTokenIsQuote})
				continue
			}
			isLexingFloat := char == '.'
			isLexingExponent := isExponent(char)
			isLexingExponentSign := isExponent(prevChar) && (char == '+' || char == '-')
			if isLexingNumber && !isLexingFloat && !isLexingExponent && !isLexingExponentSign && !unicode.IsNumber(char) {
				isLexingNumber = false
				lexNextToken(true, lexnexttokenparams{&lineTokens, &token, &prevToken, &prevChar, &char, &isLexingString, &prevTokenIsQuote})
				continue
			}
			lexNextToken(false, lexnexttokenparams{&lineTokens, &token, &prevToken, &prevChar, &char, &isLexingString, &prevTokenIsQuote})
		}
		// save the last token that was accumulated
		saveToken(&token, &lineTokens, &prevToken)
		tokens = append(tokens, lineTokens)
	}
	return tokens
}

// only save static tokens that are not part of a string.
// updates token and prevChar
func lexNextToken(saveNonStaticToken bool, params lexnexttokenparams) {
	// De Morgan's Law to the rescue. second condition was previously !(token !="\"" && isLexingString && prevTokenIsQuote)
	isStaticToken := slices.Contains(staticTokens, *params.token) && (*params.token == "\"" || !*params.isLexingString || !*params.prevTokenIsQuote)
	if isStaticToken || saveNonStaticToken {
		saveToken(params.token, params.lineTokens, params.prevToken)
	}
	*params.token += string(*params.char)
	*params.prevChar = *params.char
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
		if _, err := parseObject(tokens, &pos, true); err != nil {
			return false, err
		}
		return true, nil
	case LEFTSQUAREBRACE:

		if lastToken != RIGHTSQUAREBRACE {
			return false, parserError(len(tokens)-1, "}", lastToken)
		}
		if _, err := parseArray(tokens, &pos, true); err != nil {
			return false, err
		}
		return true, nil

	}
	nextToken(&pos)
	return false, fmt.Errorf("Invalid JSON string. Expected { or [, got %s", tokens[pos])
}

func parseObject(tokens []string, pos *int, isOuterObject bool) (bool, error) {

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
		default:
			return false, parserError(*pos, "\"", tokens[*pos+1])
		}
		nextToken(pos)
		if tokens[*pos+1] != COLON {
			return false, parserError(*pos, ":", tokens[*pos+1])
		}
		nextToken(pos)
		if _, err := parseValues(tokens, pos); err != nil {
			return false, err
		}
		if ret, err := parseValueEnding(tokens[*pos+1], RIGHTCURLYBRACE, *pos+1, isOuterObject, len(tokens)); ret || err != nil {
			return ret, err
		}
	}

}

// Parses the tokens before a comma in an object or array.
//
// Returns: bool specifying whether to return from calling function and Error value
func parseValueEnding(currentToken string, TOKEN string, pos int, isParent bool, tokensLength int) (bool, error) {

	if currentToken != COMMA {
		if currentToken != TOKEN {
			return false, parserError(pos, TOKEN, currentToken)
		}
		if isParent && pos != tokensLength-1 {
			return false, parserError(pos, "EOF", currentToken)
		}
		return true, nil // at this point we want to stop parsing the object or array
	}
	return false, nil
}
func parseArray(tokens []string, pos *int, isOuterArray bool) (bool, error) {

	for {
		nextToken(pos)

		if tokens[*pos+1] == RIGHTSQUAREBRACE {
			if matchComma(tokens[*pos]) {
				return false, parserError(*pos, "token", "]")
			}

			return true, nil
		}

		if _, err := parseValues(tokens, pos); err != nil {
			return false, err
		}
		if ret, err := parseValueEnding(tokens[*pos+1], RIGHTSQUAREBRACE, *pos+1, isOuterArray, len(tokens)); ret || err != nil {
			return ret, err
		}
	}
}

// Parse out a string,object,number, or array
func parseValues(tokens []string, pos *int) (bool, error) {
	switch tokens[*pos+1] {
	case LEFTCURLYBRACE:
		if _, err := parseObject(tokens, pos, false); err != nil {
			return false, err
		}
	case LEFTSQUAREBRACE:
		if _, err := parseArray(tokens, pos, false); err != nil {
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
