// This package will have information about our board and it's functions
package board

import "fmt"

// Structure of the Board
type Board struct {
	height     uint8
	width      uint8
	PtrToBoard *([][]Status)
}

// Status variable will define at any instance of time the status of current cell
type Status string

// Enum values taken by status
const (
	Miss  Status = "^"
	Hit   Status = "@"
	Blank Status = "-"
	Ship  Status = "S"
)

// Creates a new Board object
func CreateBoard(ht, wd uint8) *Board {

	brd := make([][]Status, ht)
	// Now we create the subslices inside each index of the main (length) slice.
	for i := range brd {
		brd[i] = make([]Status, wd)
	}

	// Now we initialize each cell of our board with "Blank" status
	for i := range brd {
		for j := range brd[i] {
			brd[i][j] = Blank
		}
	}

	return &Board{
		height:     ht,
		width:      wd,
		PtrToBoard: &brd,
	}
}

// Returns the length of the board object
func (b *Board) GetBoardHeight() uint8 {
	return b.height
}

// Returns the breadth of the board object
func (b *Board) GetBoardWidth() uint8 {
	return b.width
}

var i, j uint8

// Prints the board for Developer
func (b *Board) PrintBoardDev() {

	fmt.Printf("\nAt present the warfield looks like :\n\n")

	for i = 0; i < b.GetBoardHeight(); i++ {
		fmt.Printf("%d\t", i)
		for j = 0; j < b.GetBoardWidth(); j++ {
			fmt.Print((*b.PtrToBoard)[i][j], "\t")
			// fmt.Print("(", i, j, ")", "\t")
		}
		fmt.Println()
	}

	fmt.Print("\n\n")

	for j = 0; j < b.width; j++ { // Printing indices along the breadth
		fmt.Printf("\t%d", j)
	}

	fmt.Printf("\n\n")
}

// Gets the status of the ( dx, dy)th cell of our board
func (b *Board) GetStatus(dx, dy uint8) Status {
	return (*b.PtrToBoard)[dx][dy]
}

// Sets the status of the ( dx, dy)th cell of our board
func (b *Board) SetStatus(newStat Status, dx, dy uint8) {
	(*b.PtrToBoard)[dx][dy] = newStat
}
