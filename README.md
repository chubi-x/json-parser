# JSON Parser CLI

> [!NOTE]
> I built this as an exercise in learning how to program with Go so please excuse the shabiness!

## Installation

You can build from source or download a binary from the [releases page](https://github.com/chubi-x/json-parser/releases/tag/v1). To build from source simply clone the repo and run `go build main.go`

## Usage

The program accepts json input from a file or from stdin. You can specify a file like so:
`./jsonparse --file <path to file>`
You can also pass in input by piping data, for example `cat test.json | ./jsonparse` or simply running `./jsonparse` and typing in the json input.
