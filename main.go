package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	cube := NewCube()
	if cube.Login() {
		fmt.Println("Login success!")
		cube.OpenBoxes()
		notOwned := cube.CheckFreeGames()
		for _, game := range notOwned {
			if !cube.GetFreeGame(game) {
				break
			}
			time.Sleep(time.Duration(float64(rand.Intn(5))-rand.Float64()) * time.Second)
		}
	} else {
		fmt.Println("Login failed!")
	}
	var exit string
	fmt.Println("\nPress enter to exit...")
	fmt.Scanln(&exit)
}
