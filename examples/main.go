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
	err := world.GenerateDungeonGrid(16)
	if err != nil {
		log.Println(err)
	}
	world.WallThickness = 2
	world.CorridorSize = 2
	world.MaxRoomWidth = 8
	world.MaxRoomHeight = 8
	world.MinRoomWidth = 4
	world.MinRoomHeight = 4
	world.AddWalls()

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			fmt.Print(world.Tiles[y][x])
		}
		fmt.Println()
	}
}