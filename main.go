package main

import (
	"fmt"
	"math/rand"
)

const (
	columns  = 7
	rows     = 6
	local    = 'X'
	opponent = 'O'
	empty    = '.'
)

func main() {
	grid := initGrid()
	localTurn := true
	grid.print()
	var column int
	for {
		if localTurn {
			for {
				fmt.Println("Your move: ")
				fmt.Scan(&column)
				column-- // Convert to 0-indexed.
				availableColumn := false
				for i := range grid.nonFullCols {
					if column == grid.nonFullCols[i] {
						availableColumn = true
						break
					}
				}
				if !availableColumn {
					fmt.Println("Try again. Column is full or invalid.")
				} else {
					break
				}
			}
			grid.drop(column, local)

		} else {
			fmt.Println("Computer's move: ")
			grid.computerMove()
		}
		grid.print()
		localTurn = !localTurn
		grid.checkWin()
		if grid.winner != empty {
			if grid.winner == local {
				fmt.Println("You win!")
			} else {
				fmt.Println("Computer wins!")
			}
			break
		}
		if len(grid.nonFullCols) == 0 {
			fmt.Println("It's a tie!")
			break
		}
	}
}

type Grid struct {
	board       [rows][columns]rune
	winner      rune
	nonFullCols []int
}

// Description of grid shown below:
//   C 0   1   2   3   4   5   6
// R ------------------------------
// 0 | 00  01  02  03  04  05  06 |
// 1 | 10  11  12  13  14  15  16 |
// 2 | 20  21  22  23  24  25  26 |
// 3 | 30  31  32  33  34  35  36 |
// 4 | 40  41  42  43  44  45  46 |
// 5 | 50  51  52  53  54  55  56 |
//   ------------------------------

func initGrid() *Grid {
	grid := &Grid{
		board:       [rows][columns]rune{},
		winner:      empty,
		nonFullCols: []int{0, 1, 2, 3, 4, 5, 6},
	}
	for i := range grid.board {
		for j := range grid.board[i] {
			grid.board[i][j] = empty
		}
	}
	return grid
}

func (g *Grid) print() {
	for i := range g.board {
		for j := range g.board[i] {
			fmt.Printf("%c ", g.board[i][j]) // %c is the format specifier for a rune.
		}
		fmt.Println()
	}
}

func (g *Grid) drop(column int, player rune) {
	for i := rows - 1; i >= 0; i-- {
		if g.board[i][column] == empty { // Already converted to 0-indexed.
			g.board[i][column] = player
			return
		}
		if i == 1 {
			for idx, col := range g.nonFullCols {
				if col == column {
					g.nonFullCols = append(g.nonFullCols[:idx], g.nonFullCols[idx+1:]...)
					break
				}
			}
		}
	}
	g.print()
}

func (g *Grid) checkWin() {
	// Check horizontal
	for i := 0; i < rows; i++ {
		for j := 0; j <= columns-4; j++ {
			if g.board[i][j] != empty && g.board[i][j] == g.board[i][j+1] && g.board[i][j] == g.board[i][j+2] && g.board[i][j] == g.board[i][j+3] {
				g.winner = g.board[i][j]
			}
		}
	}

	// Check vertical
	for i := 0; i <= rows-4; i++ {
		for j := 0; j < columns; j++ {
			if g.board[i][j] != empty && g.board[i][j] == g.board[i+1][j] && g.board[i][j] == g.board[i+2][j] && g.board[i][j] == g.board[i+3][j] {
				g.winner = g.board[i][j]
			}
		}
	}

	// Check diagonal (bottom-left to top-right)
	for i := 0; i <= rows-4; i++ {
		for j := 0; j <= columns-4; j++ {
			if g.board[i][j] != empty && g.board[i][j] == g.board[i+1][j+1] && g.board[i][j] == g.board[i+2][j+2] && g.board[i][j] == g.board[i+3][j+3] {
				g.winner = g.board[i][j]
			}
		}
	}

	// Check diagonal (top-left to bottom-right)
	for i := 3; i < rows; i++ {
		for j := 0; j <= columns-4; j++ {
			if g.board[i][j] != empty && g.board[i][j] == g.board[i-1][j+1] && g.board[i][j] == g.board[i-2][j+2] && g.board[i][j] == g.board[i-3][j+3] {
				g.winner = g.board[i][j]
			}
		}
	}
}

func (g *Grid) computerMove() {
	for {
		// Selects random index from length of nonFullCols and drops in that column.
		column := g.nonFullCols[rand.Intn(len(g.nonFullCols)-1)] // Random column between 1 and 7
		if g.board[0][column] == empty {                         // Drop is 0-indexed, fixed in drop.
			g.drop(column, opponent)
			fmt.Println("Computer dropped at column", column+1)
			return
		}
	}
}
