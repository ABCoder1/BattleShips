// This file will contain info about ship structure and all it's functions
package ship

import (
	"battleship/components/board"
	"fmt"
	"math/rand"
)

type Coordinates struct {
	x uint8
	y uint8
}

type orientation uint8

const (
	Horizontal orientation = iota
	Vertical
)

// Map to store coordinates of each ship
var AllShips = make(map[int][]Coordinates) // map[size][(),(),()...Coordinates]

func GenerateRandomShips(numberOfShips int, b *board.Board) {
	for size := 1; size <= numberOfShips; size++ {
		orient := orientation(rand.Intn(2))
		AllShips[size] = createShipInstance(int(orient), uint8(size), b)
	}
}

func createShipInstance(orient int, size uint8, b *board.Board) []Coordinates {
	var i int
	var xCoordinate, yCoordinate int
	ship := make([]Coordinates, size)

	switch orient {
	case 0: // Horizontally placed ship
		for {
			xCoordinate, yCoordinate = getRandomCoordinates()
			if (xCoordinate+int(size)) < int(b.GetBoardWidth()) && (yCoordinate < int(b.GetBoardHeight())) { // Checking if ship's indices go out of bounds
				noCollisions := true
				for i = range int(size) {
					if (*b.PtrToBoard)[yCoordinate][xCoordinate+i] == board.Ship { // Checking if any indices of our chosen slice is already occupied by another ship.
						noCollisions = false
						break
					}
				}

				if noCollisions {
					break
				}
			}
		}

		for i = range int(size) {
			(*b.PtrToBoard)[yCoordinate][xCoordinate+i] = board.Ship

			ship[i] = Coordinates{
				x: uint8(xCoordinate + i),
				y: uint8(yCoordinate),
			}
		}

	case 1: // Vertically placed ship
		for {
			xCoordinate, yCoordinate = getRandomCoordinates()
			if (yCoordinate+int(size)) < int(b.GetBoardHeight()) && (xCoordinate < int(b.GetBoardWidth())) { // Checking if ship's indices go out of bounds
				noCollisions := true
				for i := range int(size) {
					if (*b.PtrToBoard)[yCoordinate+i][xCoordinate] == board.Ship { // Checking if any indices of our chosen slice is already occupied by another ship.
						noCollisions = false
						break
					}
				}

				if noCollisions {
					break
				}
			}
		}

		for i = range int(size) {
			(*b.PtrToBoard)[yCoordinate+i][xCoordinate] = board.Ship

			ship[i] = Coordinates{
				x: uint8(xCoordinate),
				y: uint8(yCoordinate + i),
			}
		}

	}

	return ship
}

func getRandomCoordinates() (int, int) {
	x := rand.Intn(256)
	y := rand.Intn(256)

	return x, y
}

func IsAnyShipDestroyed(b *board.Board) {
	var destroyed bool
	for shipSize, shipCoordinates := range AllShips {
		destroyed = true
		for j := range shipSize {
			if b.GetStatus(shipCoordinates[j].y, shipCoordinates[j].x) != board.Hit {
				destroyed = false
			}
		}
		if destroyed == true {
			fmt.Print("\n\n", "Ship ", shipSize, " is completely destroyed", "\n")
			delete(AllShips, shipSize)
		}
	}
}
