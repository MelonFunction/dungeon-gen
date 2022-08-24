// Package generate provides types and functions to generate different dungeons and environments
package generate

import (
	"errors"
	"math/rand"
	"time"
)

// Tile represents the type of tile being used
type Tile int8

// Tiles
const (
	TileVoid Tile = iota
	TileWall
	TileFloor
)

// Tiles aliases
const (
	V = TileVoid
	W = TileWall
	F = TileFloor
)

func (t Tile) String() string {
	switch t {
	case TileVoid:
		return "◾"
	case TileWall:
		return "⬜"
	case TileFloor:
		return "⬛"
	}

	return "UNDEFINED"
}

// World represents the map, Tiles are stored in [y][x] order, but GetTile can be used with (x,y) order to simplify some
// processes
type World struct {
	Width, Height int
	Tiles         [][]Tile // indexed [y][x]
	Border        int      // don't place tiles in this area
	WallThickness int      // how many tiles thick the walls are
}

var (
	rng *rand.Rand
	// ErrOutOfBounds is returned when a tile is attempted to be placed out of bounds
	ErrOutOfBounds = errors.New("Coordinate out of bounds")
)

// NewWorld returns a new World instance
func NewWorld(height, width int) *World {
	s1 := rand.NewSource(time.Now().UnixNano())
	rng = rand.New(s1)

	tiles := make([][]Tile, height)
	for i := range tiles {
		tiles[i] = make([]Tile, width)
	}
	return &World{
		Width:         width,
		Height:        height,
		Tiles:         tiles,
		Border:        2,
		WallThickness: 2,
	}
}

// GetTile returns a tile
func (world *World) GetTile(x, y int) Tile {
	return world.Tiles[y][x]
}

// SetTile sets a tile
func (world *World) SetTile(x, y int, t Tile) error {
	w, h, b := world.Width, world.Height, world.Border
	if t == TileFloor && (x >= w-b || x < 0+b || y >= h-b || y < 0+b) {
		return ErrOutOfBounds
	}

	world.Tiles[y][x] = t
	return nil
}

// AddWalls adds a TileWall around every TileFloor
func (world *World) AddWalls() {
	w, h, t := world.Width, world.Height, world.WallThickness
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if world.GetTile(x, y) == TileFloor {
				for dx := -t; dx <= t; dx++ {
					for dy := -t; dy <= t; dy++ {
						if world.GetTile(x+dx, y+dy) != TileFloor {
							world.SetTile(x+dx, y+dy, TileWall)
						}
					}
				}
			}
		}
	}

}

// GenerateRandomWalk generates the world using the random walk function
func (world *World) GenerateRandomWalk(steps int) {
	w, h := world.Width, world.Height
	x, y := w/2, h/2
	for ; steps > 0; steps-- {
		switch rng.Int() % 4 {
		case 0:
			x--
		case 1:
			x++
		case 2:
			y--
		case 3:
			y++
		}
		if world.SetTile(x, y, TileFloor) == ErrOutOfBounds {
			x = w / 2
			y = h / 2
			steps++
		}
	}
}
