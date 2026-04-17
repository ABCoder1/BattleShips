// This package contains info about player and it's functions
package player

type Player struct {
	name         string
	noOfAttempts uint16
}

func CreatePlayer(name string) *Player {

	return &Player{
		name:         name,
		noOfAttempts: 0, // Added for readability, but Golang initializes it to 0 by default
	}
}

func (p *Player) GetNoOfAttempts() uint16 {
	return p.noOfAttempts
}

func (p *Player) IncreaseNoOfAttempts() {
	p.noOfAttempts++
}
