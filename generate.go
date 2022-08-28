// Package generate provides types and functions to generate different dungeons and environments
package generate

import (
	"errors"
	"fmt"
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
	TilePreWall // placeholder for walls during generation
	TileFloor
)

// Tiles aliases
const (
	V = TileVoid
	W = TileWall
	P = TilePreWall
	F = TileFloor
)

func (t Tile) String() string {
	switch t {
	case TileVoid:
		return "◾"
	case TilePreWall:
		return "🔳"
	case TileWall:
		return "⬜"
	case TileFloor:
		return "⬛"
	}

	return "🚧"
}

// World represents the map, Tiles are stored in [y][x] order, but GetTile can be used with (x,y) order to simplify some
// processes
type World struct {
	Width, Height int
	Tiles         [][]Tile // indexed [y][x]
	Border        int      // don't place tiles in this area

	startTime           time.Time // for generation retry
	DurationBeforeRetry time.Duration
	genStartTime        time.Time // for error
	DurationBeforeError time.Duration

	WallThickness             int // how many tiles thick the walls are
	CorridorSize              int
	AllowRandomCorridorOffset bool
	MaxRoomWidth              int
	MaxRoomHeight             int
	MinRoomWidth              int
	MinRoomHeight             int
}

var (
	rng *rand.Rand
	// ErrOutOfBounds is returned when a tile is attempted to be placed out of bounds
	ErrOutOfBounds = errors.New("Coordinate out of bounds")
	// ErrNotEnoughSpace is returned when there isn't enough space to generate the dungeon
	ErrNotEnoughSpace = errors.New("Not enough space to generate dungeon")
	// ErrGenerationTimeout is returned when generation has deadlocked
	ErrGenerationTimeout = errors.New("Took too long to generate dungeon")
	// ErrFloorAlreadyPlaced is returned when a floor tile is already placed
	ErrFloorAlreadyPlaced = errors.New("Floor tile already placed")
)

// ClearTiles clears the tiles from the world
func (world *World) ClearTiles(width, height int) {
	tiles := make([][]Tile, height)
	for i := range tiles {
		tiles[i] = make([]Tile, width)
	}
	world.Tiles = tiles
}

// NewWorld returns a new World instance
func NewWorld(width, height int) *World {
	s1 := rand.NewSource(time.Now().UnixNano())
	rng = rand.New(s1)

	world := &World{
		Width:  width,
		Height: height,
		Border: 2,

		startTime:           time.Now(),
		DurationBeforeRetry: time.Millisecond * 250,
		DurationBeforeError: time.Second * 2,

		WallThickness: 2,

		CorridorSize:              2,
		AllowRandomCorridorOffset: false,
		MaxRoomWidth:              8,
		MaxRoomHeight:             8,
		MinRoomWidth:              4,
		MinRoomHeight:             4,
	}
	world.ClearTiles(width, height)
	return world
}

// GetTile returns a tile
func (world *World) GetTile(x, y int) (Tile, error) {
	w, h, b := world.Width, world.Height, world.Border
	if x >= w-b || x < 0+b || y >= h-b || y < 0+b {
		return TileVoid, ErrOutOfBounds
	}
	return world.Tiles[y][x], nil
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
	b := world.Border
	world.Border = 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if tile, err := world.GetTile(x, y); err == nil {
				switch tile {
				case TileFloor:
					for dx := -t; dx <= t; dx++ {
						for dy := -t; dy <= t; dy++ {
							if tile, err := world.GetTile(x+dx, y+dy); err == nil && tile == TileVoid {
								world.SetTile(x+dx, y+dy, TileWall)
							}
						}
					}
				case TilePreWall:
					world.SetTile(x, y, TileWall)
				}
			}
		}
	}
	world.Border = b
}

// CleanWalls replaces walls which don't have mustSurroundCount walls around them
func (world *World) CleanWalls(mustSurroundCount int) {
	w, h := world.Width, world.Height
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if tile, err := world.GetTile(x, y); err == nil && tile == TileWall {
				var count int
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 1; dy++ {
						if !(dx == 0 && dy == 0) {
							if tile, err := world.GetTile(x+dx, y+dy); err == nil && tile == TileFloor {
								count++
							}
						}
					}
				}
				if count >= mustSurroundCount {
					world.SetTile(x, y, TileFloor)
				}
			}
		}
	}
}

// GenerateRandomWalk generates the world using the random walk function
// The world will look chaotic yet natural and all tiles will be touching each other
// world.Convexity, world.WallThickness and world.CorridorSize is used
// Ensure that tileCount isn't too high or else world generation can take a while
func (world *World) GenerateRandomWalk(tileCount int) error {
	world.genStartTime = time.Now()

	w, h := world.Width, world.Height

	var g func() error
	g = func() error {
		world.ClearTiles(world.Width, world.Height)
		x, y := w/2, h/2
		minX, maxX, minY, maxY := w, 0, h, 0

		for tc := 0; tc < tileCount; {
			if time.Now().Sub(world.genStartTime) > world.DurationBeforeError {
				return ErrGenerationTimeout
			} else if time.Now().Sub(world.startTime) > world.DurationBeforeRetry {
				log.Println("Timeout, retrying gen")
				return g()
			}

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
			for dx := x - world.CorridorSize/2; dx < x+world.CorridorSize/2; dx++ {
				for dy := y - world.CorridorSize/2; dy < y+world.CorridorSize/2; dy++ {
					tc++
					if tile, err := world.GetTile(dx, dy); err == nil && tile != TileVoid {
						tc--
					} else if world.SetTile(dx, dy, TileFloor) == ErrOutOfBounds {
						x = w / 2
						y = h / 2
						tc--
						goto cont
					}
				}
			}

			minX = minInt(minX, x)
			maxX = maxInt(maxX, x)
			minY = minInt(minY, y)
			maxY = maxInt(maxY, y)
		cont:
		}

		// Check that the generation is acceptable
		// TODO params
		// Bounds
		if (maxX-minX < w/2) && (maxY-minY < h/2) {
			log.Println("bounds too small, retrying gen")
			return g()
		}
		// Convexity
		var convX, convY bool
		// log.Println(minX, maxX, minY, maxY)
		cy := minY + (maxY-minY)/2
		world.SetTile(0, cy, TileWall)
		var foundFloor, inGap bool
		for cx := minX; cx < maxX; cx++ {
			if tile, err := world.GetTile(cx, cy); err == nil {
				switch tile {
				case TileFloor:
					if foundFloor && inGap {
						convX = true
						goto done
					}
					foundFloor = true
				case TileVoid:
					if foundFloor {
						inGap = true
					}
				}
			}
		}
	done:
		if !(convX || convY) {
			log.Println("no convexity, retrying gen")
			return g()
		}

		return nil
	}

	return g()
}

type coord struct {
	x, y int
	w, h int
}

// GenerateDungeonGrid generates the world using the dungeon grid function
// The world will look neat, with rooms aligned perfectly in a grid. world.MaxRoomWidth is used for both the width and
// the height of the rooms as all rooms are the same size and shape.
// world.WallThickness, world.MaxRoomWidth and world.CorridorSize and world.AllowRandomCorridorOffset are used
func (world *World) GenerateDungeonGrid(roomCount int) error {
	world.genStartTime = time.Now()

	s := world.MaxRoomWidth
	mw := (world.Width - world.Border*2) / s
	mh := (world.Height - world.Border*2) / s

	fmt.Printf("Max grid size is %d x %d, so max roomCount is %d. Use fewer rooms for a better result.\n", mw-2, mh-2, (mw-2)*(mh-2))

	if roomCount > (mw-2)*(mh-2) {
		return ErrNotEnoughSpace
	}

	var g func() error
	g = func() error {
		world.ClearTiles(world.Width, world.Height)
		sx, sy := int(mw/2), int(mh/2)
		world.startTime = time.Now()
		// Create rooms layout data structure
		rooms := make([][]bool, mh)
		for i := range rooms {
			rooms[i] = make([]bool, mw)
		}

		previousRooms := make([][]coord, 1)
		for rc := roomCount; rc > 0; rc-- {
			if time.Now().Sub(world.genStartTime) > world.DurationBeforeError {
				return ErrGenerationTimeout
			} else if time.Now().Sub(world.startTime) > world.DurationBeforeRetry {
				log.Println("Timeout, retrying gen")
				return g()
			}
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

			if sx >= mw-1 || sx <= 0 || sy >= mh-1 || sy <= 0 || (countAdj(sy, sx) >= 2 && rooms[sy][sx]) {
				rc++
				for l := 0; l < len(previousRooms); l++ {
					for i := 0; i < len(previousRooms[l]); i++ { // start from beginning
						roomCoord := previousRooms[l][i]
						rc := countAdj(roomCoord.y, roomCoord.x)
						if rc >= 0 && rc <= 2 {
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
			// Append room coord for rewinding purposes
			rooms[sy][sx] = true
			previousRooms[len(previousRooms)-1] = append(previousRooms[len(previousRooms)-1], coord{x: sx, y: sy})
		}

		for pr := 0; pr < len(previousRooms); pr++ {
			// log.Println(previousRooms[pr])
			for i, cur := range previousRooms[pr] {
				sy, sx = cur.y, cur.x

				// Fill in the world's tiles with the room
				for dx := -world.MaxRoomWidth / 2; dx <= world.MaxRoomWidth/2-1; dx++ {
					for dy := -world.MaxRoomWidth / 2; dy <= world.MaxRoomWidth/2-1; dy++ {
						world.SetTile(sx*s+dx+sx*world.WallThickness, sy*s+dy+sy*world.WallThickness, TileFloor)
					}
				}

				if i == 0 {
					continue
				}

				// Corridors
				prev := previousRooms[pr][i-1]
				dx, dy := cur.x-prev.x, cur.y-prev.y
				var x1, x2, y1, y2 = prev.x * s, cur.x * s, prev.y * s, cur.y * s
				var offsetCy, offsetCx int
				if world.AllowRandomCorridorOffset {
					offsetCy = (world.MaxRoomWidth - world.CorridorSize)
					offsetCy = randInt(-offsetCy/2, offsetCy/2)
					offsetCx = (world.MaxRoomWidth - world.CorridorSize)
					offsetCx = randInt(-offsetCx/2, offsetCx/2)
				}
				switch {
				case dx == -1: // right
					x1, x2 = x2, x1
					y1 = y1 - world.CorridorSize/2 - offsetCy
					y2 = y2 + world.CorridorSize/2 - offsetCy
				case dx == 1:
					y1 = y1 - world.CorridorSize/2 - offsetCy
					y2 = y2 + world.CorridorSize/2 - offsetCy
				case dy == -1:
					y1, y2 = y2, y1
					x1 = x1 - world.CorridorSize/2 - offsetCx
					x2 = x2 + world.CorridorSize/2 - offsetCx
				case dy == 1:
					x1 = x1 - world.CorridorSize/2 - offsetCx
					x2 = x2 + world.CorridorSize/2 - offsetCx
				default:
					log.Println("somehow, dx,dy > abs 1", cur, prev, dx, dy)
				}

				for x := x1; x < x2; x++ {
					for y := y1; y < y2; y++ {
						world.SetTile(x+sx*world.WallThickness, y+sy*world.WallThickness, TileFloor)
					}
				}
			}
		}
		return nil
	}
	return g()
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func absInt(a int) int {
	if a < 0 {
		return a * -1
	}
	return a
}
func randInt(a, b int) int {
	return rng.Int()%(b+1-a) + a
}

// GenerateDungeon generates the world using a more fluid algorithm
// The world will have randomly sized rooms
// world.WallThickness, world.MinRoomWidth|Height, world.MaxRoomWidth|Height, world.CorridorSize and
// world.AllowRandomCorridorOffset are used
func (world *World) GenerateDungeon(roomCount int) error {
	world.genStartTime = time.Now()

	s := world.MaxRoomWidth
	mw := (world.Width - world.Border*2) / s
	mh := (world.Height - world.Border*2) / s

	if roomCount > (mw-2)*(mh-2) {
		return ErrNotEnoughSpace
	}

	var g func() error
	g = func() error {
		world.ClearTiles(world.Width, world.Height)
		world.startTime = time.Now()

		// Helper func to place rooms
		placeRoom := func(x, y, w, h int) error {
			log.Println(x, y, w, h)
			// Check area
			for dx := -w/2 - world.WallThickness; dx <= w/2-1+world.WallThickness; dx++ {
				for dy := -h/2 - world.WallThickness; dy <= h/2-1+world.WallThickness; dy++ {
					if tile, err := world.GetTile(x+dx, y+dy); err == nil && tile == TileFloor {
						return ErrFloorAlreadyPlaced
					} else if err != nil {
						return err
					}
				}
			}
			// Place
			for dx := -w/2 - world.WallThickness; dx <= w/2-1+world.WallThickness; dx++ {
				for dy := -h/2 - world.WallThickness; dy <= h/2-1+world.WallThickness; dy++ {
					if dx < -w/2 || dx > w/2-1 || dy < -h/2 || dy > h/2-1 {
						// Temp wall
						if tile, err := world.GetTile(x+dx, y+dy); err == nil && tile == TileVoid {
							if err := world.SetTile(x+dx, y+dy, TilePreWall); err != nil {
								return err
							}
						}
					} else {
						// Floor
						if err := world.SetTile(x+dx, y+dy, TileFloor); err != nil {
							return err
						}
					}
				}
			}
			return nil
		}

		// Random first room size
		sx, sy := world.Width/2, world.Height/2
		rw := randInt(world.MinRoomWidth, world.MaxRoomWidth)
		rh := randInt(world.MinRoomHeight, world.MaxRoomHeight)

		// Place the first room into the world
		placeRoom(sx, sy, rw, rh)

		previousRooms := make([]coord, 1)
		for rc := roomCount - 1; rc > 0; rc-- {
			if time.Now().Sub(world.genStartTime) > world.DurationBeforeError {
				return ErrGenerationTimeout
			} else if time.Now().Sub(world.startTime) > world.DurationBeforeRetry {
				log.Println("Timeout, retrying gen")
				return g()
			}

			log.Println(rc)

			// Offset position by last room
			osx := sx
			osy := sy
			orw := rw
			orh := rh
			rw = randInt(world.MinRoomWidth, world.MaxRoomWidth)
			rh = randInt(world.MinRoomHeight, world.MaxRoomHeight)
			cx, cy := osx, osy // corridor position
			var cw, ch int
			var offsetCy, offsetCx int
			if world.AllowRandomCorridorOffset {
				offsetCy = (minInt(rh, orh) - ch)
				offsetCy = randInt(-offsetCy/2+world.CorridorSize/2, offsetCy/2-world.CorridorSize/2)
				offsetCx = (minInt(rw, orw) - cw)
				offsetCx = randInt(-offsetCx/2+world.CorridorSize/2, offsetCx/2-world.CorridorSize/2)
			}
			switch rng.Int() % 4 {
			case 0: // left
				cw = world.WallThickness
				ch = world.CorridorSize
				sx = sx - orw/2 - world.WallThickness - rw/2
				cx = sx + rw/2
				cy = cy - (ch / 2) + offsetCy
			case 1: // right
				cw = world.WallThickness
				ch = world.CorridorSize
				sx = sx + orw/2 + world.WallThickness + rw/2
				cx = sx - rw/2 - world.WallThickness
				cy = cy - (ch / 2) + offsetCy
			case 2: // up
				cw = world.CorridorSize
				ch = world.WallThickness
				sy = sy - orh/2 - world.WallThickness - rh/2
				cy = sy + rh/2
				cx = cx - (cw / 2) + offsetCx
			case 3: // down
				cw = world.CorridorSize
				ch = world.WallThickness
				sy = sy + orh/2 + world.WallThickness + rh/2
				cy = sy - rh/2 - world.WallThickness
				cx = cx - (cw / 2) + offsetCx
			}

			if err := placeRoom(sx, sy, rw, rh); err != nil {
				log.Println("rollback:", err)
				// rollback
				c := previousRooms[rng.Int()%len(previousRooms)]
				sx = c.x
				sy = c.y
				rw = c.w
				rh = c.h
				rc++
				continue
			}

			// Corridors
			for x := cx; x < cx+cw; x++ {
				for y := cy; y < cy+ch; y++ {
					log.Println(x, y)
					if tile, err := world.GetTile(x, y); err == nil && tile != TileFloor {
						if err := world.SetTile(x, y, TileFloor); err != nil {
						}
					}
				}
			}

			previousRooms = append(previousRooms, coord{x: sx, y: sy, w: rw, h: rh})
		}

		return nil
	}
	return g()

}
