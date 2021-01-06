//
// Copyright 2020, 2021 Kenneth J. Gibson
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

//
// Main function for cloud-hosted sudoku solver Microservice.  Accepts
// a JSON-encoded sudoku grid (Go JsonGrid struct) via an http Post message,
// verifies a valid puzzle and returns a solved JsonGrid struct, if solvable.
// Otherwise returns error.
//

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"github.com/kenjgibson/sudoku/sudoku"
)

var getString = `Sudoku Solver API.

Invoke at this endpoint using POST, Content-Type application/json,
and with body containing the following Go/JSON struct representing
the Sudoku game to solve:

type JsonGrid struct {
	Solution Grid
	Status   string
}

Where type Grid is a 9x9 array of uint8 values with 0 representing a blank cel.

The service will populate the Status field with a status string.  If a solution
is possible, the Solution grid will contain a solved Sudoku puzzle.`

func main() {
	http.HandleFunc("/sudoku/solve", solver)
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func solver(respP http.ResponseWriter, reqP *http.Request) {

	//  Post is the recommended Method for invoking a uService
	//  Reespond to Get with a short message.  Otherwise, reject
	//  Any other method.

	switch reqP.Method {
	case http.MethodGet:
		fmt.Fprintf(respP, "%s\n", getString)
		return

	case http.MethodPost:
		var jGrid sudoku.JsonGrid

		decoder := json.NewDecoder(reqP.Body)
		if err := decoder.Decode(&jGrid); err != nil {
			err = fmt.Errorf("Can't decode JSON: %s", err)
			log.Printf("%v", err)
			respP.WriteHeader(http.StatusBadRequest)
			respP.Write([]byte("400 - Bad Request"))
			return
		}

		defer reqP.Body.Close()

		sudoku.Jsolve(&jGrid)

		encoder := json.NewEncoder(respP)
		if err := encoder.Encode(jGrid); err != nil {
			err = fmt.Errorf("Can't encode: %s", err)
			log.Printf("%v", err)
		}
		return

	default:
		// Some other unsupported http verb
		respP.WriteHeader(http.StatusMethodNotAllowed)
		respP.Write([]byte("405 - Method Not Allowed\n"))
	}
}
