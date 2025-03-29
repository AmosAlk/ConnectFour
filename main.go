package main

import (
	"fmt"
)

func main() {
	RunEbitenGUI()
}

func printBoard(board GameBoard) {
	for _, row := range board {
		for _, cell := range row {
			switch cell {
			case Empty:
				fmt.Print(". ")
			case Player:
				fmt.Print("X ")
			case Computer:
				fmt.Print("O ")
			}
		}
		fmt.Println()
	}
}
