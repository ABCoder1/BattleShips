package gameService

import (
	"battleship/components/board"
	"battleship/components/player"
	"battleship/components/ship"
	"fmt"
)

type GameStatus uint8

const (
	Continue GameStatus = 1
	End      GameStatus = 0
)

type Cordinate struct {
	latitude  uint8
	longitude uint8
}

var printMap = make(map[Cordinate]bool)

var shipHitCount uint8 = 0

func NewGameService(name string, latitude, longitude uint8, numberOfShips int) (*player.Player, *board.Board) {
	p1 := player.CreatePlayer(name)

	fmt.Println("Player created successfully")

	gameBoard := board.CreateBoard(latitude, longitude)

	fmt.Println("Board created successfully")

	ship.GenerateRandomShips(numberOfShips, gameBoard)

	fmt.Println("Ships created successfully")

	PrintBoard(gameBoard)

	return p1, gameBoard
}

func GamePlay(h, w uint8, p *player.Player, b *board.Board, numberOfShips int) GameStatus {

	p.IncreaseNoOfAttempts()
	if (h > b.GetBoardHeight()-1) || (w > b.GetBoardWidth()-1) { // Removed the check for h < 0 & w < 0 as uint8 cannot be negative
		fmt.Println("\nIndices out of bounds. Try again...")
	} else if b.GetStatus(h, w) == board.Ship {
		printMap[Cordinate{h, w}] = true
		shipHitCount++
		b.SetStatus(board.Hit, h, w)
	} else if b.GetStatus(h, w) == board.Blank {
		printMap[Cordinate{h, w}] = true

		b.SetStatus(board.Miss, h, w)
	} else if b.GetStatus(h, w) == board.Miss {
		fmt.Print("\nYou have already missed at this place...")
	} else if b.GetStatus(h, w) == board.Hit {
		fmt.Println("\nYou have already hit at this place...")
	}

	PrintBoard(b)
	if shipHitCount == uint8(numberOfShips*(numberOfShips+1))/2 {
		return End
	} else {
		return Continue
	}
}

var i, j uint8

// Prints the board for player
func PrintBoard(b *board.Board) {
	fmt.Print("\nAt present the warfield looks like :\n\n")

	for i = 0; i < b.GetBoardHeight(); i++ {
		fmt.Printf("%d\t", i)

		for j = 0; j < b.GetBoardWidth(); j++ {
			if printMap[Cordinate{i, j}] == true {
				fmt.Print(b.GetStatus(i, j), "\t")
			} else {
				fmt.Print(board.Blank, "\t")

			}
		}
		fmt.Println()
	}

	fmt.Print("\n\n")

	for j = 0; j < b.GetBoardWidth(); j++ { // Printing indices along the breadth
		fmt.Printf("\t%d", j)
	}

	fmt.Print("\n\n", ship.AllShips, "\n")

	ship.IsAnyShipDestroyed(b)

	fmt.Printf("\n\n")

}
