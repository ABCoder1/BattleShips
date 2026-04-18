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

type RenderCoordinate struct {
	X uint8
	Y uint8
}

type RenderedShip struct {
	Size       int
	IsVertical bool
	Coords     []RenderCoordinate
}

type orientation uint8

const (
	Horizontal orientation = iota
	Vertical
)

// Map to store coordinates of each ship
var AllShips = make(map[int][]Coordinates) // map[size][(),(),()...Coordinates]

// NOTE : Persistent snapshot captured at generation time. This is not mutated when
// ships are destroyed, so it can be used to render the full fleet layout on the post-game reveal screen.
var renderedShipsSnapshot []RenderedShip

// NOTE : Reusable scratch buffer for candidate positions. Reusing this across
// placements avoids repeated heap allocations during ship generation, which
// used to spike memory when many games were played back-to-back or when
// generating large fleets.
var candidateScratch []placement

type placement struct {
	x, y       uint8
	isVertical bool
}

func ResetRenderedShips() {
	// NOTE : Truncate rather than reassign so the underlying array can be reused by
	// the next game, reducing allocations handed to the GC.
	renderedShipsSnapshot = renderedShipsSnapshot[:0]
}

func SnapshotRenderedShips() []RenderedShip {
	out := make([]RenderedShip, len(renderedShipsSnapshot))
	for i, s := range renderedShipsSnapshot {
		copied := make([]RenderCoordinate, len(s.Coords))
		copy(copied, s.Coords)
		out[i] = RenderedShip{
			Size:       s.Size,
			IsVertical: s.IsVertical,
			Coords:     copied,
		}
	}
	return out
}

func SnapshotAllShips() map[int][]RenderCoordinate {
	out := make(map[int][]RenderCoordinate, len(AllShips))
	for size, coords := range AllShips {
		copied := make([]RenderCoordinate, len(coords))
		for i, c := range coords {
			copied[i] = RenderCoordinate{X: c.x, Y: c.y}
		}
		out[size] = copied
	}
	return out
}

// NOTE : GenerateRandomShips places `numberOfShips` ships (sizes 1..numberOfShips) on
// the supplied board. The placement algorithm enumerates all legal positions
// for each ship size then picks one uniformly at random. This is bounded
// (O(width*height*size) per ship) and always terminates, which is critical
// because earlier versions could spin indefinitely for infeasible configurations
// (e.g. many large ships on a small grid), leading to unbounded memory growth
// from stalled frames and never-freed allocations.
func GenerateRandomShips(numberOfShips int, b *board.Board) {
	renderedShipsSnapshot = renderedShipsSnapshot[:0]
	for k := range AllShips {
		delete(AllShips, k)
	}

	if b == nil || numberOfShips <= 0 {
		return
	}

	for size := 1; size <= numberOfShips; size++ {
		coords, isVertical, ok := tryPlaceShip(uint8(size), b)
		if !ok {
			// NOTE : Infeasible layout for the remaining ships. Abort rather than
			// spin forever; upstream code still renders whatever has been placed so far.
			fmt.Printf("ship.GenerateRandomShips: no legal placement for size %d on %dx%d board\n", size, b.GetBoardWidth(), b.GetBoardHeight())
			return
		}

		AllShips[size] = coords

		rendered := make([]RenderCoordinate, len(coords))
		for i, c := range coords {
			rendered[i] = RenderCoordinate{X: c.x, Y: c.y}
		}
		renderedShipsSnapshot = append(renderedShipsSnapshot, RenderedShip{
			Size:       size,
			IsVertical: isVertical,
			Coords:     rendered,
		})
	}
}

// NOTE : tryPlaceShip enumerates every legal position (both orientations) for a ship
// of the given size on the current board, then chooses one uniformly at random.
// Returns the placed coordinates, its orientation, and true on success.
func tryPlaceShip(size uint8, b *board.Board) ([]Coordinates, bool, bool) {
	width := b.GetBoardWidth()
	height := b.GetBoardHeight()
	grid := *b.PtrToBoard

	candidateScratch = candidateScratch[:0]

	if size <= width {
		limitX := width - size
		for y := uint8(0); y < height; y++ {
			for x := uint8(0); x <= limitX; x++ {
				if !segmentFree(grid, x, y, size, false) {
					continue
				}
				candidateScratch = append(candidateScratch, placement{x: x, y: y, isVertical: false})
			}
		}
	}

	if size <= height {
		limitY := height - size
		for x := uint8(0); x < width; x++ {
			for y := uint8(0); y <= limitY; y++ {
				if !segmentFree(grid, x, y, size, true) {
					continue
				}
				candidateScratch = append(candidateScratch, placement{x: x, y: y, isVertical: true})
			}
		}
	}

	if len(candidateScratch) == 0 {
		return nil, false, false
	}

	chosen := candidateScratch[rand.Intn(len(candidateScratch))]
	coords := make([]Coordinates, size)
	if chosen.isVertical {
		for i := uint8(0); i < size; i++ {
			grid[chosen.y+i][chosen.x] = board.Ship
			coords[i] = Coordinates{x: chosen.x, y: chosen.y + i}
		}
	} else {
		for i := uint8(0); i < size; i++ {
			grid[chosen.y][chosen.x+i] = board.Ship
			coords[i] = Coordinates{x: chosen.x + i, y: chosen.y}
		}
	}

	return coords, chosen.isVertical, true
}

func segmentFree(grid [][]board.Status, x, y, size uint8, vertical bool) bool {
	for i := uint8(0); i < size; i++ {
		if vertical {
			if grid[y+i][x] == board.Ship {
				return false
			}
		} else {
			if grid[y][x+i] == board.Ship {
				return false
			}
		}
	}
	return true
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
