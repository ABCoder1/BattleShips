package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type gameConfig struct {
	NumberOfShips int   `json:"number_of_ships"`
	GridHeight    uint8 `json:"grid_height"`
	GridWidth     uint8 `json:"grid_width"`
}

func configPath() (string, error) {
	dir, cfgErr := os.UserConfigDir()
	if cfgErr != nil {
		return "", cfgErr
	}
	subdir := filepath.Join(dir, "battleships")
	if mkErr := os.MkdirAll(subdir, 0o755); mkErr != nil {
		return "", mkErr
	}
	return filepath.Join(subdir, "config.json"), nil
}

func loadConfig() *gameConfig {
	path, err := configPath()
	if err != nil {
		return nil
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		return nil
	}
	var cfg gameConfig
	if jsonErr := json.Unmarshal(data, &cfg); jsonErr != nil {
		return nil
	}
	if cfg.NumberOfShips <= 0 || cfg.GridHeight == 0 || cfg.GridWidth == 0 {
		return nil
	}
	return &cfg
}

func saveConfig(cfg gameConfig) {
	path, err := configPath()
	if err != nil {
		return
	}
	data, marshalErr := json.MarshalIndent(cfg, "", "  ")
	if marshalErr != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o644)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return windowWidth, windowHeight
}

func (g *Game) handleSettingsKeyboard() {
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.cycleFocus()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.commitInputBuffer()
		g.clearFocus()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		g.commitInputBuffer()
		g.incrementFocused()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		g.commitInputBuffer()
		g.decrementFocused()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && g.focusedField != "" && len(g.inputBuffer) > 0 {
		g.inputBuffer = g.inputBuffer[:len(g.inputBuffer)-1]
	}

	if g.focusedField != "" {
		for _, r := range ebiten.AppendInputChars(nil) {
			if r >= '0' && r <= '9' && len(g.inputBuffer) < 2 {
				g.inputBuffer += string(r)
			}
		}
	}
}

func (g *Game) setFocus(field string) {
	if g.focusedField == field {
		return
	}
	g.commitInputBuffer()
	g.focusedField = field
	g.inputBuffer = ""
	if field == "gridHeight" || field == "gridWidth" {
		g.lastFocusedGrid = field
	}
}

func (g *Game) clearFocus() {
	g.commitInputBuffer()
	g.focusedField = ""
	g.inputBuffer = ""
}

func (g *Game) cycleFocus() {
	switch g.focusedField {
	case "":
		g.setFocus("ships")
	case "ships":
		g.setFocus("gridHeight")
	case "gridHeight":
		g.setFocus("gridWidth")
	case "gridWidth":
		g.setFocus("ships")
	}
}

func (g *Game) commitInputBuffer() {
	if g.focusedField == "" || g.inputBuffer == "" {
		return
	}
	value, parseErr := strconv.Atoi(g.inputBuffer)
	g.inputBuffer = ""
	if parseErr != nil {
		return
	}
	switch g.focusedField {
	case "ships":
		if value < 1 {
			value = 1
		}
		maxShips := int(minUint8(g.gridHeight, g.gridWidth)) - 1
		if maxShips < 1 {
			maxShips = 1
		}
		if value > maxShips {
			value = maxShips
		}
		g.numberOfShips = value
	case "gridHeight":
		if value > 20 {
			value = 20
		}
		minVal := g.numberOfShips + 1
		if value < minVal {
			value = minVal
		}
		g.gridHeight = uint8(value)
	case "gridWidth":
		if value > 20 {
			value = 20
		}
		minVal := g.numberOfShips + 1
		if value < minVal {
			value = minVal
		}
		g.gridWidth = uint8(value)
	}
}

func (g *Game) incrementShips() {
	maxByGrid := int(minUint8(g.gridHeight, g.gridWidth)) - 1
	if maxByGrid < 1 {
		maxByGrid = 1
	}
	if g.numberOfShips < maxByGrid {
		g.numberOfShips++
	}
}

func (g *Game) decrementShips() {
	if g.numberOfShips > 1 {
		g.numberOfShips--
	}
}

func (g *Game) incrementGridHeight() {
	if g.gridHeight < 20 {
		g.gridHeight++
	}
	g.normalizeShipsToGrid()
}

func (g *Game) decrementGridHeight() {
	minAllowed := uint8(g.numberOfShips + 1)
	if g.gridHeight > minAllowed {
		g.gridHeight--
	}
}

func (g *Game) incrementGridWidth() {
	if g.gridWidth < 20 {
		g.gridWidth++
	}
	g.normalizeShipsToGrid()
}

func (g *Game) decrementGridWidth() {
	minAllowed := uint8(g.numberOfShips + 1)
	if g.gridWidth > minAllowed {
		g.gridWidth--
	}
}

func (g *Game) normalizeShipsToGrid() {
	maxShips := int(minUint8(g.gridHeight, g.gridWidth)) - 1
	if maxShips < 1 {
		maxShips = 1
	}
	if g.numberOfShips > maxShips {
		g.numberOfShips = maxShips
	}
}

func (g *Game) recalculateCellSize() {
	maxDim := int(g.gridHeight)
	if int(g.gridWidth) > maxDim {
		maxDim = int(g.gridWidth)
	}
	if maxDim < 1 {
		maxDim = 1
	}
	g.cellSize = 420.0 / float64(maxDim)
	if g.cellSize > 48 {
		g.cellSize = 48
	}
	if g.cellSize < 18 {
		g.cellSize = 18
	}
}

func (g *Game) elapsedTimeText() string {
	if g.gameStartedAt.IsZero() {
		return "00:00"
	}

	end := time.Now()
	if !g.gameFinishedAt.IsZero() {
		end = g.gameFinishedAt
	}
	d := end.Sub(g.gameStartedAt)
	totalSec := int(d.Seconds())
	if totalSec < 0 {
		totalSec = 0
	}
	min := totalSec / 60
	sec := totalSec % 60
	return pad2(min) + ":" + pad2(sec)
}

func (g *Game) incrementFocused() {
	switch g.focusedField {
	case "ships":
		g.incrementShips()
	case "gridHeight":
		g.incrementGridHeight()
	case "gridWidth":
		g.incrementGridWidth()
	}
}

func (g *Game) decrementFocused() {
	switch g.focusedField {
	case "ships":
		g.decrementShips()
	case "gridHeight":
		g.decrementGridHeight()
	case "gridWidth":
		g.decrementGridWidth()
	}
}

func (g *Game) incrementFocusedGrid() {
	if g.lastFocusedGrid == "gridWidth" {
		g.incrementGridWidth()
		return
	}
	g.incrementGridHeight()
}

func (g *Game) decrementFocusedGrid() {
	if g.lastFocusedGrid == "gridWidth" {
		g.decrementGridWidth()
		return
	}
	g.decrementGridHeight()
}

// NOTE : drawFocusRing renders a 4-edge highlight rectangle around the given bounds.
// It uses vector.DrawFilledRect (no image allocations) so it is safe to call
// every frame. The previous implementation allocated two GPU-backed *ebiten.Image instances per invocation, which leaked GPU memory while the user sat on the Settings screen with a field focused (~120 textures/sec at 60 FPS).
func drawFocusRing(screen *ebiten.Image, x, y, w, h float64) {
	const thickness = 2.0
	ringColor := color.RGBA{R: 255, G: 215, B: 0, A: 255}

	vector.FillRect(screen, float32(x), float32(y), float32(w), thickness, ringColor, false)
	vector.FillRect(screen, float32(x), float32(y+h-thickness), float32(w), thickness, ringColor, false)
	vector.FillRect(screen, float32(x), float32(y), thickness, float32(h), ringColor, false)
	vector.FillRect(screen, float32(x+w-thickness), float32(y), thickness, float32(h), ringColor, false)
}

func minUint8(a, b uint8) uint8 {
	if a < b {
		return a
	}
	return b
}

func pad2(value int) string {
	if value < 10 {
		return "0" + itoa(value)
	}
	return itoa(value)
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	isNegative := v < 0
	if isNegative {
		v = -v
	}
	var digits [20]byte
	i := len(digits)
	for v > 0 {
		i--
		digits[i] = byte('0' + (v % 10))
		v /= 10
	}
	if isNegative {
		i--
		digits[i] = '-'
	}
	return string(digits[i:])
}

func mustLoadAssets(name string) *ebiten.Image {
	f, err := assets.Open(name)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	return ebiten.NewImageFromImage(img)
}

func mustLoadFont() *text.GoTextFaceSource {
	fontData, err := assets.ReadFile("assets/SpecialElite-Regular.ttf")
	if err != nil {
		log.Fatal(err)
	}

	fontSource, err := text.NewGoTextFaceSource(bytes.NewReader(fontData))
	if err != nil {
		log.Fatal(err)
	}
	return fontSource
}

func shipImageForSegment(size, index, total int) string {
	switch {
	case size <= 1:
		return "battleshipSmall"
	case size == 2:
		if index == 0 {
			return "battleshipMediumFront"
		}
		return "battleshipMediumEnd"
	default:
		if index == 0 {
			return "battleshipLargeFront"
		}
		if index == total-1 {
			return "battleshipLargeEnd"
		}
		return "battleshipLargeMid"
	}
}

func pointInRect(px, py, x, y, w, h float64) bool {
	return px >= x && px <= x+w && py >= y && py <= y+h
}

func fetchImagesFromMasterAsset(masterSheet *ebiten.Image) map[string]*ebiten.Image {
	playButton := image.Rect(68, 96, 395, 175)
	howToPlayButton := image.Rect(65, 195, 395, 275)
	settingsButton := image.Rect(65, 295, 395, 375)
	mainMenuButton := image.Rect(531, 1260, 935, 1350)
	quitButton := image.Rect(65, 395, 395, 470)
	backgroundTile := image.Rect(62, 875, 345, 1145)
	unOpenedCell := image.Rect(65, 515, 220, 660)
	missedCell := image.Rect(255, 515, 410, 660)
	numberOfShipsInputBox := image.Rect(565, 510, 940, 590)
	gridSizeInputBox := image.Rect(515, 610, 980, 690)
	restartButton := image.Rect(90, 1260, 465, 1350)

	return map[string]*ebiten.Image{
		"playButton":       masterSheet.SubImage(playButton).(*ebiten.Image),
		"howToPlayButton":  masterSheet.SubImage(howToPlayButton).(*ebiten.Image),
		"restartButton":    masterSheet.SubImage(restartButton).(*ebiten.Image),
		"settingsButton":   masterSheet.SubImage(settingsButton).(*ebiten.Image),
		"mainMenuButton":   masterSheet.SubImage(mainMenuButton).(*ebiten.Image),
		"quitButton":       masterSheet.SubImage(quitButton).(*ebiten.Image),
		"backgroundTile":   masterSheet.SubImage(backgroundTile).(*ebiten.Image),
		"unopenedCell":     masterSheet.SubImage(unOpenedCell).(*ebiten.Image),
		"missedCell":       masterSheet.SubImage(missedCell).(*ebiten.Image),
		"numberOfShipsBox": masterSheet.SubImage(numberOfShipsInputBox).(*ebiten.Image),
		"gridSizeBox":      masterSheet.SubImage(gridSizeInputBox).(*ebiten.Image),
	}
}
