# JSON Parser CLI

## Installation

You can build from source or download a binary from the [releases page](https://github.com/chubi-x/json-parser/releases/tag/v1). To build from source simply clone the repo and run `go build main.go`

## Usage

The program accepts json input from a file, from stdin, or as a string of text. You can specify a file like so:
`./jsonparse --file <path to file>`
You can also pass in input by piping data, for example `cat test.json | ./jsonparse` or simply running `./jsonparse` and typing in the json input.
If pasing in a json string directly you can do so like: `./jsonparse {"key":"value"}`
