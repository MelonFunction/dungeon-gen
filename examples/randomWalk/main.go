// Package main
package main

import (
	"fmt"

	gen "github.com/melonfunction/dungeon-gen"
)

func main() {
	w, h := 80, 80
	world := gen.NewWorld(w, h)
	world.GenerateRandomWalk(1000)
	world.AddWalls()

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			fmt.Print(world.Tiles[y][x])
		}
		fmt.Println()
	}
}
