// Simple "go test" program to test the REST API interface
// to the Sudoku solver web service
//
// Currently assumes the main server has been started as a separate process
// on the local machine listening on port 8000

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kenjgibson/sudoku/sudoku"
	"net/http"
	"testing"
)

const targetURL = "http://localhost:8000/sudoku/solve"
const contType = "application/json"
const horzLine = "\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500"

//  Hard-code some test games

//  Puzzle with out-of-range value
var ooRangeGrid = [9][9]sudoku.CelVal{
	{0, 0, 9, 0, 0, 3, 0, 0, 0},
	{0, 0, 0, 6, 2, 0, 9, 0, 4},
	{8, 2, 7, 0, 0, 0, 6, 0, 3},
	{2, 1, 0, 3, 6, 0, 0, 4, 5},
	{0, 9, 6, 25, 7, 0, 0, 0, 0},
	{7, 0, 0, 0, 4, 0, 1, 9, 0},
	{0, 6, 2, 4, 5, 0, 3, 0, 0},
	{1, 0, 0, 7, 0, 6, 4, 0, 0},
	{3, 0, 0, 9, 8, 2, 0, 6, 0}}

// Puzzle with an illegal initial config
var illegalGrid = [9][9]sudoku.CelVal{
	{0, 0, 9, 0, 0, 3, 0, 0, 0},
	{0, 0, 0, 6, 2, 0, 9, 0, 4},
	{8, 2, 7, 0, 0, 0, 6, 0, 3},
	{2, 1, 0, 3, 6, 0, 0, 4, 5},
	{0, 9, 6, 0, 7, 0, 5, 0, 0},
	{7, 0, 0, 0, 4, 0, 1, 9, 0},
	{0, 6, 2, 4, 5, 0, 3, 0, 0},
	{1, 0, 0, 7, 0, 6, 4, 0, 0},
	{3, 0, 0, 9, 8, 2, 0, 6, 0}}

// Easy one, can be solved through direct calculation
var easyGrid = [9][9]sudoku.CelVal{
	{0, 0, 9, 0, 0, 3, 0, 0, 0},
	{0, 0, 0, 6, 2, 0, 9, 0, 4},
	{8, 2, 7, 0, 0, 0, 6, 0, 3},
	{2, 1, 0, 3, 6, 0, 0, 4, 5},
	{0, 9, 6, 0, 7, 0, 0, 0, 0},
	{7, 0, 0, 0, 4, 0, 1, 9, 0},
	{0, 6, 2, 4, 5, 0, 3, 0, 0},
	{1, 0, 0, 7, 0, 6, 4, 0, 0},
	{3, 0, 0, 9, 8, 2, 0, 6, 0}}

// Harder ones that require recursion
var medGrid = [9][9]sudoku.CelVal{
	{0, 0, 0, 0, 5, 1, 0, 0, 0},
	{5, 6, 1, 9, 0, 0, 0, 0, 0},
	{4, 0, 0, 7, 0, 0, 0, 0, 0},
	{0, 0, 2, 0, 0, 5, 4, 0, 0},
	{0, 4, 5, 0, 0, 0, 0, 0, 8},
	{1, 9, 0, 0, 4, 0, 0, 0, 3},
	{0, 8, 0, 0, 2, 7, 0, 3, 1},
	{6, 0, 0, 0, 0, 0, 0, 2, 0},
	{0, 5, 0, 8, 0, 0, 6, 4, 9}}

var hardGrid = [9][9]sudoku.CelVal{
	{3, 0, 5, 0, 7, 1, 0, 0, 9},
	{0, 0, 0, 3, 4, 0, 0, 0, 0},
	{0, 9, 0, 2, 0, 0, 0, 0, 0},
	{0, 3, 0, 0, 0, 4, 0, 0, 0},
	{0, 6, 0, 0, 0, 0, 0, 0, 7},
	{0, 0, 0, 0, 0, 2, 8, 5, 0},
	{0, 0, 0, 0, 0, 0, 0, 8, 0},
	{0, 5, 4, 0, 0, 0, 9, 0, 1},
	{0, 0, 7, 0, 0, 0, 4, 0, 0}}

func printGrid(gp *sudoku.Grid) {

	for row := 0; row < sudoku.GridSize; row++ {
		if row%3 == 0 {
			fmt.Printf("%s\n", horzLine)
		}
		for col := 0; col < sudoku.GridSize; col++ {
			if col%3 == 0 {
				fmt.Printf("|")
			}
			fmt.Printf("%d ", gp[row][col])
		}
		fmt.Printf("|\n")
	}
	fmt.Printf("%s\n", horzLine)
}

//
//  Internal function to handle marshalling to/from json and
//  sending/receiving the Post message
//
func doPost(testGrid *sudoku.JsonGrid) error {

	// Marshal the puzzle to solve into a json byte array
	jData, err := json.Marshal(testGrid)
	if err != nil {
		err = fmt.Errorf("Marshal grid failed: %s", err)
		return err
	}

	// Send the Post message with the json-encoded array in the body
	// Marshal returns a Byte array but http functions require a
	// Buffer object that references the underlying Byte array
	resp, err := http.Post(targetURL, contType, bytes.NewBuffer(jData))
	if err != nil {
		err = fmt.Errorf("Error sending Post: %s", err)
		return err
	}

	//  Body needs to get closed on all return scenarios
	defer resp.Body.Close()

	// First check status from the server
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Error response from Post: %s", resp.Status)
		return err
	}

	// Finally decode (unmarshal) the resulting puzzle and status
	// from json back to our JsonGrid structure
	if err = json.NewDecoder(resp.Body).Decode(testGrid); err != nil {
		err = fmt.Errorf("json Decode failure: %s", err)
	}
	return err
}

func TestJsonOoRange(t *testing.T) {

	var testGrid sudoku.JsonGrid
	var err error

	testGrid.Solution = ooRangeGrid
	testGrid.Status = ""

	if err = doPost(&testGrid); err != nil {
		err = fmt.Errorf("ooRange test: %s", err)
		t.Error(err)
		return
	}

	if testGrid.Status == "Success" {
		errString := fmt.Sprintf("Failed to catch ooRange.  Returned: %s", testGrid.Status)
		t.Error(errString)
	} else {
		fmt.Printf("\nCaught ooRange: %s\n", testGrid.Status)
	}
}

func TestIllegal(t *testing.T) {

	var testGrid sudoku.JsonGrid
	var err error

	testGrid.Solution = illegalGrid
	testGrid.Status = ""

	if err = doPost(&testGrid); err != nil {
		err = fmt.Errorf("illegal test: %s", err)
		t.Error(err)
		return
	}

	if testGrid.Status == "Success" {
		errString := fmt.Sprintf("Failed to catch illegal game config: %s", testGrid.Status)
		t.Error(errString)
	} else {
		fmt.Printf("\nCaught illegal: %s\n", testGrid.Status)
	}
}

func TestEasy(t *testing.T) {

	var testGrid sudoku.JsonGrid
	var err error

	testGrid.Solution = easyGrid
	testGrid.Status = ""

	if err = doPost(&testGrid); err != nil {
		err = fmt.Errorf("illegal test: %s", err)
		t.Error(err)
		return
	}

	if testGrid.Status != "Success" {
		errString := fmt.Sprintf("Failed easy puzzle.  Returned: %s", testGrid.Status)
		t.Error(errString)
	} else {
		fmt.Printf("\nReturned success solving easy puzzle:\n")
		printGrid(&testGrid.Solution)
	}
}

func TestHard(t *testing.T) {

	var testGrid sudoku.JsonGrid
	var err error

	testGrid.Solution = hardGrid
	testGrid.Status = ""

	if err = doPost(&testGrid); err != nil {
		err = fmt.Errorf("illegal test: %s", err)
		t.Error(err)
		return
	}

	if testGrid.Status != "Success" {
		errString := fmt.Sprintf("Failed hard puzzle.  Returned:: %s", testGrid.Status)
		t.Error(errString)
	} else {
		fmt.Printf("\nReturned success solving hard puzzle:\n")
		printGrid(&testGrid.Solution)
	}

}
