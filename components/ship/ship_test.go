package ship

import (
	"battleship/components/board"
	"fmt"
	"testing"
)

func TestGenerateRandomShip(t *testing.T) {
	var flag uint8 = 0

	b := board.CreateBoard(10, 10)

	GenerateRandomShips(1, b)

	for sz, ship := range AllShips {
		for j := 0; j < int(sz); j++ {
			if (*b.PtrToBoard)[ship[j].y][ship[j].x] != board.Ship {
				flag = 1
				break
			}
		}

	}

	if flag == 1 {
		t.Errorf("Ships were not created properly.")
	}

	fmt.Println(AllShips)

	b.PrintBoardDev()

}
