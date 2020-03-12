# sudoku
Learn golang, solve sudoku

[![Go Report Card](https://goreportcard.com/badge/github.com/nanw1103/sudoku)](https://goreportcard.com/report/github.com/nanw1103/sudoku)

[![Build Status](https://api.travis-ci.org/nanw1103/sudoku.svg?branch=master)](https://api.travis-ci.org/nanw1103/sudoku.svg?branch=master)

[![godoc](https://godoc.org/github.com/nanw1103/sudoku?status.svg)](https://godoc.org/github.com/nanw1103/sudoku)

Standard sudoku solution should be Exact Cover problem.

To learn go, the implemented algorithm adopts a straightforward backtracing with some prune and bitset tricks.


```
go build
sudoku [fileName]
```
or
```
go run sudoku.go main.go [fileName]
```
