package player

import (
	"testing"
)

func TestCreatePlayer(t *testing.T) {
	p_test := Player{
		name:           "Aditya",
		noOfAttempts: 0,
	}

	p := CreatePlayer("Aditya")

	if p_test != *p {
		t.Errorf("Error creating player object")
	}
}

func TestGetNoOfAttempts(t *testing.T) {
	p_test := Player{
		name:           "Aditya",
		noOfAttempts: 5,
	}

	val := p_test.GetNoOfAttempts()

	if val != 5 {
		t.Errorf("Coudn't properly get no. of attempts for the test player.")
	}

}

func TestIncreaseNoOfAttempts(t *testing.T) {
	p_test := Player{
		name:           "Aditya",
		noOfAttempts: 9,
	}

	p_test.IncreaseNoOfAttempts()

	if p_test.GetNoOfAttempts() != 10 {

		t.Errorf("Coudn't increase the no. of attempts for the test player.")
	}

}
