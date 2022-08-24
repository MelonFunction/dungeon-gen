// Package generate provides types and functions to generate different dungeons and environments
package generate

import (
	"errors"
	"log"
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

	WallThickness int // how many tiles thick the walls are
	CorridorSize  int
	MaxRoomWidth  int
	MaxRoomHeight int
	MinRoomWidth  int
	MinRoomHeight int
}

var (
	rng *rand.Rand
	// ErrOutOfBounds is returned when a tile is attempted to be placed out of bounds
	ErrOutOfBounds = errors.New("Coordinate out of bounds")
	// ErrNotEnoughSpace is returned when there isn't enough space to generate the dungeon
	ErrNotEnoughSpace = errors.New("Not enough space to generate dungeon")
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
		Width:  width,
		Height: height,
		Tiles:  tiles,
		Border: 2,

		WallThickness: 2,

		CorridorSize:  2,
		MaxRoomWidth:  8,
		MaxRoomHeight: 8,
		MinRoomWidth:  4,
		MinRoomHeight: 4,
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
// The world will look chaotic yet natural and all tiles will be touching each other
// world.WallThickness is used
// Ensure that tileCount isn't too high or else world generation can take a while
func (world *World) GenerateRandomWalk(tileCount int) error {
	w, h := world.Width, world.Height
	x, y := w/2, h/2
	for ; tileCount > 0; tileCount-- {
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
		if world.GetTile(x, y) == TileFloor {
			tileCount++
		} else if world.SetTile(x, y, TileFloor) == ErrOutOfBounds {
			x = w / 2
			y = h / 2
			tileCount++
		}
	}

	return nil
}

// GenerateDungeonGrid generates the world using the dungeon grid function
// The world will look neat, with rooms aligned perfectly in a grid. world.MaxRoomWidth is used for both the width and
// the height of the rooms as all rooms are the same size and shape.
// world.WallThickness, world.MaxRoomWidth and world.CorridorSize are used
func (world *World) GenerateDungeonGrid(roomCount int) error {
	s := world.MaxRoomWidth + world.WallThickness*2
	mw := (world.Width - world.Border*2) / s
	mh := (world.Height - world.Border*2) / s
	sx, sy := int(mw/2), int(mh/2)

	// Create rooms layout data structure
	rooms := make([][]bool, mh)
	for i := range rooms {
		rooms[i] = make([]bool, mh)
	}

	// Create the actual rooms layout
	type coord struct {
		x, y int
	}
	previousRooms := make([][]coord, 1)
	for ; roomCount > 0; roomCount-- {
		switch rng.Int() % 4 {
		case 0:
			sx--
		case 1:
			sx++
		case 2:
			sy--
		case 3:
			sy++
		}

		countAdj := func(iy, ix int) int {
			var count int
			if iy > 0 && rooms[iy-1][ix] {
				count++
			}
			if iy+1 < mh && rooms[iy+1][ix] {
				count++
			}
			if ix > 0 && rooms[iy][ix-1] {
				count++
			}
			if ix+1 < mw && rooms[iy][ix+1] {
				count++
			}
			return count
		}

		if sx >= mw || sx <= 0 || sy >= mh || sy <= 0 {
			// roomCount++
			// l := len(previousRooms[len(previousRooms)-1])
			for l := 0; l < len(previousRooms); l++ {
				for i := 0; i < len(previousRooms[l]); i++ { // start from beginning
					roomCoord := previousRooms[l][i]
					roomCount := countAdj(roomCoord.y, roomCoord.x)
					if roomCount > 0 && roomCount <= 2 {
						sx = roomCoord.x
						sy = roomCoord.y
						previousRooms = append(previousRooms, make([]coord, 0))
						goto good
					}
				}
			}

			return ErrNotEnoughSpace
		}

	good:

		if rooms[sy][sx] {
			roomCount++
		}

		// Append room coord for rewinding purposes
		rooms[sy][sx] = true
		previousRooms[len(previousRooms)-1] = append(previousRooms[len(previousRooms)-1], coord{x: sx, y: sy})

	}

	for pr := 0; pr < len(previousRooms); pr++ {
		for i, cur := range previousRooms[pr] {
			sy, sx = cur.y, cur.x

			// Fill in the world's tiles with the room
			for dx := -world.MaxRoomWidth / 2; dx <= world.MaxRoomWidth/2-1; dx++ {
				for dy := -world.MaxRoomWidth / 2; dy <= world.MaxRoomWidth/2-1; dy++ {
					world.SetTile(sx*s+dx, sy*s+dy, TileFloor)
				}
			}

			if i == 0 {
				continue
			}
			prev := previousRooms[pr][i-1]
			dx, dy := cur.x-prev.x, cur.y-prev.y
			var x1, x2, y1, y2 = prev.x * s, cur.x * s, prev.y * s, cur.y * s
			switch {
			case dx == -1: // right
				x1, x2 = x2, x1
				y1--
				y2++
			case dx == 1:
				y1--
				y2++
			case dy == -1:
				y1, y2 = y2, y1
				x1--
				x2++
			case dy == 1:
				x1--
				x2++
			default:
				log.Println("somehow, dx,dy > abs 1", cur, prev, dx, dy)
			}

			for x := x1; x < x2; x++ {
				for y := y1; y < y2; y++ {
					world.SetTile(x, y, TileFloor)
				}
			}
		}
	}

	return nil
}
