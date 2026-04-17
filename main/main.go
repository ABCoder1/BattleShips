// This will act as caller and printer to our game
package main

import (
	"battleship/services/gameService"
	"fmt"
)

func main() {
	fmt.Println("Welcome to Battle Ship")
	fmt.Println()

	var name string
	fmt.Print("Please enter the name of the player : ")
	fmt.Scanln(&name)
	fmt.Println()

	var numberOfShips int
	fmt.Print("Please enter the number of ships (min: 1, max: 15) : ")
	fmt.Scanln(&numberOfShips)
	fmt.Println()

	var latitude uint8
	fmt.Print("Please enter the height of the board (min: number of Ships + 1, max: 255) : ")
	fmt.Scanln(&latitude)
	fmt.Println()

	var longitude uint8
	fmt.Print("Please enter the width of the board (min: number of Ships + 1, max: 255) : ")
	fmt.Scanln(&longitude)
	fmt.Println()

	player1, gameBoard := gameService.NewGameService(name, latitude, longitude, numberOfShips)

	var yCoordinate int
	var xCoordinate int
	for { // Ask the player where he wants to attack untill all the ships are destroyed...

		fmt.Println("Where do you want to attack : ")

		fmt.Println("Enter height coord:")
		fmt.Scanln(&yCoordinate)
		if yCoordinate < 0 || yCoordinate > 255 {
			fmt.Println("Invalid Input")
			continue
		}
		yCoordinate = yCoordinate % 256 // To ensure that the input is always between 0 and 255

		fmt.Println("Enter width coord:")
		fmt.Scanln(&xCoordinate)
		if xCoordinate < 0 || xCoordinate > 255 {
			fmt.Println("Invalid Input")
			continue
		}
		xCoordinate = xCoordinate % 256 // To ensure that the input is always between 0 and 255

		s := gameService.GamePlay(uint8(yCoordinate), uint8(xCoordinate), player1, gameBoard, numberOfShips)
		if s == gameService.End {
			break
		}
	}

	fmt.Println("Congratulations")
	fmt.Printf("You finished the game in : %d turns\n", player1.GetNoOfAttempts())
}
