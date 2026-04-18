package main

import (
	"battleship/components/board"
	"battleship/components/player"
	"battleship/components/ship"
	"battleship/services/gameService"
	"log"
	"math"
	"runtime"
	"time"

	"embed"
	"image/color"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed assets/*
var assets embed.FS

var (
	err            error
	masterSheet    *ebiten.Image
	images         map[string]*ebiten.Image
	fontSource     *text.GoTextFaceSource
	engravedColor  = color.RGBA{R: 182, G: 142, B: 101, A: 255}
	highlightColor = color.RGBA{R: 193, G: 154, B: 107, A: 150}
)

const (
	windowWidth  = 820
	windowHeight = 615
)

type screenState uint8

const (
	homeScreen screenState = iota
	howToPlayScreen
	settingsScreen
	gameScreen
	revealScreen
	resultScreen
)

const revealDuration = 3500 * time.Millisecond

type Game struct {
	state screenState

	homeButtons    map[string]*animatedButton
	playButton     *animatedButton
	restartButton  *animatedButton
	mainMenuButton *animatedButton
	quitButton     *animatedButton

	shipsMinus *controlHotspot
	shipsPlus  *controlHotspot
	gridMinus  *controlHotspot
	gridPlus   *controlHotspot

	player         *player.Player
	gameBoard      *board.Board
	numberOfShips  int
	gridHeight     uint8
	gridWidth      uint8
	gameStartedAt  time.Time
	gameFinishedAt time.Time
	revealStartAt  time.Time
	turnsTaken     uint16

	gridOriginX float64
	gridOriginY float64
	cellSize    float64

	prevMousePressed bool
	shouldQuit       bool

	focusedField    string
	lastFocusedGrid string
	inputBuffer     string
}

func NewGame() *Game {
	g := &Game{
		state: homeScreen,
		homeButtons: map[string]*animatedButton{
			"playButton":      newAnimatedButton("playButton", 225, 225, 1.0, 1.0),
			"howToPlayButton": newAnimatedButton("howToPlayButton", 225, 325, 1.0, 1.0),
			"settingsButton":  newAnimatedButton("settingsButton", 225, 425, 1.0, 1.0),
			"quitButton":      newAnimatedButton("quitButton", 225, 525, 1.0, 1.0),
		},
		playButton:      newAnimatedButton("playButton", 425, 510, 0.99, 0.99),
		restartButton:   newAnimatedButton("restartButton", 425, 510, 0.87, 0.87),
		mainMenuButton:  newAnimatedButton("mainMenuButton", 65, 510, 0.85, 0.85),
		quitButton:      newAnimatedButton("quitButton", 430, 500, 1.0, 1.0),
		shipsMinus:      newControlHotspot(515, 264, 70, 50),
		shipsPlus:       newControlHotspot(593, 264, 70, 50),
		gridMinus:       newControlHotspot(529, 363, 70, 50),
		gridPlus:        newControlHotspot(607, 363, 70, 50),
		gridHeight:      10,
		gridWidth:       10,
		numberOfShips:   5,
		gridOriginX:     65,
		gridOriginY:     75,
		cellSize:        42,
		lastFocusedGrid: "gridHeight",
	}

	if cfg := loadConfig(); cfg != nil {
		g.numberOfShips = cfg.NumberOfShips
		g.gridHeight = cfg.GridHeight
		g.gridWidth = cfg.GridWidth
		g.normalizeShipsToGrid()
	}

	return g
}

func (g *Game) persistSettings() {
	saveConfig(gameConfig{
		NumberOfShips: g.numberOfShips,
		GridHeight:    g.gridHeight,
		GridWidth:     g.gridWidth,
	})
}

func (g *Game) startNewGame() {
	g.persistSettings()

	// NOTE : Drop hard references to the previous game's board/player so the GC can
	// collect them; without this they stayed alive through the assignment
	// below, causing peak memory to spike on back-to-back restarts.
	g.gameBoard = nil
	g.player = nil
	gameService.ResetGameState()

	g.player, g.gameBoard = gameService.NewGameService("Commander", g.gridHeight, g.gridWidth, g.numberOfShips)
	g.turnsTaken = 0
	g.gameStartedAt = time.Now()
	g.gameFinishedAt = time.Time{}
	g.revealStartAt = time.Time{}
	g.recalculateCellSize()
	g.restartButton.scale = g.restartButton.baseScale
	g.restartButton.pressed = false
	g.mainMenuButton.scale = g.mainMenuButton.baseScale
	g.mainMenuButton.pressed = false
	g.state = gameScreen

	// NOTE : Hint the runtime to reclaim transient allocations from the previous
	// session (old board slices, ship coordinate slices, candidate scratch growth, etc.) before the new game starts rendering.
	runtime.GC()
}

func (g *Game) updateHome(mousePressed, justPressed bool) {
	x, y := ebiten.CursorPosition()
	for key, button := range g.homeButtons {
		hovered := button.contains(x, y)
		button.update(hovered, mousePressed)

		if !hovered || !justPressed {
			continue
		}

		switch key {
		case "playButton":
			g.startNewGame()
		case "howToPlayButton":
			g.state = howToPlayScreen
		case "settingsButton":
			g.state = settingsScreen
		case "quitButton":
			g.shouldQuit = true
		}
	}
}

func (g *Game) updateHowToPlay(mousePressed, justPressed bool) {
	x, y := ebiten.CursorPosition()
	mainMenuHovered := g.mainMenuButton.contains(x, y)
	playHovered := g.playButton.contains(x, y)

	g.mainMenuButton.update(mainMenuHovered, mousePressed)
	g.playButton.update(playHovered, mousePressed)

	if justPressed {
		if mainMenuHovered {
			g.state = homeScreen
		} else if playHovered {
			g.startNewGame()
		}
	}
}

func (g *Game) updateSettings(mousePressed, justPressed bool) {
	x, y := ebiten.CursorPosition()

	playHovered := g.playButton.contains(x, y)
	mainMenuHovered := g.mainMenuButton.contains(x, y)
	g.playButton.update(playHovered, mousePressed)
	g.mainMenuButton.update(mainMenuHovered, mousePressed)

	shipsMinusHovered := g.shipsMinus.contains(x, y)
	shipsPlusHovered := g.shipsPlus.contains(x, y)
	gridMinusHovered := g.gridMinus.contains(x, y)
	gridPlusHovered := g.gridPlus.contains(x, y)
	g.shipsMinus.update(shipsMinusHovered, mousePressed)
	g.shipsPlus.update(shipsPlusHovered, mousePressed)
	g.gridMinus.update(gridMinusHovered, mousePressed)
	g.gridPlus.update(gridPlusHovered, mousePressed)

	if justPressed {
		fx, fy := float64(x), float64(y)
		switch {
		case shipsMinusHovered:
			g.commitInputBuffer()
			g.decrementShips()
		case shipsPlusHovered:
			g.commitInputBuffer()
			g.incrementShips()
		case gridMinusHovered:
			g.commitInputBuffer()
			g.decrementFocusedGrid()
		case gridPlusHovered:
			g.commitInputBuffer()
			g.incrementFocusedGrid()
		case pointInRect(fx, fy, 100, 250, 400, 80):
			g.setFocus("ships")
		case pointInRect(fx, fy, 102, 353, 197, 68):
			g.setFocus("gridHeight")
		case pointInRect(fx, fy, 299, 353, 196, 68):
			g.setFocus("gridWidth")
		case playHovered:
			g.commitInputBuffer()
			g.startNewGame()
		case mainMenuHovered:
			g.commitInputBuffer()
			g.clearFocus()
			g.persistSettings()
			g.state = homeScreen
		default:
			g.clearFocus()
		}
	}

	g.handleSettingsKeyboard()
}

func (g *Game) updateGame(mousePressed, justPressed bool) {
	mx, my := ebiten.CursorPosition()

	mainMenuHovered := g.mainMenuButton.contains(mx, my)
	restartHovered := g.restartButton.contains(mx, my)

	g.mainMenuButton.update(mainMenuHovered, mousePressed)
	g.restartButton.update(restartHovered, mousePressed)

	if g.gameBoard == nil || !justPressed {
		return
	}

	if mainMenuHovered {
		g.state = homeScreen
		return
	}
	if restartHovered {
		g.startNewGame()
		return
	}

	cellX := int((float64(mx) - g.gridOriginX) / g.cellSize)
	cellY := int((float64(my) - g.gridOriginY) / g.cellSize)

	if cellX < 0 || cellY < 0 {
		return
	}

	if cellX >= int(g.gameBoard.GetBoardWidth()) || cellY >= int(g.gameBoard.GetBoardHeight()) {
		return
	}

	gameState := gameService.GamePlay(uint8(cellY), uint8(cellX), g.player, g.gameBoard, g.numberOfShips)
	g.turnsTaken = g.player.GetNoOfAttempts()

	if gameState == gameService.End {
		g.gameFinishedAt = time.Now()
		g.revealStartAt = time.Now()
		g.state = revealScreen
	}
}

func (g *Game) updateReveal(mousePressed, justPressed bool) {
	if time.Since(g.revealStartAt) >= revealDuration {
		g.state = resultScreen
		return
	}
	if justPressed {
		g.state = resultScreen
	}
}

func (g *Game) updateResult(mousePressed, justPressed bool) {
	x, y := ebiten.CursorPosition()
	mainMenuHovered := g.mainMenuButton.contains(x, y)
	restartHovered := g.restartButton.contains(x, y)

	g.mainMenuButton.update(mainMenuHovered, mousePressed)
	g.restartButton.update(restartHovered, mousePressed)

	if justPressed {
		if mainMenuHovered {
			g.state = homeScreen
		} else if restartHovered {
			g.startNewGame()
		}
	}
}

func (g *Game) Update() error {
	mousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	justPressed := mousePressed && !g.prevMousePressed
	g.prevMousePressed = mousePressed

	switch g.state {
	case homeScreen:
		g.updateHome(mousePressed, justPressed)
	case howToPlayScreen:
		g.updateHowToPlay(mousePressed, justPressed)
	case settingsScreen:
		g.updateSettings(mousePressed, justPressed)
	case gameScreen:
		g.updateGame(mousePressed, justPressed)
	case revealScreen:
		g.updateReveal(mousePressed, justPressed)
	case resultScreen:
		g.updateResult(mousePressed, justPressed)
	}

	if g.shouldQuit {
		return ebiten.Termination
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	backgroundTileOpts := &ebiten.DrawImageOptions{}
	backgroundTileOpts.GeoM.Scale(
		windowWidth/float64(images["backgroundTile"].Bounds().Dx()),
		windowHeight/float64(images["backgroundTile"].Bounds().Dy()),
	)
	screen.DrawImage(images["backgroundTile"], backgroundTileOpts)

	switch g.state {
	case homeScreen:
		g.drawHome(screen)
	case howToPlayScreen:
		g.drawHowToPlay(screen)
	case settingsScreen:
		g.drawSettings(screen)
	case gameScreen:
		g.drawGame(screen)
	case revealScreen:
		g.drawReveal(screen)
	case resultScreen:
		g.drawResult(screen)
	}
}

func (g *Game) drawHome(screen *ebiten.Image) {
	logoOpts := &ebiten.DrawImageOptions{}
	logoOpts.GeoM.Scale(0.25, 0.25)
	logoOpts.GeoM.Translate(200, -15)
	screen.DrawImage(images["battleshipsLogo"], logoOpts)

	g.homeButtons["playButton"].draw(screen)
	g.homeButtons["howToPlayButton"].draw(screen)
	g.homeButtons["settingsButton"].draw(screen)
	g.homeButtons["quitButton"].draw(screen)
}

func (g *Game) drawGame(screen *ebiten.Image) {
	backgroundBannerOpts := &ebiten.DrawImageOptions{}
	backgroundBannerOpts.GeoM.Scale(0.8, 0.10)
	backgroundBannerOpts.GeoM.Translate(18, 8)
	screen.DrawImage(images["battleshipsBackgroundBanner"], backgroundBannerOpts)

	textOpts := &text.DrawOptions{
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign:   text.AlignCenter,
			SecondaryAlign: text.AlignCenter,
		},
	}
	textOpts.GeoM.Translate(275, 35)

	face := &text.GoTextFace{
		Source: fontSource,
		Size:   24,
	}
	text.Draw(screen, "GAME BOARD - Click a cell to attack", face, textOpts)

	backgroundBannerOpts = &ebiten.DrawImageOptions{}
	backgroundBannerOpts.GeoM.Reset()
	backgroundBannerOpts.GeoM.Scale(0.50, 0.10)
	backgroundBannerOpts.GeoM.Translate(485, 165)
	screen.DrawImage(images["battleshipsBackgroundBanner"], backgroundBannerOpts)

	textOpts.GeoM.Reset()
	textOpts.GeoM.Translate(650, 195)
	text.Draw(screen, "Find all ships to win", face, textOpts)
	textOpts.GeoM.Reset()

	face.Size = 35
	turnOpts := &ebiten.DrawImageOptions{}
	turnOpts.GeoM.Scale(0.12, 0.12)
	turnOpts.GeoM.Translate(535, 285)
	screen.DrawImage(images["turnsBox"], turnOpts)

	textOpts.GeoM.Reset()
	textOpts.GeoM.Translate(705, 315)
	text.Draw(screen, itoa(int(g.turnsTaken)), face, textOpts)
	textOpts.GeoM.Reset()

	timeOpts := &ebiten.DrawImageOptions{}
	timeOpts.GeoM.Scale(0.15, 0.15)
	timeOpts.GeoM.Translate(535, 345)
	screen.DrawImage(images["timeBox"], timeOpts)

	textOpts.GeoM.Translate(705, 375)
	text.Draw(screen, g.elapsedTimeText(), face, textOpts)

	if g.gameBoard == nil {
		return
	}

	for row := 0; row < int(g.gameBoard.GetBoardHeight()); row++ {
		for col := 0; col < int(g.gameBoard.GetBoardWidth()); col++ {
			status := g.gameBoard.GetStatus(uint8(row), uint8(col))
			x := g.gridOriginX + float64(col)*g.cellSize
			y := g.gridOriginY + float64(row)*g.cellSize

			cellImg := images["unopenedCell"]
			tintHit := false
			if status == board.Miss || status == board.Hit {
				cellImg = images["missedCell"]
			}
			if status == board.Hit {
				tintHit = true
			}

			opts := &ebiten.DrawImageOptions{}
			sx := g.cellSize / float64(cellImg.Bounds().Dx())
			sy := g.cellSize / float64(cellImg.Bounds().Dy())
			opts.GeoM.Scale(sx, sy)
			opts.GeoM.Translate(x, y)
			if tintHit {
				opts.ColorScale.Scale(1.2, 0.45, 0.45, 1.0)
			}
			screen.DrawImage(cellImg, opts)
		}
	}

	// g.drawShips(screen)
	g.mainMenuButton.draw(screen)
	g.restartButton.draw(screen)
}

func (g *Game) drawResult(screen *ebiten.Image) {
	backgroundBannerOpts := &ebiten.DrawImageOptions{}
	backgroundBannerOpts.GeoM.Scale(0.75, 0.75)
	backgroundBannerOpts.GeoM.Translate(165, 85)
	screen.DrawImage(images["battleshipsBackgroundBanner"], backgroundBannerOpts)

	textOpts := &text.DrawOptions{}
	face := &text.GoTextFace{
		Source: fontSource,
		Size:   25,
	}

	// 1.> Creating "Highlight" Effect
	textOpts.ColorScale.Reset()
	textOpts.ColorScale.ScaleWithColor(highlightColor)
	textOpts.GeoM.Reset()
	textOpts.GeoM.Translate(305, 151) // Original Y + 1
	text.Draw(screen, "MISSION COMPLETE", face, textOpts)

	// 2.> Creating "Engraving" Effect
	textOpts.ColorScale.Reset()
	textOpts.ColorScale.ScaleWithColor(engravedColor)
	textOpts.GeoM.Reset()
	textOpts.GeoM.Translate(305, 150) // Original Y
	text.Draw(screen, "MISSION COMPLETE", face, textOpts)

	face.Size = 20
	textOpts.GeoM.Reset()
	textOpts.GeoM.Translate(210, 190)
	textOpts.ColorScale.Reset()
	text.Draw(screen, "Your fleet has cleared all enemy ships.", face, textOpts)

	turnOpts := &ebiten.DrawImageOptions{}
	textOpts.GeoM.Reset()
	turnOpts.GeoM.Scale(0.15, 0.20)
	turnOpts.GeoM.Translate(250, 230)
	screen.DrawImage(images["turnsBox"], turnOpts)

	face.Size = 35
	textOpts.GeoM.Reset()
	textOpts.GeoM.Translate(455, 263)
	text.Draw(screen, itoa(int(g.turnsTaken)), face, textOpts)

	timeOpts := &ebiten.DrawImageOptions{}
	textOpts.GeoM.Reset()
	timeOpts.GeoM.Scale(0.25, 0.25)
	timeOpts.GeoM.Translate(210, 335)
	screen.DrawImage(images["timeBox"], timeOpts)

	textOpts.GeoM.Reset()
	textOpts.GeoM.Translate(435, 368)
	text.Draw(screen, g.elapsedTimeText(), face, textOpts)

	g.mainMenuButton.draw(screen)
	g.restartButton.draw(screen)
}

func (g *Game) drawHowToPlay(screen *ebiten.Image) {
	backgroundBannerOpts := &ebiten.DrawImageOptions{}
	backgroundBannerOpts.GeoM.Scale(0.76, 0.5)
	backgroundBannerOpts.GeoM.Translate(-25, 0)
	screen.DrawImage(images["battleshipBackground"], backgroundBannerOpts)

	textOpts := &text.DrawOptions{}
	face := &text.GoTextFace{
		Source: fontSource,
		Size:   25,
	}

	// 1.> Creating "Highlight" Effect
	textOpts.ColorScale.Reset()
	textOpts.ColorScale.ScaleWithColor(highlightColor)
	textOpts.GeoM.Reset()
	textOpts.GeoM.Translate(355, 46) // Original Y + 1
	text.Draw(screen, "HOW TO PLAY", face, textOpts)

	// 2.> Creating "Engraving" Effect
	textOpts.ColorScale.Reset()
	textOpts.ColorScale.ScaleWithColor(color.RGBA{66, 40, 14, 255})
	textOpts.GeoM.Reset()
	textOpts.GeoM.Translate(355, 45) // Original Y
	text.Draw(screen, "HOW TO PLAY", face, textOpts)

	// Reducing the font size for the text
	textOpts.GeoM.Reset()
	textOpts.ColorScale.Reset()
	face.Size = 20
	textOpts.ColorScale.ScaleWithColor(color.RGBA{40, 20, 10, 255})

	textOpts.GeoM.Translate(125, 155)
	text.Draw(screen, "1) Choose grid size and number of ships in Settings.", face, textOpts)
	textOpts.GeoM.Translate(0, 45)
	text.Draw(screen, "2) Start the game and click cells to attack enemy fleet.", face, textOpts)
	textOpts.GeoM.Translate(0, 45)
	text.Draw(screen, "3) There are 3 possibilities: unexplored, miss and hit.", face, textOpts)
	textOpts.GeoM.Translate(0, 45)
	text.Draw(screen, "4) Destroy all ship segments in minimum turns and time.", face, textOpts)
	textOpts.GeoM.Translate(0, 45)
	text.Draw(screen, "Ship visuals:", face, textOpts)
	textOpts.GeoM.Translate(0, 45)
	text.Draw(screen, "Size 1: single-cell small ship", face, textOpts)
	textOpts.GeoM.Translate(0, 45)
	text.Draw(screen, "Size 2: medium front + medium end", face, textOpts)
	textOpts.GeoM.Translate(0, 45)
	text.Draw(screen, "Size 3+: large front + large mids + large end", face, textOpts)

	g.mainMenuButton.draw(screen)
	g.playButton.draw(screen)
}

func (g *Game) drawSettings(screen *ebiten.Image) {
	backgroundBannerOpts := &ebiten.DrawImageOptions{}
	backgroundBannerOpts.GeoM.Scale(1.25, 1.25)
	backgroundBannerOpts.GeoM.Translate(7.5, 0)
	screen.DrawImage(images["battleshipsSettingsBackground"], backgroundBannerOpts)

	textOpts := &text.DrawOptions{}
	textOpts.GeoM.Translate(365, 70)
	face := &text.GoTextFace{
		Source: fontSource,
		Size:   25,
	}

	// 1.> Creating "Highlight" Effect
	textOpts.ColorScale.Reset()
	textOpts.ColorScale.ScaleWithColor(highlightColor)
	textOpts.GeoM.Reset()
	textOpts.GeoM.Translate(365, 71) // Original Y + 1
	text.Draw(screen, "SETTINGS", face, textOpts)

	// 2.> Creating "Engraving" Effect
	textOpts.ColorScale.Reset()
	textOpts.ColorScale.ScaleWithColor(color.RGBA{150, 111, 51, 255})
	textOpts.GeoM.Reset()
	textOpts.GeoM.Translate(365, 70) // Original Y
	text.Draw(screen, "SETTINGS", face, textOpts)

	// Reducing the font size for the text
	textOpts.GeoM.Reset()
	textOpts.ColorScale.Reset()
	face.Size = 20
	// textOpts.ColorScale.ScaleWithColor(color.RGBA{193, 154, 107, 255})

	textOpts.GeoM.Translate(120, 110)
	text.Draw(screen, "Configure fleet and grid before starting.", face, textOpts)
	textOpts.GeoM.Translate(0, 30)
	text.Draw(screen, "Rule: min(height,width) >= ships + 1", face, textOpts)
	textOpts.GeoM.Translate(0, 30)
	text.Draw(screen, "Click a field and type, or use Tab / Arrow keys / Enter / +-.", face, textOpts)

	shipsBoxOpts := &ebiten.DrawImageOptions{}
	shipsBoxOpts.GeoM.Translate(100, 250)
	screen.DrawImage(images["numberOfShipsBox"], shipsBoxOpts)
	g.drawInputValue(screen, "ships", 330, 282, g.numberOfShips)
	if g.focusedField == "ships" {
		drawFocusRing(screen, 252, 265, 177, 47)
	}

	controlButtonsOpts := &ebiten.DrawImageOptions{}
	controlButtonsOpts.GeoM.Scale(0.5, 0.45)
	controlButtonsOpts.GeoM.Translate(500, 250)
	screen.DrawImage(images["controlButtons"], controlButtonsOpts)
	g.shipsMinus.draw(screen)
	g.shipsPlus.draw(screen)

	gridSizeBoxOpts := &ebiten.DrawImageOptions{}
	gridSizeBoxOpts.GeoM.Translate(120, 415)
	gridSizeBoxOpts.GeoM.Scale(0.85, 0.85)
	screen.DrawImage(images["gridSizeBox"], gridSizeBoxOpts)

	g.drawInputValue(screen, "gridHeight", 290, 380, int(g.gridHeight))
	g.drawInputValue(screen, "gridWidth", 415, 380, int(g.gridWidth))
	switch g.focusedField {
	case "gridHeight":
		drawFocusRing(screen, 272, 365, 72, 40)
	case "gridWidth":
		drawFocusRing(screen, 398, 365, 72, 40)
	}

	gridControlOpts := &ebiten.DrawImageOptions{}
	gridControlOpts.GeoM.Scale(0.5, 0.425)
	gridControlOpts.GeoM.Translate(515, 350)
	screen.DrawImage(images["controlButtons"], gridControlOpts)
	g.gridMinus.draw(screen)
	g.gridPlus.draw(screen)

	adjustLabel := "Adjusting: Height"
	if g.lastFocusedGrid == "gridWidth" {
		adjustLabel = "Adjusting: Width"
	}
	textOpts.GeoM.Translate(690, 380)
	text.Draw(screen, adjustLabel, face, textOpts)

	g.playButton.draw(screen)
	g.mainMenuButton.draw(screen)
}

func (g *Game) drawInputValue(screen *ebiten.Image, field string, x, y int, value int) {
	label := ""
	if g.focusedField == field {
		label = g.inputBuffer
		if (time.Now().UnixMilli()/500)%2 == 0 {
			label += "_"
		}
	} else {
		label = itoa(value)
	}

	textOpts := &text.DrawOptions{}
	textOpts.GeoM.Translate(float64(x), float64(y))
	face := &text.GoTextFace{
		Source: fontSource,
		Size:   24,
	}

	text.Draw(screen, label, face, textOpts)
}

func (g *Game) drawShipsRevealed(screen *ebiten.Image, onlyHit bool) {
	for _, s := range ship.SnapshotRenderedShips() {
		for idx, c := range s.Coords {
			if onlyHit && g.gameBoard.GetStatus(c.Y, c.X) != board.Hit {
				continue
			}
			g.drawShipSegment(screen, s, idx, c, onlyHit)
		}
	}
}

func (g *Game) drawShipSegment(screen *ebiten.Image, s ship.RenderedShip, idx int, c ship.RenderCoordinate, tintRed bool) {
	key := shipImageForSegment(s.Size, idx, len(s.Coords))
	img := images[key]
	imgW := float64(img.Bounds().Dx())
	imgH := float64(img.Bounds().Dy())

	cellCenterX := g.gridOriginX + (float64(c.X)+0.5)*g.cellSize
	cellCenterY := g.gridOriginY + (float64(c.Y)+0.5)*g.cellSize

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(-imgW/2, -imgH/2)

	if s.IsVertical && s.Size > 1 {
		opts.GeoM.Rotate(math.Pi / 2)
		opts.GeoM.Scale(g.cellSize/imgH, g.cellSize/imgW)
	} else {
		opts.GeoM.Scale(g.cellSize/imgW, g.cellSize/imgH)
	}

	opts.GeoM.Translate(cellCenterX, cellCenterY)

	if tintRed {
		opts.ColorScale.Scale(1.25, 0.35, 0.35, 1.0)
	}
	screen.DrawImage(img, opts)
}

func (g *Game) drawReveal(screen *ebiten.Image) {
	backgroundBannerOpts := &ebiten.DrawImageOptions{}
	backgroundBannerOpts.GeoM.Scale(0.8, 0.10)
	backgroundBannerOpts.GeoM.Translate(18, 8)
	screen.DrawImage(images["battleshipsBackgroundBanner"], backgroundBannerOpts)

	textOpts := &text.DrawOptions{}
	face := &text.GoTextFace{
		Source: fontSource,
		Size:   24,
	}

	textOpts.ColorScale.Reset()
	textOpts.ColorScale.ScaleWithColor(highlightColor)
	textOpts.GeoM.Reset()
	textOpts.GeoM.Translate(220, 36)
	text.Draw(screen, "ENEMY FLEET REVEALED", face, textOpts)

	textOpts.ColorScale.Reset()
	textOpts.ColorScale.ScaleWithColor(engravedColor)
	textOpts.GeoM.Reset()
	textOpts.GeoM.Translate(220, 35)
	text.Draw(screen, "ENEMY FLEET REVEALED", face, textOpts)

	face.Size = 18
	textOpts.ColorScale.Reset()
	textOpts.GeoM.Reset()
	textOpts.GeoM.Translate(530, 220)
	text.Draw(screen, "Showing all ship", face, textOpts)
	textOpts.GeoM.Reset()
	textOpts.GeoM.Translate(530, 245)
	text.Draw(screen, "positions...", face, textOpts)

	remaining := revealDuration - time.Since(g.revealStartAt)
	if remaining < 0 {
		remaining = 0
	}
	secs := int(remaining/time.Second) + 1
	if secs < 1 {
		secs = 1
	}
	textOpts.GeoM.Reset()
	textOpts.GeoM.Translate(530, 295)
	text.Draw(screen, "Next in "+itoa(secs)+"s", face, textOpts)

	textOpts.GeoM.Reset()
	textOpts.GeoM.Translate(530, 335)
	text.Draw(screen, "(click to skip)", face, textOpts)

	if g.gameBoard == nil {
		return
	}

	for row := 0; row < int(g.gameBoard.GetBoardHeight()); row++ {
		for col := 0; col < int(g.gameBoard.GetBoardWidth()); col++ {
			status := g.gameBoard.GetStatus(uint8(row), uint8(col))
			x := g.gridOriginX + float64(col)*g.cellSize
			y := g.gridOriginY + float64(row)*g.cellSize

			cellImg := images["unopenedCell"]
			if status == board.Miss {
				cellImg = images["missedCell"]
			}

			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Scale(
				g.cellSize/float64(cellImg.Bounds().Dx()),
				g.cellSize/float64(cellImg.Bounds().Dy()),
			)
			opts.GeoM.Translate(x, y)
			screen.DrawImage(cellImg, opts)
		}
	}

	g.drawShipsRevealed(screen, false)
}

func main() {
	masterSheet = mustLoadAssets("assets/master-sheet.png")
	fontSource = mustLoadFont()
	images = fetchImagesFromMasterAsset(masterSheet)

	battleshipsLogo := mustLoadAssets("assets/battleships-logo.png")
	battleshipSmall := mustLoadAssets("assets/battleship-small.png")
	battleshipMediumFront := mustLoadAssets("assets/battleship-medium-front.png")
	battleshipMediumEnd := mustLoadAssets("assets/battleship-medium-end.png")
	battleshipLargeFront := mustLoadAssets("assets/battleship-large-front.png")
	battleshipLargeMid := mustLoadAssets("assets/battleship-large-mid.png")
	battleshipLargeEnd := mustLoadAssets("assets/battleship-large-end.png")
	turnsBox := mustLoadAssets("assets/turns-box.png")
	timeBox := mustLoadAssets("assets/time-box.png")
	controlButtons := mustLoadAssets("assets/control-buttons.png")
	battleshipBackground := mustLoadAssets("assets/battleships-background.png")
	battleshipsSettingsBackground := mustLoadAssets("assets/battleships-settings-background.png")
	battleshipsBackgroundBanner := mustLoadAssets("assets/battleships-background-banner.png")

	images["battleshipsLogo"] = battleshipsLogo
	images["battleshipSmall"] = battleshipSmall
	images["battleshipMediumFront"] = battleshipMediumFront
	images["battleshipMediumEnd"] = battleshipMediumEnd
	images["battleshipLargeFront"] = battleshipLargeFront
	images["battleshipLargeMid"] = battleshipLargeMid
	images["battleshipLargeEnd"] = battleshipLargeEnd
	images["turnsBox"] = turnsBox
	images["timeBox"] = timeBox
	images["controlButtons"] = controlButtons
	images["battleshipBackground"] = battleshipBackground
	images["battleshipsSettingsBackground"] = battleshipsSettingsBackground
	images["battleshipsBackgroundBanner"] = battleshipsBackgroundBanner

	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowTitle("BattleShips 🚢")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
