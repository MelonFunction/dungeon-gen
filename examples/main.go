// Package main
package main

import (
	"fmt"
	"log"

	gen "github.com/melonfunction/dungeon-gen"
)

// Generation styles
const (
	RandomWalk int = iota
	DungeonGrid
	Dungeon
)

func main() {
	log.SetFlags(log.Lshortfile)

	w, h := 80, 80 // default terminal width (hopefully)
	world := gen.NewWorld(w, h)
	world.Border = 2
	world.CorridorSize = 2
	world.MaxRoomWidth = 8
	world.MaxRoomHeight = 8
	world.MinRoomWidth = 4
	world.MinRoomHeight = 4

	style := Dungeon
	var err error
	switch style {
	case RandomWalk:
		world.WallThickness = 2
		world.MinIslandSize = 26 // 26 is default
		err = world.GenerateRandomWalk((80 * 80) / 4)
		world.CleanIslands()
		world.CleanWalls(5)
		world.CleanWalls(5)
		world.CleanIslands()
		world.CleanWalls(6)
		world.CleanWalls(6)
		world.AddWalls()
	case DungeonGrid:
		world.WallThickness = 2
		world.AllowRandomCorridorOffset = false
		err = world.GenerateDungeonGrid(5 * 5)
		world.AddWalls()
	case Dungeon:
		world.WallThickness = 1
		world.AllowRandomCorridorOffset = true
		err = world.GenerateDungeon(30)
		world.AddWalls()
	}

	if err != nil {
		log.Println(err)
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			fmt.Print(world.Tiles[y][x])
		}
		fmt.Println()
	}
}
