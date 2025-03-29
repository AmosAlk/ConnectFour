package main

import (
	"math"
	"math/rand"
	"time"
)

const (
	Rows     = 6
	Columns  = 7
	Empty    = 0
	Player   = 1
	Computer = 2
)

type GameBoard [Rows][Columns]int

// Evaluate the board to determine the score for the computer.
func evaluateBoard(board GameBoard) int {
	// Scoring logic for the board
	// Positive score favors the computer, negative favors the player
	score := 0

	// Check horizontal, vertical, and diagonal lines for scoring
	score += evaluateLines(board, Computer)
	score -= evaluateLines(board, Player)

	return score
}

// Evaluate lines for a specific player
func evaluateLines(board GameBoard, player int) int {
	score := 0

	// Horizontal
	for row := 0; row < Rows; row++ {
		for col := 0; col < Columns-3; col++ {
			score += evaluateSegment(board[row][col:col+4], player)
		}
	}

	// Vertical
	for col := 0; col < Columns; col++ {
		for row := 0; row < Rows-3; row++ {
			segment := []int{board[row][col], board[row+1][col], board[row+2][col], board[row+3][col]}
			score += evaluateSegment(segment, player)
		}
	}

	// Diagonal (top-left to bottom-right)
	for row := 0; row < Rows-3; row++ {
		for col := 0; col < Columns-3; col++ {
			segment := []int{board[row][col], board[row+1][col+1], board[row+2][col+2], board[row+3][col+3]}
			score += evaluateSegment(segment, player)
		}
	}

	// Diagonal (bottom-left to top-right)
	for row := 3; row < Rows; row++ {
		for col := 0; col < Columns-3; col++ {
			segment := []int{board[row][col], board[row-1][col+1], board[row-2][col+2], board[row-3][col+3]}
			score += evaluateSegment(segment, player)
		}
	}

	return score
}

// Evaluate a segment of 4 cells for scoring
func evaluateSegment(segment []int, player int) int {
	score := 0
	countPlayer := 0
	countEmpty := 0

	for _, cell := range segment {
		if cell == player {
			countPlayer++
		} else if cell == Empty {
			countEmpty++
		}
	}

	if countPlayer == 4 {
		score += 100
	} else if countPlayer == 3 && countEmpty == 1 {
		score += 10
	} else if countPlayer == 2 && countEmpty == 2 {
		score += 5
	}

	return score
}

// Check if the game is over
func isTerminalNode(board GameBoard) bool {
	return checkWin(board, Player) || checkWin(board, Computer) || isBoardFull(board)
}

// Check if a player has won
func checkWin(board GameBoard, player int) bool {
	// Horizontal
	for row := 0; row < Rows; row++ {
		for col := 0; col < Columns-3; col++ {
			if board[row][col] == player && board[row][col+1] == player && board[row][col+2] == player && board[row][col+3] == player {
				return true
			}
		}
	}

	// Vertical
	for col := 0; col < Columns; col++ {
		for row := 0; row < Rows-3; row++ {
			if board[row][col] == player && board[row+1][col] == player && board[row+2][col] == player && board[row+3][col] == player {
				return true
			}
		}
	}

	// Diagonal (top-left to bottom-right)
	for row := 0; row < Rows-3; row++ {
		for col := 0; col < Columns-3; col++ {
			if board[row][col] == player && board[row+1][col+1] == player && board[row+2][col+2] == player && board[row+3][col+3] == player {
				return true
			}
		}
	}

	// Diagonal (bottom-left to top-right)
	for row := 3; row < Rows; row++ {
		for col := 0; col < Columns-3; col++ {
			if board[row][col] == player && board[row-1][col+1] == player && board[row-2][col+2] == player && board[row-3][col+3] == player {
				return true
			}
		}
	}

	return false
}

// Check if the board is full
func isBoardFull(board GameBoard) bool {
	for col := 0; col < Columns; col++ {
		if board[0][col] == Empty {
			return false
		}
	}
	return true
}

// Get all valid columns for the next move
func getValidColumns(board GameBoard) []int {
	validColumns := []int{}
	for col := 0; col < Columns; col++ {
		if board[0][col] == Empty {
			validColumns = append(validColumns, col)
		}
	}
	return validColumns
}

// Drop a piece in the specified column
func dropPiece(board GameBoard, col, player int) GameBoard {
	for row := Rows - 1; row >= 0; row-- {
		if board[row][col] == Empty {
			board[row][col] = player
			break
		}
	}
	return board
}

// Minimax algorithm with alpha-beta pruning
func minimax(board GameBoard, depth int, alpha float64, beta float64, maximizingPlayer bool) (int, float64) {
	validColumns := getValidColumns(board)
	isTerminal := isTerminalNode(board)

	if depth == 0 || isTerminal {
		if isTerminal {
			if checkWin(board, Computer) {
				return -1, math.Inf(1)
			} else if checkWin(board, Player) {
				return -1, math.Inf(-1)
			} else {
				return -1, 0
			}
		}
		return -1, float64(evaluateBoard(board))
	}

	if maximizingPlayer {
		value := math.Inf(-1)
		column := validColumns[rand.Intn(len(validColumns))]
		for _, col := range validColumns {
			newBoard := dropPiece(board, col, Computer)
			_, newScore := minimax(newBoard, depth-1, alpha, beta, false)
			if newScore > value {
				value = newScore
				column = col
			}
			alpha = math.Max(alpha, value)
			if alpha >= beta {
				break
			}
		}
		return column, value
	} else {
		value := math.Inf(1)
		column := validColumns[rand.Intn(len(validColumns))]
		for _, col := range validColumns {
			newBoard := dropPiece(board, col, Player)
			_, newScore := minimax(newBoard, depth-1, alpha, beta, true)
			if newScore < value {
				value = newScore
				column = col
			}
			beta = math.Min(beta, value)
			if alpha >= beta {
				break
			}
		}
		return column, value
	}
}

// Get the computer's move
func getComputerMove(board GameBoard, depth int) int {
	rand.Seed(time.Now().UnixNano())
	column, _ := minimax(board, depth, math.Inf(-1), math.Inf(1), true)
	return column
}
