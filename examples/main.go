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
	world.WallThickness = 2
	world.CorridorSize = 2
	world.MaxRoomWidth = 8
	world.MaxRoomHeight = 8
	world.MinRoomWidth = 4
	world.MinRoomHeight = 4
	world.AllowRandomCorridorOffset = true

	err := world.GenerateRandomWalk((80 * 80) / 4)
	// err := world.GenerateDungeonGrid(5 * 5)
	// err := world.GenerateDungeon(30)
	if err != nil {
		log.Println(err)
	}

	world.AddWalls()
	// for i := 0; i < 5; i++ {
	// 	world.CleanWalls(5)
	// }

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			fmt.Print(world.Tiles[y][x])
		}
		fmt.Println()
	}
}
