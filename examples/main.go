// Package main
package main

import (
	"fmt"
	"log"

	gen "github.com/melonfunction/dungeon-gen"
)

func main() {
	log.SetFlags(log.Lshortfile)

	w, h := 80, 80
	world := gen.NewWorld(w, h)
	world.Border = 2
	world.WallThickness = 1
	world.CorridorSize = 4
	world.MaxRoomWidth = 8
	world.MaxRoomHeight = 8
	world.MinRoomWidth = 4
	world.MinRoomHeight = 4

	// err := world.GenerateRandomWalk(500)
	// err := world.GenerateDungeonGrid(5 * 5)
	err := world.GenerateDungeon(10)
	if err != nil {
		log.Println(err)
	}

	// world.AddWalls()

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			fmt.Print(world.Tiles[y][x])
		}
		fmt.Println()
	}
}
