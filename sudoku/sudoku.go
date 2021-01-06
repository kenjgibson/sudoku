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
// Sudoku solver engine package.  This package is fully contained in this file.
// Implements an engine for solving sudoku puzzles.
// Exports the types used to describe a sudoku puzzle, check for a legal config,
// and for solving.  The solver returns a solved puzzle, or error if the puzzle is
// either invalid or unsolvable.
//
// Terminology used:
//	Grid:	The full 9x9 sudoku board
//	Box:	The 3x3 subsections of the grid
//	Cel:	The individual cels that each hold one number
//

package sudoku

import (
	"fmt"
)

// Declare the exported types for describing a Soduku grid
type CelVal uint8 // Type placed in the cels.  Inherits integer operations

const GridSize = 9      // Sudoku grid is a 9x9 array of cels
const MaxVal CelVal = 9 // Unfortunately, Go does not have enums
const MinVal CelVal = 1
const Blank CelVal = 0

// The array describing the sudoku grid for passing to/from the engine
type Grid [GridSize][GridSize]CelVal

// Exported struct for marshaling to/from JSON for communication
// with clients
type JsonGrid struct {
	Solution Grid   `json:"solution"`
	Status   string `json:"status"`
}

//  CelVal method to verify the value is within range
func (val CelVal) IsValid() bool {
	return val <= MaxVal
}

//
//  Internal object describing each cel while iterating on a solution
//  Note compiler init defaults are good so no explicit 'init' required
//
//  A cel can be in one of three states:
//	1. Fixed:  Either pre-set by the game, or only one legal value based on the pre-set game.
//	2. Blank with a list of possible legal candidate values.
//	3. Temp: Temporary candidate value while searching for a valid solution
//
type cel struct {
	fixed  bool     // True if pre-set by caller.  Value cannot be changed
	temp   bool     // Value is a candidate solution for the current iteration
	value  CelVal   // Current value
	opList []CelVal // Track possible solution options for this cel
}

//  Internal array of cels.  Holds the state of the solution while solving
type grid [GridSize][GridSize]cel

// Private cel method for setting to a fixed number
func (cp *cel) setFixed(val CelVal) {
	cp.fixed = true
	cp.temp = false
	cp.value = val
	cp.opList = nil
}

// Re-initialize a non-fixed cel.  Used when backtracking
func (cp *cel) reInit() {
	cp.fixed = false
	cp.value = 0
	cp.temp = false
	cp.opList = nil
}

// Set this cel to 'Temporarily fixed' for the current iteration.
// Temp means it may be cleared when backtracking up to a higher recursion level
func (cp *cel) setTemp(val CelVal) {
	cp.fixed = false
	cp.temp = true
	cp.value = val
	cp.opList = nil
}

func (cp *cel) setValue(val CelVal) {
	cp.value = val
}

func (cp *cel) setOptionList(ol []CelVal) {
	cp.opList = ol
}

// Get methods.  Supposedly not the "Go way" but it's good OO programming

func (cp *cel) isFixed() bool {
	return cp.fixed
}

func (cp *cel) isTemp() bool {
	return cp.temp
}

func (cp *cel) getNumOptions() int {
	return len(cp.opList)
}

func (cp *cel) getOptionList() []CelVal {
	return cp.opList
}

// Internal function for determining if a map contains exactly one of each 9 digits
// Per Sudoku rules, each row, column and 3x3 box must contain one of each digit.

func checkDigits(celMap map[CelVal]bool) bool {
	if len(celMap) != GridSize {
		// Can't contain exactly one of each value
		return false
	}

	for val := MinVal; val <= MaxVal; val++ {
		if !celMap[val] {
			return false
		}
	}
	return true
}

// Private grid methods to check if a row, col, or box already has a given number
// Will use when building up the solution option list for each cel

func (gp *grid) inRow(val CelVal, row int) bool {
	for col := 0; col < GridSize; col++ {
		if val == gp[row][col].value {
			return true
		}
	}
	return false
}

func (gp *grid) inCol(val CelVal, col int) bool {
	for row := 0; row < GridSize; row++ {
		if val == gp[row][col].value {
			return true
		}
	}
	return false
}

func (gp *grid) inBox(val CelVal, row int, col int) bool {
	top := row - (row % 3)
	left := col - (col % 3)

	for c := left; c < (left + 3); c++ {
		for r := top; r < (top + 3); r++ {
			if val == gp[r][c].value {
				return true
			}
		}
	}
	return false
}

// Check if a row, column, or box has a legal config

func (gp *grid) checkRow(row int) bool {
	rm := make(map[CelVal]bool)
	for col := 0; col < GridSize; col++ {
		rm[gp[row][col].value] = true
	}
	return checkDigits(rm)
}

func (gp *grid) checkCol(col int) bool {
	cm := make(map[CelVal]bool)
	for row := 0; row < GridSize; row++ {
		cm[gp[row][col].value] = true
	}
	return checkDigits(cm)
}

//  Assumes passed the top left cel in a box
func (gp *grid) checkBox(topRow int, leftCol int) bool {
	bm := make(map[CelVal]bool)
	for row := topRow; row < topRow+3; row++ {
		for col := leftCol; col < leftCol+3; col++ {
			bm[gp[row][col].value] = true
		}
	}
	return checkDigits(bm)
}

//  Check to determine if a board is solved.

func (gp *grid) checkGrid() bool {
	for i := 0; i < GridSize; i++ {
		if !gp.checkRow(i) {
			return false
		}
		if !gp.checkCol(i) {
			return false
		}
	}
	for row := 0; row < GridSize; row += 3 {
		for col := 0; col < GridSize; col += 3 {
			if !gp.checkBox(row, col) {
				return false
			}
		}
	}
	return true
}

// Build the option list for the cel at the row and column specified
// Returns the number of options found.
// Assumes this cel is not fixed.  Should be confirmed by the caller

func (gp *grid) buildOptionList(row int, col int) int {

	var opList []CelVal

	for val := MinVal; val <= MaxVal; val++ {
		if !gp.inRow(val, row) {
			if !gp.inCol(val, col) {
				if !gp.inBox(val, row, col) {
					opList = append(opList, val)
				}
			}
		}
	}
	gp[row][col].setOptionList(opList)
	return len(opList)
}

// Clear option lists and Temp cels.  Used when backtracking

func (gp *grid) clearTempOptions() {

	for row := 0; row < GridSize; row++ {
		for col := 0; col < GridSize; col++ {
			if !gp[row][col].isFixed() {
				gp[row][col].reInit()
			}
		}
	}
}

//  Find the first/next cel wth the minimum number of solution options.
//  When the program recurses down through the soluiton tree, it will
//  start with this cel as the path with the highest probability of
//  leading to a successful solution

func (gp *grid) findMinOptionCel() (minCelP *cel) {

	var minOptCnt int = int(MaxVal)

	for row := 0; row < GridSize; row++ {
		for col := 0; col < GridSize; col++ {
			if gp[row][col].isFixed() {
				continue
			}
			if gp[row][col].isTemp() {
				continue
			}

			if gp[row][col].getNumOptions() < minOptCnt {
				minCelP = &gp[row][col]
				minOptCnt = gp[row][col].getNumOptions()
			}
		}
	}

	return minCelP
}

//  As a first step in solving a puzzle, solve for cels that only have one legal
//  option based on the client-supplied initial values.
//  Solving a cel may mean additional cels have only one solution so iterate
//  until no more cels can be directly solved.
//
//  The simplest puzzles can be solved through this direct calculation
//  Return true if solved.
//
//  Assumes the client supplied cels are populated and others are initialized to 0
//  Returns an error if any cels have no legal solution options

func (gp *grid) firstPassSolve() (bool, error) {

	var changes bool = true

	for changes {
		changes = false

		for row := 0; row < GridSize; row++ {
			for col := 0; col < GridSize; col++ {
				celPtr := &gp[row][col]

				if celPtr.isFixed() {
					continue
				}

				// for this cel, build its option list.
				switch gp.buildOptionList(row, col) {
				case 0:
					// No options, this is an illegal initial config so return error
					return false, fmt.Errorf("illegal config.  No legal value for cel %d, %d", row, col)
				case 1:
					// If only one option for this cel.  Make it fixed
					val := celPtr.opList[0]
					celPtr.setFixed(val)
					changes = true // Repeat with this cel now solved
				}
			}
			if changes {
				// Restart at cel 0,0 to recalculate new (shorter) option lists
				break
			}
		}
	}

	//  If every cel is fixed, the puzzle is solved
	for row := 0; row < GridSize; row++ {
		for col := 0; col < GridSize; col++ {
			if !gp[row][col].isFixed() {
				return false, nil
			}
		}
	}

	return true, nil // Solved!
}

//  Function to recalculate option lists for all the blank cels after
//  setting a temporary trial value in one cel.

func (gp *grid) recalcOptionLists() bool {

	var recalc bool = true

	for recalc {

		recalc = false

		for row := 0; row < GridSize; row++ {
			for col := 0; col < GridSize; col++ {
				if gp[row][col].isFixed() {
					continue
				}
				if gp[row][col].isTemp() {
					continue
				}

				switch gp.buildOptionList(row, col) {
				case 0:
					// No options for this cel
					// Need to backtrack and try the next temp value
					return false

				case 1:
					// Only one option, set to 'Temp fixed' for the current iteration
					// Since we have set a new value in a cel, recalc all option lists
					val := gp[row][col].opList[0]
					gp[row][col].setTemp(val)
					recalc = true
					break
				}
			}
			if recalc {
				break // Restart at cel 0, 0
			}
		}
	}
	return true
}

//  Takes a Sudoku puzzle that has been initialized with fixed values
//  and initial option lists for each cel.
//  Solves for the cels that have multiple solution options
//
//  The algorithm starts with the first cel with the minimum number
//  of solution options, sets that as fixed with each candidate value,
//  then recursively tries to solve for the remaining cels.  If no
//  solution found, backtracks and tries the next candidate value.
//
//  Returns true of puzzle solved

func (gp *grid) recursiveSolve() bool {

	curCel := gp.findMinOptionCel()
	optionList := curCel.getOptionList()
	if len(optionList) == 0 {
		return gp.checkGrid()
	}

	for _, celVal := range optionList {
		curCel.setFixed(celVal)

		if !gp.recalcOptionLists() {
			// Some cels have no legal option
			// try the next value
			gp.clearTempOptions()
			continue
		}

		// Options lists are recalculated
		// Any cels with only one option are set to Temp Fixed
		// See if we have a solution
		if gp.checkGrid() {
			return true
		}

		// Else, recurse to look for a solution with the current cel fixed
		if gp.recursiveSolve() {
			return true
		}
	}
	//  Tried all options, no solution found.
	//  Re-init this cel and return back to the next higher level
	gp.clearTempOptions()
	curCel.reInit()
	return false
}

//  The public entry point for solving a puzzle.
//  Takes a pointer to a Sudoku grid with initial values.
//  Remaining cels must be blank
//  Returns error for:
//	invalid entry
//	initial config that violates Sudoku rules
//	unsolvable puzzle
//  Otherwise, populates with a solved Grid

func Solve(configP *Grid) error {

	// Allocate a grid structure for maintaining state while solving
	// Note default compiler init values are fine for empty cels
	var solnGrid grid
	var gp *grid = &solnGrid

	// Internal anonymous function to copy the result into the return Grid
	// used in multiple places so declare as anonymous func.
	var cpOut = func(configP *Grid, gp *grid) {
		for row := 0; row < GridSize; row++ {
			for col := 0; col < GridSize; col++ {
				configP[row][col] = gp[row][col].value
			}
		}
	}

	// Initialize.  Return error if any intializers are out of range
	for row := 0; row < GridSize; row++ {
		for col := 0; col < GridSize; col++ {
			if !configP[row][col].IsValid() {
				// Parameter out of range
				err := fmt.Errorf("illegal value for cel %d, %d", row, col)
				return err
			}
			if configP[row][col] == Blank {
				// Compiler init is fine
				continue
			}
			gp[row][col].setFixed(configP[row][col])
		}
	}

	// Verify supplied config meets Sudoku rules.
	// Also set any cels with only one solution option as fixed.
	// Also builds initial option lists for each cel
	solved, err := gp.firstPassSolve()
	if err != nil {
		return err
	}

	// The simplest puzzles can be solved above.
	if solved {
		cpOut(configP, gp)
		return nil
	}

	if !gp.recursiveSolve() {
		err := fmt.Errorf("No solution found.")
		return err
	}

	// Copy the solution and return solution to caller
	cpOut(configP, gp)
	return nil
}

// JSON Solve
// Takes a Json Grid structure.
// Calls Solve and copies status into the JsonGrid struct
// If a solution is found, Solve copies directly into the
// Solution grid in the JsonGrid struct

func Jsolve(jGridP *JsonGrid) {

	if err := Solve(&jGridP.Solution); err != nil {
		jGridP.Status = fmt.Sprintf("%v", err)
	} else {
		jGridP.Status = fmt.Sprintf("Success")
	}
	return
}
