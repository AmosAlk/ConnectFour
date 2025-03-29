package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// Game state constants
const (
	StateLogin = iota
	StateGameMode
	StateGame
	StateGameOver
)

// Colors
var (
	colorBackground = color.RGBA{240, 240, 240, 255}
	colorEmpty      = color.RGBA{200, 200, 200, 255}
	colorPlayer     = color.RGBA{255, 50, 50, 255}
	colorComputer   = color.RGBA{50, 50, 255, 255}
	colorButton     = color.RGBA{100, 100, 220, 255}
	colorButtonText = color.RGBA{255, 255, 255, 255}
	colorText       = color.RGBA{10, 10, 10, 255}
	colorHover      = color.RGBA{255, 50, 50, 50}    // Even more transparent (50 alpha)
	colorBoardBg    = color.RGBA{180, 180, 180, 255} // Neutral gray for board background
	colorSlotBg     = color.RGBA{220, 220, 220, 255} // Lighter slots for better contrast
	colorTitleText  = color.RGBA{50, 50, 220, 255}   // Blue title text
)

// Button represents a clickable UI element
type Button struct {
	x, y, w, h float64
	text       string
	action     func()
}

// TextInput represents a text input field
type TextInput struct {
	x, y, w, h float64
	label      string
	value      string
	focused    bool
	isPassword bool
	scrollPos  int // New field for text scrolling
}

// FallingDisc represents a decorative animated disc
type FallingDisc struct {
	x, y    float64
	color   color.RGBA
	size    float64
	speed   float64
	opacity uint8
}

// ConnectFourGame is the main game structure
type ConnectFourGame struct {
	state          int
	board          GameBoard
	turn           int // 1 for player, 2 for computer
	gameInProgress bool
	gameResult     string
	username       string
	password       string

	// UI elements
	buttons      []*Button
	textInputs   []*TextInput
	activeInput  *TextInput
	screenWidth  int
	screenHeight int

	// For responsive layout
	baseWidth  int
	baseHeight int
	scaleX     float64
	scaleY     float64

	// Game board display properties
	cellSize     float64
	boardOffsetX float64
	boardOffsetY float64

	// Hover effect
	hoverColumn int
	isHovering  bool

	// For computer thinking delay
	computerThinking bool
	thinkingTimer    int

	// Decorative elements
	fallingDiscs []FallingDisc
	animTimer    float64

	// For backspace repeat
	backspacePressed bool
	backspaceDelay   int
	backspaceRepeat  int

	// Pre-rendered circle images for better performance
	circleImages map[color.RGBA]*ebiten.Image
}

// Update the NewConnectFourGame function to remove parameters
func NewConnectFourGame() *ConnectFourGame {
	g := &ConnectFourGame{
		state:            StateLogin,
		baseWidth:        800,
		baseHeight:       600,
		screenWidth:      800,
		screenHeight:     600,
		scaleX:           1.0,
		scaleY:           1.0,
		cellSize:         60,
		hoverColumn:      -1,
		computerThinking: false,
		thinkingTimer:    0,
		fallingDiscs:     make([]FallingDisc, 20), // Initialize with 20 decorative discs
		backspacePressed: false,
		backspaceDelay:   15, // Frames to wait before starting to repeat (250ms)
		backspaceRepeat:  3,  // Frames between repeats once started (50ms)
		circleImages:     make(map[color.RGBA]*ebiten.Image),
	}

	// Initialize random falling discs
	rand.Seed(time.Now().UnixNano())
	for i := range g.fallingDiscs {
		g.fallingDiscs[i] = FallingDisc{
			x:       float64(rand.Intn(g.screenWidth)),
			y:       float64(-rand.Intn(g.screenHeight)), // Start above screen
			color:   g.randomDiscColor(),
			size:    float64(20 + rand.Intn(30)),
			speed:   float64(1 + rand.Intn(3)),
			opacity: uint8(100 + rand.Intn(155)), // Semi-transparent
		}
	}

	g.preRenderCircles()
	g.updateLayout() // Apply layout with default dimensions
	g.initUI()       // Initialize UI with default dimensions
	return g
}

// randomDiscColor returns a random disc color
func (g *ConnectFourGame) randomDiscColor() color.RGBA {
	colors := []color.RGBA{
		{255, 50, 50, 255},  // Red (player)
		{50, 50, 255, 255},  // Blue (computer)
		{50, 200, 50, 255},  // Green
		{200, 200, 50, 255}, // Yellow
		{200, 50, 200, 255}, // Purple
	}
	return colors[rand.Intn(len(colors))]
}

// updateLayout recalculates layout based on screen size
func (g *ConnectFourGame) updateLayout() {
	// Set scaling factors
	g.scaleX = float64(g.screenWidth) / float64(g.baseWidth)
	g.scaleY = float64(g.screenHeight) / float64(g.baseHeight)

	// Scale cell size based on screen width (with a minimum size)
	scaleFactor := g.scaleX
	if g.scaleY < g.scaleX {
		scaleFactor = g.scaleY
	}

	g.cellSize = 60 * scaleFactor
	g.boardOffsetX = float64(g.screenWidth-int(float64(Columns)*g.cellSize)) / 2
	g.boardOffsetY = float64(g.screenHeight) * 0.25
}

// initUI sets up the initial UI elements
func (g *ConnectFourGame) initUI() {
	g.buttons = []*Button{}
	g.textInputs = []*TextInput{}

	switch g.state {
	case StateLogin:
		// Username input
		g.textInputs = append(g.textInputs, &TextInput{
			x:         float64(g.screenWidth)/2 - 100*g.scaleX,
			y:         220 * g.scaleY, // Moved down a bit
			w:         200 * g.scaleX,
			h:         30 * g.scaleY,
			label:     "Username:",
			focused:   true,
			scrollPos: 0,
		})
		// Password input
		g.textInputs = append(g.textInputs, &TextInput{
			x:          float64(g.screenWidth)/2 - 100*g.scaleX,
			y:          290 * g.scaleY, // Moved down a bit
			w:          200 * g.scaleX,
			h:          30 * g.scaleY,
			label:      "Password:",
			isPassword: true,
			scrollPos:  0,
		})
		// Login button
		g.buttons = append(g.buttons, &Button{
			x:    float64(g.screenWidth)/2 - 50*g.scaleX,
			y:    350 * g.scaleY, // Moved down a bit
			w:    100 * g.scaleX,
			h:    40 * g.scaleY,
			text: "Login",
			action: func() {
				g.username = g.textInputs[0].value
				g.password = g.textInputs[1].value
				g.state = StateGameMode
				g.initUI()
			},
		})
		g.activeInput = g.textInputs[0]

	case StateGameMode:
		// Play against computer button
		g.buttons = append(g.buttons, &Button{
			x:    float64(g.screenWidth)/2 - 120*g.scaleX,
			y:    200 * g.scaleY,
			w:    240 * g.scaleX,
			h:    40 * g.scaleY,
			text: "Play Against Computer",
			action: func() {
				g.initializeGame()
				g.state = StateGame
				g.initUI()
			},
		})
		// Play online button
		g.buttons = append(g.buttons, &Button{
			x:    float64(g.screenWidth)/2 - 120*g.scaleX,
			y:    260 * g.scaleY,
			w:    240 * g.scaleX,
			h:    40 * g.scaleY,
			text: "Play Online (Coming Soon)",
			action: func() {
				// No action - feature not implemented
			},
		})

	case StateGame:
		// No visible buttons for columns, we'll use hover effect
		// Back button
		g.buttons = append(g.buttons, &Button{
			x:    float64(g.screenWidth) - 120*g.scaleX,
			y:    20 * g.scaleY,
			w:    100 * g.scaleX,
			h:    30 * g.scaleY,
			text: "Back",
			action: func() {
				g.state = StateGameMode
				g.initUI()
			},
		})

	case StateGameOver:
		// Play again button - positioned ABOVE the board
		g.buttons = append(g.buttons, &Button{
			x:    float64(g.screenWidth)/2 - 80*g.scaleX,
			y:    g.boardOffsetY - 100*g.scaleY, // Position above board
			w:    160 * g.scaleX,
			h:    40 * g.scaleY,
			text: "Play Again",
			action: func() {
				g.initializeGame()
				g.state = StateGame
				g.initUI()
			},
		})
		// Back to menu button
		g.buttons = append(g.buttons, &Button{
			x:    float64(g.screenWidth)/2 - 80*g.scaleX,
			y:    g.boardOffsetY - 50*g.scaleY, // Position above board
			w:    160 * g.scaleX,
			h:    40 * g.scaleY,
			text: "Back to Menu",
			action: func() {
				g.state = StateGameMode
				g.initUI()
			},
		})
	}
}

// initializeGame sets up a new game
func (g *ConnectFourGame) initializeGame() {
	g.board = GameBoard{}
	g.gameInProgress = true
	g.turn = Player
	g.gameResult = ""
	g.hoverColumn = -1
	g.isHovering = false
	g.computerThinking = false
	g.thinkingTimer = 0

	// Initialize the board to empty
	for row := range g.board {
		for col := range g.board[row] {
			g.board[row][col] = Empty
		}
	}
}

// Update is called every frame to update the game state
func (g *ConnectFourGame) Update() error {
	// Check if window size changed and update layout
	if w, h := ebiten.WindowSize(); w != g.screenWidth || h != g.screenHeight {
		g.screenWidth = w
		g.screenHeight = h
		g.updateLayout()
		g.initUI()
	}

	// Rest of the Update function remains unchanged
	// ...

	// Update animation timer and falling discs
	g.animTimer += 1.0 / 60.0
	if g.state == StateLogin {
		for i := range g.fallingDiscs {
			disc := &g.fallingDiscs[i]
			disc.y += disc.speed
			if disc.y > float64(g.screenHeight) {
				disc.y = -float64(disc.size)
				disc.x = float64(rand.Intn(g.screenWidth))
				disc.color = g.randomDiscColor()
				disc.opacity = uint8(100 + rand.Intn(155))
			}
		}
	}

	// Handle mouse for hover effects in game state
	if g.state == StateGame && g.gameInProgress && g.turn == Player {
		x, y := ebiten.CursorPosition()

		// Check if mouse is over the board area
		if y >= int(g.boardOffsetY) && y < int(g.boardOffsetY)+int(float64(Rows)*g.cellSize) {
			g.isHovering = false
			g.hoverColumn = -1

			for col := 0; col < Columns; col++ {
				colX := int(g.boardOffsetX) + int(float64(col)*g.cellSize)
				if x >= colX && x < colX+int(g.cellSize) {
					g.hoverColumn = col
					g.isHovering = true
					break
				}
			}
		} else {
			g.isHovering = false
			g.hoverColumn = -1
		}
	}

	// Handle mouse clicks
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()

		// Check if we're in game state and clicking on the board
		if g.state == StateGame && g.gameInProgress && g.turn == Player &&
			g.isHovering && g.hoverColumn >= 0 && g.hoverColumn < Columns {
			if g.board[0][g.hoverColumn] == Empty {
				// Player move
				g.board = dropPiece(g.board, g.hoverColumn, Player)

				// Check for win or tie
				if checkWin(g.board, Player) {
					g.gameResult = "You Won!"
					g.gameInProgress = false
					g.state = StateGameOver
					g.initUI()
				} else if isBoardFull(g.board) {
					g.gameResult = "It's a Tie!"
					g.gameInProgress = false
					g.state = StateGameOver
					g.initUI()
				} else {
					g.turn = Computer
				}
			}
		}

		// Check button clicks
		for _, btn := range g.buttons {
			if float64(x) >= btn.x && float64(x) < btn.x+btn.w &&
				float64(y) >= btn.y && float64(y) < btn.y+btn.h {
				btn.action()
				return nil
			}
		}

		// Check text input focus
		for _, input := range g.textInputs {
			if float64(x) >= input.x && float64(x) < input.x+input.w &&
				float64(y) >= input.y && float64(y) < input.y+input.h {
				// Set this input as active
				g.activeInput = input

				// Only this input should be focused
				for _, otherInput := range g.textInputs {
					otherInput.focused = (otherInput == input)
				}
				break
			}
		}
	}

	// Handle keyboard input for text fields
	if g.activeInput != nil {
		// Handle text input
		runes := ebiten.InputChars()
		if len(runes) > 0 {
			g.activeInput.value += string(runes)

			// Update scroll position if needed
			g.updateTextScroll(g.activeInput)
		}

		// Handle backspace with key repeat
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			g.backspacePressed = true
			g.backspaceDelay = 15 // Reset delay counter

			// Process the first backspace immediately
			if len(g.activeInput.value) > 0 {
				g.activeInput.value = g.activeInput.value[:len(g.activeInput.value)-1]

				// Update scroll position when deleting
				if g.activeInput.scrollPos > 0 {
					g.activeInput.scrollPos = max(0, g.activeInput.scrollPos-1)
				}
			}
		} else if ebiten.IsKeyPressed(ebiten.KeyBackspace) && g.backspacePressed {
			// Key is being held down
			if g.backspaceDelay > 0 {
				g.backspaceDelay--
			} else {
				// After delay, start repeating at the defined interval
				g.backspaceRepeat--
				if g.backspaceRepeat <= 0 {
					g.backspaceRepeat = 3 // Reset repeat counter (3 frames ≈ 50ms at 60fps)

					if len(g.activeInput.value) > 0 {
						g.activeInput.value = g.activeInput.value[:len(g.activeInput.value)-1]

						// Update scroll position when deleting
						if g.activeInput.scrollPos > 0 {
							g.activeInput.scrollPos = max(0, g.activeInput.scrollPos-1)
						}
					}
				}
			}
		} else if inpututil.IsKeyJustReleased(ebiten.KeyBackspace) {
			g.backspacePressed = false
		}

		// Handle tab to switch inputs
		if inpututil.IsKeyJustPressed(ebiten.KeyTab) && len(g.textInputs) > 1 {
			for i, input := range g.textInputs {
				if input == g.activeInput {
					g.activeInput.focused = false
					g.activeInput = g.textInputs[(i+1)%len(g.textInputs)]
					g.activeInput.focused = true
					break
				}
			}
		}

		// Handle enter key
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			if g.state == StateLogin {
				g.username = g.textInputs[0].value
				g.password = g.textInputs[1].value
				g.state = StateGameMode
				g.initUI()
			}
		}
	}

	// Computer move logic
	if g.state == StateGame && g.gameInProgress && g.turn == Computer {
		if !g.computerThinking {
			// Start thinking
			g.computerThinking = true
			g.thinkingTimer = 18 // 18 frames = ~300ms at 60fps
		} else {
			// Continue thinking until timer expires
			g.thinkingTimer--
			if g.thinkingTimer <= 0 {
				// Make move after thinking
				computerCol := getComputerMove(g.board, 5)
				g.board = dropPiece(g.board, computerCol, Computer)
				g.computerThinking = false

				// Check if computer won
				if checkWin(g.board, Computer) {
					g.gameResult = "Computer Won!"
					g.gameInProgress = false
					g.state = StateGameOver
					g.initUI()
				} else if isBoardFull(g.board) {
					g.gameResult = "It's a Tie!"
					g.gameInProgress = false
					g.state = StateGameOver
					g.initUI()
				} else {
					g.turn = Player
				}
			}
		}
	}

	return nil
}

// updateTextScroll updates the text scroll position when input exceeds visible space
func (g *ConnectFourGame) updateTextScroll(input *TextInput) {
	// Calculate the visible width of the text field
	visibleWidth := int(input.w) - 10 // 10 pixels for padding

	// Calculate the text width (approximate)
	displayText := input.value
	if input.isPassword {
		displayText = strings.Repeat("*", len(input.value))
	}

	// Each character is approximately 7 pixels wide
	textWidth := len(displayText) * 7

	// If text exceeds visible width, adjust scroll position
	if textWidth > visibleWidth {
		// Set scroll position to show the end of the text
		input.scrollPos = len(displayText) - (visibleWidth / 7)
	}
}

// Draw renders the game screen
func (g *ConnectFourGame) Draw(screen *ebiten.Image) {
	// Clear screen
	screen.Fill(colorBackground)

	// Draw different screens based on state
	switch g.state {
	case StateLogin:
		g.drawLoginScreen(screen)
	case StateGameMode:
		g.drawGameModeScreen(screen)
	case StateGame, StateGameOver:
		g.drawGameScreen(screen)
	}
}

// Update the drawLoginScreen function with larger title
func (g *ConnectFourGame) drawLoginScreen(screen *ebiten.Image) {
	// Draw animated background discs
	for _, disc := range g.fallingDiscs {
		g.drawSmoothCircle(screen, int(disc.x), int(disc.y), disc.size/2,
			color.RGBA{disc.color.R, disc.color.G, disc.color.B, disc.opacity})
	}

	// Draw decorative board image in background
	boardSize := 300 * g.scaleX
	boardX := float64(g.screenWidth)/2 - boardSize/2
	boardY := 50 * g.scaleY

	// Draw board outline
	ebitenutil.DrawRect(screen, boardX, boardY, boardSize, boardSize,
		color.RGBA{100, 100, 180, 100})

	// Draw title at 3x size
	title := "CONNECT FOUR"
	titleFont := basicfont.Face7x13

	// Create a temporary image for the title text
	titleImg := ebiten.NewImage(titleFont.Metrics().Height.Round()*len(title), titleFont.Metrics().Height.Round()*2)

	// Draw the title to the temporary image
	text.Draw(titleImg, title, titleFont, 0, titleFont.Metrics().Height.Round(), colorTitleText)

	// Draw main text with an additional offset to the right
	mainOp := &ebiten.DrawImageOptions{}
	mainOp.GeoM.Scale(3.0, 3.0) // 3x size
	mainOp.GeoM.Translate(
		float64(g.screenWidth)/2-float64(titleFont.Metrics().Height.Round()*len(title))*1.5+110, // Added 110px offset to the right
		float64(120*g.scaleY)-float64(titleFont.Metrics().Height.Round())*1.5,
	)
	screen.DrawImage(titleImg, mainOp)

	// Draw text inputs
	for _, input := range g.textInputs {
		g.drawTextInput(screen, input)
	}

	// Draw buttons
	for _, btn := range g.buttons {
		g.drawButton(screen, btn)
	}
}

// drawGameModeScreen renders the game mode selection UI
func (g *ConnectFourGame) drawGameModeScreen(screen *ebiten.Image) {
	// Welcome message
	welcome := fmt.Sprintf("Welcome, %s", g.username)
	welcomeBounds := text.BoundString(basicfont.Face7x13, welcome)
	text.Draw(screen, welcome, basicfont.Face7x13,
		g.screenWidth/2-welcomeBounds.Dx()/2, int(100*g.scaleY), colorText)

	// Subtitle
	subtitle := "Select Game Mode:"
	subtitleBounds := text.BoundString(basicfont.Face7x13, subtitle)
	text.Draw(screen, subtitle, basicfont.Face7x13,
		g.screenWidth/2-subtitleBounds.Dx()/2, int(150*g.scaleY), colorText)

	// Draw buttons
	for _, btn := range g.buttons {
		g.drawButton(screen, btn)
	}
}

// Update the drawGameScreen function to ensure white circles look good
func (g *ConnectFourGame) drawGameScreen(screen *ebiten.Image) {
	// Draw game status with better positioning
	var statusText string
	var statusY int

	if g.state == StateGameOver {
		statusText = g.gameResult
		statusY = int(g.boardOffsetY - 130*g.scaleY)
	} else if g.turn == Player {
		statusText = "Your turn - select a column"
		statusY = int(100 * g.scaleY)
	} else {
		statusText = "Computer is thinking..."
		statusY = int(100 * g.scaleY)
	}

	statusBounds := text.BoundString(basicfont.Face7x13, statusText)
	text.Draw(screen, statusText, basicfont.Face7x13,
		g.screenWidth/2-statusBounds.Dx()/2, statusY, colorText)

	// Draw board background (gray border)
	boardWidth := float64(Columns) * g.cellSize
	boardHeight := float64(Rows) * g.cellSize
	ebitenutil.DrawRect(screen,
		g.boardOffsetX-4, g.boardOffsetY-4,
		boardWidth+8, boardHeight+8,
		colorBoardBg)

	// Draw board background (solid color)
	ebitenutil.DrawRect(screen,
		g.boardOffsetX, g.boardOffsetY,
		boardWidth, boardHeight,
		color.RGBA{160, 160, 160, 255}) // Darker gray background for contrast

	// Draw board with proper spacing between circles
	for row := 0; row < Rows; row++ {
		for col := 0; col < Columns; col++ {
			x := int(g.boardOffsetX + float64(col)*g.cellSize + g.cellSize/2)
			y := int(g.boardOffsetY + float64(row)*g.cellSize + g.cellSize/2)

			// First draw white background hole (slightly larger)
			g.drawSmoothCircle(screen, x, y, g.cellSize*0.42, colorSlotBg)

			// Then draw game piece if not empty
			if g.board[row][col] != Empty {
				var pieceColor color.Color
				if g.board[row][col] == Player {
					pieceColor = colorPlayer
				} else {
					pieceColor = colorComputer
				}
				g.drawSmoothCircle(screen, x, y, g.cellSize*0.38, pieceColor)
			}
		}
	}

	// Draw hover effect
	if g.state == StateGame && g.isHovering && g.hoverColumn >= 0 && g.turn == Player {
		if g.board[0][g.hoverColumn] == Empty {
			x := int(g.boardOffsetX + float64(g.hoverColumn)*g.cellSize + g.cellSize/2)
			y := int(g.boardOffsetY + g.cellSize/2) // Top row
			radius := g.cellSize * 0.4
			g.drawSmoothCircle(screen, x, y, radius, colorHover)
		}
	}

	// Draw buttons
	for _, btn := range g.buttons {
		g.drawButton(screen, btn)
	}
}

// drawButton renders a button on the screen
func (g *ConnectFourGame) drawButton(screen *ebiten.Image, btn *Button) {
	// Draw button background
	ebitenutil.DrawRect(screen, btn.x, btn.y,
		btn.w, btn.h, colorButton)

	// Draw button text
	textBounds := text.BoundString(basicfont.Face7x13, btn.text)
	text.Draw(screen, btn.text, basicfont.Face7x13,
		int(btn.x+btn.w/2)-textBounds.Dx()/2,
		int(btn.y+btn.h/2)+textBounds.Dy()/4, colorButtonText)
}

// drawTextInput renders a text input field with scrolling text
func (g *ConnectFourGame) drawTextInput(screen *ebiten.Image, input *TextInput) {
	// Draw label
	text.Draw(screen, input.label, basicfont.Face7x13,
		int(input.x), int(input.y-5), colorText)

	// Draw input background (white with blue border if focused)
	bgColor := color.RGBA{240, 240, 240, 255}
	borderColor := color.RGBA{180, 180, 180, 255}
	if input == g.activeInput {
		borderColor = color.RGBA{100, 100, 220, 255}
	}

	// Draw border
	ebitenutil.DrawRect(screen, input.x-1, input.y-1,
		input.w+2, input.h+2, borderColor)
	// Draw background
	ebitenutil.DrawRect(screen, input.x, input.y,
		input.w, input.h, bgColor)

	// Draw text or placeholder with scrolling
	displayValue := input.value
	if input.isPassword && len(input.value) > 0 {
		displayValue = strings.Repeat("*", len(input.value))
	}

	if len(displayValue) > 0 {
		// Calculate visible portion of text based on scroll position
		startPos := min(input.scrollPos, len(displayValue))
		visibleText := displayValue[startPos:]

		// Calculate max visible characters
		maxVisibleChars := int(input.w-10) / 7 // Assuming 7 pixels per character
		if len(visibleText) > maxVisibleChars {
			visibleText = visibleText[:maxVisibleChars]
		}

		text.Draw(screen, visibleText, basicfont.Face7x13,
			int(input.x+5), int(input.y+input.h/2+5), colorText)
	} else {
		placeholder := fmt.Sprintf("Enter %s", strings.ToLower(input.label[:len(input.label)-1]))
		text.Draw(screen, placeholder, basicfont.Face7x13,
			int(input.x+5), int(input.y+input.h/2+5), color.RGBA{180, 180, 180, 255})
	}

	// Draw cursor ONLY if this is the active input
	if input == g.activeInput {
		// Calculate cursor position based on visible text
		cursorPos := len(displayValue) - input.scrollPos
		cursorPos = min(cursorPos, int(input.w-10)/7) // Don't go outside visible area

		textWidth := cursorPos * 7
		ebitenutil.DrawLine(screen,
			input.x+5+float64(textWidth), input.y+5,
			input.x+5+float64(textWidth), input.y+input.h-5,
			color.RGBA{0, 0, 0, 255})
	}
}

// drawSmoothCircle draws an anti-aliased circle
func (g *ConnectFourGame) drawSmoothCircle(screen *ebiten.Image, centerX, centerY int, radius float64, clr color.Color) {
	// Convert the color to RGBA
	rr, gg, bb, aa := extractRGBA(clr)
	rgba := color.RGBA{R: rr, G: gg, B: bb, A: aa}

	// Try to get the pre-rendered circle
	circleImg, exists := g.circleImages[rgba]

	if !exists {
		// If we don't have this color pre-rendered, create a one-time template
		// (This should rarely happen since we pre-render common colors)
		size := 64
		circleImg = ebiten.NewImage(size, size)
		circleImg.Fill(color.RGBA{0, 0, 0, 0})

		center := float64(size) / 2
		templateRadius := center - 1

		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				dx := float64(x) - center
				dy := float64(y) - center
				dist := math.Sqrt(dx*dx + dy*dy)

				if dist <= templateRadius-2 {
					circleImg.Set(x, y, rgba)
				} else if dist <= templateRadius {
					t := (templateRadius - dist) / 2
					t = t * t * (3 - 2*t)
					alpha := uint8(float64(rgba.A) * t)
					if alpha > 0 {
						circleImg.Set(x, y, color.RGBA{rgba.R, rgba.G, rgba.B, alpha})
					}
				}
			}
		}

		// Save for future use
		g.circleImages[rgba] = circleImg
	}

	// Draw the template with appropriate scaling
	op := &ebiten.DrawImageOptions{}
	scale := (radius * 2) / float64(circleImg.Bounds().Dx())
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(centerX)-radius, float64(centerY)-radius)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(circleImg, op)
}

// preRenderCircles pre-renders circles for common colors
func (g *ConnectFourGame) preRenderCircles() {
	// Define the colors we'll need circles for
	colors := []color.RGBA{
		colorPlayer,
		colorComputer,
		colorSlotBg,
		{255, 50, 50, 50}, // hover color
	}

	// Use a much higher resolution template for better quality
	size := 128 // Double the previous size for better quality

	for _, clr := range colors {
		img := ebiten.NewImage(size, size)
		img.Fill(color.RGBA{0, 0, 0, 0}) // transparent background

		center := float64(size) / 2
		radius := center - 2 // leave a 2px border to avoid clipping

		// Use a wider anti-aliasing region for smoother circles
		aaWidth := 4.0 // 4px wide anti-aliasing border

		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				dx := float64(x) - center
				dy := float64(y) - center
				dist := math.Sqrt(dx*dx + dy*dy)

				if dist <= radius-aaWidth {
					// Solid inner part
					img.Set(x, y, clr)
				} else if dist <= radius {
					// Anti-aliased edge with smoother transition
					t := 1.0 - (dist-(radius-aaWidth))/aaWidth

					// Apply smoothstep function for better transition
					t = t * t * (3 - 2*t)

					r, g, b, a := clr.R, clr.G, clr.B, clr.A
					alpha := uint8(float64(a) * t)
					if alpha > 0 {
						img.Set(x, y, color.RGBA{r, g, b, alpha})
					}
				}
			}
		}

		g.circleImages[clr] = img
	}
}

// extractRGBA extracts uint8 RGBA components from a color.Color
func extractRGBA(c color.Color) (r, g, b, a uint8) {
	rr, gg, bb, aa := c.RGBA()
	return uint8(rr >> 8), uint8(gg >> 8), uint8(bb >> 8), uint8(aa >> 8)
}

// Helper functions for min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Layout returns the game's screen dimensions
func (g *ConnectFourGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight // Make the game fully resizable
}

// Update the RunEbitenGUI function to remove maximization
func RunEbitenGUI() {
	// Set window properties
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Connect Four")
	ebiten.SetWindowResizable(true)

	// Create the game with default dimensions
	game := NewConnectFourGame()

	// Run the game
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
