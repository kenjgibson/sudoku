// Simple "go test" program to test the Sudoku module through direct calls (no http in the path)

package sudoku

import (
	"fmt"
	"testing"
)

const horzLine = "\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500"

//  Create some simple test games to test various scenarios

//  Puzzle with out-of-range value
var ooRangeGrid = [9][9]CelVal{
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
var illegalGrid = [9][9]CelVal{
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
var easyGrid = [9][9]CelVal{
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
var medGrid = [9][9]CelVal{
	{0, 0, 0, 0, 5, 1, 0, 0, 0},
	{5, 6, 1, 9, 0, 0, 0, 0, 0},
	{4, 0, 0, 7, 0, 0, 0, 0, 0},
	{0, 0, 2, 0, 0, 5, 4, 0, 0},
	{0, 4, 5, 0, 0, 0, 0, 0, 8},
	{1, 9, 0, 0, 4, 0, 0, 0, 3},
	{0, 8, 0, 0, 2, 7, 0, 3, 1},
	{6, 0, 0, 0, 0, 0, 0, 2, 0},
	{0, 5, 0, 8, 0, 0, 6, 4, 9}}

var hardGrid = [9][9]CelVal{
	{3, 0, 5, 0, 7, 1, 0, 0, 9},
	{0, 0, 0, 3, 4, 0, 0, 0, 0},
	{0, 9, 0, 2, 0, 0, 0, 0, 0},
	{0, 3, 0, 0, 0, 4, 0, 0, 0},
	{0, 6, 0, 0, 0, 0, 0, 0, 7},
	{0, 0, 0, 0, 0, 2, 8, 5, 0},
	{0, 0, 0, 0, 0, 0, 0, 8, 0},
	{0, 5, 4, 0, 0, 0, 9, 0, 1},
	{0, 0, 7, 0, 0, 0, 4, 0, 0}}

func printGrid(gp *Grid) {

	for row := 0; row < GridSize; row++ {
		if row%3 == 0 {
			fmt.Printf("%s\n", horzLine)
		}
		for col := 0; col < GridSize; col++ {
			if col%3 == 0 {
				fmt.Printf("|")
			}
			fmt.Printf("%d ", gp[row][col])
		}
		fmt.Printf("|\n")
	}
	fmt.Printf("%s\n", horzLine)
}

//  Populate the JsonGrid var with initializer values
func copyGrid(src [9][9]CelVal, dst *Grid) {

	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			dst[row][col] = src[row][col]
		}
	}
}

func TestJsonOoRange(t *testing.T) {

	var testGrid JsonGrid

	testGrid.Solution = ooRangeGrid
//	copyGrid(ooRangeGrid, &testGrid.Solution)

	Jsolve(&testGrid)

	if testGrid.Status == "Success" {
		errString := fmt.Sprintf("\nFailed to catch ooRange.  Returned: %s", testGrid.Status)
		t.Error(errString)
	} else {
		fmt.Printf("\nCaught ooRange: %s\n", testGrid.Status)
	}
}

func TestIllegal(t *testing.T) {

	var testGrid JsonGrid

	testGrid.Solution = illegalGrid
	// copyGrid(illegalGrid, &testGrid.Solution)

	Jsolve(&testGrid)

	if testGrid.Status == "Success" {
		errString := fmt.Sprintf("Failed to catch illegal game config: %s", testGrid.Status)
		t.Error(errString)
	} else {
		fmt.Printf("Caught illegal config: %s\n", testGrid.Status)
	}
}

func TestEasy(t *testing.T) {

	var testGrid JsonGrid

	testGrid.Solution = easyGrid
	// copyGrid(easyGrid, &testGrid.Solution)

	Jsolve(&testGrid)

	if testGrid.Status != "Success" {
		errString := fmt.Sprintf("Failed easy puzzle.  Returned: %s", testGrid.Status)
		t.Error(errString)
	} else {
		fmt.Printf("\nSuccess from easy puzzle:\n")
		printGrid(&testGrid.Solution)
	}
}

func TestHard(t *testing.T) {

	var testGrid JsonGrid

	testGrid.Solution = hardGrid
	// copyGrid(hardGrid, &testGrid.Solution)

	Jsolve(&testGrid)

	if testGrid.Status != "Success" {
		errString := fmt.Sprintf("Failed hard puzzle.  Returned:: %s", testGrid.Status)
		t.Error(errString)
	} else {
		fmt.Printf("\nSuccess from hard puzzle\n")
		printGrid(&testGrid.Solution)
	}

}
