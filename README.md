# sudoku
Sudoku solver uService written in Go.  Accessible via REST API.

Currently runs as a local process hard-coded to listen on
localhost:8000/sudoku/solve

Accepts the http Post verb with the following Go/JSON data
structure in the body:

type JsonGrid struct {
	Solution Grid
	Status   string
}

Where type Grid is a 9x9 array of uint8 values with 0
representing a blank cel.

The service will populate the Status field with a status string.
If a solution is possible, the Solution grid will contain a solved puzzle.`

