package board

import (
	"fmt"
	"testing"
)

func TestCreateBoard(t *testing.T) {
	brd := make([][]Status, 5) // Make outer slice of size == width
	// Make inner slice of size == Height
	for i := range brd {
		brd[i] = make([]Status, 10)
	}
	// Assign each cell the status "Blank"
	for i := range brd {
		for j := range brd[i] {
			brd[i][j] = Blank
		}
	}

	expectedObject := Board{
		height:     10,
		width:      5,
		PtrToBoard: &brd,
	}

	outputObject := *CreateBoard(10, 5)

	if expectedObject.height != outputObject.height || expectedObject.width != outputObject.width {
		t.Errorf("The outputObject : %+v != %+v : The expectedObject \n", outputObject, expectedObject)
		t.Errorf("The ptr to board for outputObject contains : %+v \n", (*outputObject.PtrToBoard))
		t.Errorf("The ptr to board for expectedObject contains : %+v \n", (*expectedObject.PtrToBoard))

	} else {
		fmt.Println("Both the outputObject and expectedObject are the same.")
		fmt.Printf("The ptr to board for outputObject contains : %+v \n", (*outputObject.PtrToBoard))
	}
}

func TestGetBoardHeight(t *testing.T) {
	brd := Board{
		height: 100,
	}

	var expectedHeight uint8 = 100

	actualHeight := brd.GetBoardHeight()

	if expectedHeight != actualHeight {
		t.Errorf("The expectedHeight : %d != %d : The actualHeight \n", expectedHeight, actualHeight)
	} else {
		fmt.Printf("The expectedHeight : %d == %d : The actualHeight \n", expectedHeight, actualHeight)
		fmt.Println("Both the expectedHeight and actualHeight are the same.")
	}
}

func TestGetBoardWidth(t *testing.T) {
	brd := Board{
		width: 50,
	}

	var expectedWidth uint8 = 50

	actualWidth := brd.GetBoardWidth()

	if expectedWidth != actualWidth {
		t.Errorf("The expectedWidth : %d != %d : The actualWidth \n", expectedWidth, actualWidth)
	} else {
		fmt.Printf("The expectedWidth : %d == %d : The actualWidth \n", expectedWidth, actualWidth)
		fmt.Println("Both the expectedWidth and actualWidth are the same.")
	}
}

func TestPrintBoard(t *testing.T) {
	var flag uint8 = 0
	// Creating our test board
	expectedBoard := make([][]Status, 5)
	for i := range expectedBoard {
		expectedBoard[i] = make([]Status, 10)
	}
	// Setting Status for the test board
	for i := range expectedBoard {
		for j := range expectedBoard[i] {
			expectedBoard[i][j] = Blank
		}
	}

	outputBoard := *CreateBoard(5, 10)

	for i := 0; i < 5; i++ {
		for j := 0; j < 10; j++ {
			if expectedBoard[i][j] != (*outputBoard.PtrToBoard)[i][j] {
				flag = 1
				break
			}
		}
	}

	if flag != 0 {
		t.Errorf("The Expected Board and Output Board aren't the same.")
		t.Errorf("Expected Board is : %+v \n", (expectedBoard))
		t.Errorf("Output Board is : %+v \n", (*outputBoard.PtrToBoard))
	} else {
		fmt.Println("Both the boards are same.")
		fmt.Printf("Expected Board is : %+v \n", (expectedBoard))
		fmt.Printf("Output Board is : %+v \n", (*outputBoard.PtrToBoard))
	}
}

func TestGetStatus(t *testing.T) {
	// Creating a 2D array and assigning every cell of it the status "Blank"
	arr := make([][]Status, 20)
	for i := range arr {
		arr[i] = make([]Status, 40)
	}
	for i := range arr {
		for j := range arr[i] {
			arr[i][j] = Blank
		}
	}

	brd := Board{
		height:     20,
		width:      40,
		PtrToBoard: &arr,
	}

	if brd.GetStatus(17, 31) != Blank {
		t.Errorf("The status returned is not the expected status.")
		t.Errorf("Resultant Status : %+v \n", brd.GetStatus(17, 31))
		t.Errorf("Expected Status : %+v \n", Blank)
	} else {
		fmt.Println("Both the Status are same.")
		fmt.Printf("Resultant Status : %+v \n", brd.GetStatus(17, 31))
		fmt.Printf("Expected Status : %+v \n", Blank)
	}
}

func TestSetStatus(t *testing.T) {
	// Creating a 2D array and assigning every cell of it the status "Blank"

	arr := make([][]Status, 20)
	for i := range arr {
		arr[i] = make([]Status, 40)
	}
	for i := range arr {
		for j := range arr[i] {
			arr[i][j] = Blank
		}
	}

	brd := Board{
		height:     20,
		width:      40,
		PtrToBoard: &arr,
	}

	brd.SetStatus(Hit, 19, 39)

	if brd.GetStatus(19, 39) != Hit {
		t.Errorf("The status set is not the expected status.")
		t.Errorf("Resultant Status : %+v \n", brd.GetStatus(19, 39))
		t.Errorf("Expected Status : %+v \n", Hit)
	} else {
		fmt.Println("Both the Status are same.")
		fmt.Printf("Resultant Status : %+v \n", brd.GetStatus(19, 39))
		fmt.Printf("Expected Status : %+v \n", Hit)
	}
}
