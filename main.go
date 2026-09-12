package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	cellSize  = 40
	margin    = 1
	boardSize = 10
)

var (
	cellStep = cellSize + margin
	boardPx  = cellStep*boardSize + margin
	designW  = cellStep*2*boardSize + margin
	designH  = boardPx
)

var (
	black     = color.RGBA{0, 0, 0, 255}
	white     = color.RGBA{255, 255, 255, 255}
	darkGrey  = color.RGBA{25, 25, 30, 255}
	ghostBlue = color.RGBA{0xAA, 0xDD, 0xFF, 255}
)

var undoRect = image.Rect(designW-350, designH-80, designW-350+140, designH-80+60)
var resetRect = image.Rect(designW-200, designH-80, designW-200+140, designH-80+60)

// Each block is a set of [row, col] offsets relative to its anchor cell.
var blocks = [][][2]int{
	{{0, 0}, {1, 0}, {2, 0}, {3, 0}}, // I shape
	{{0, 0}, {0, 1}, {0, 2}, {0, 3}}, // I horizontal shape
	{{0, 0}, {0, 1}, {0, 2}, {1, 2}}, // L shape
	{{0, 0}, {1, 0}, {2, 0}, {0, 1}}, // L horizontal shape
	{{0, 0}, {0, 1}, {0, 2}, {1, 0}}, // J shape
	{{0, 0}, {1, 0}, {2, 0}, {2, 1}}, // J horizontal shape
	{{0, 0}, {0, 1}, {1, 1}, {1, 0}}, // O shape
	{{0, 0}, {0, 1}, {1, 1}, {1, 2}}, // S shape
	{{0, 1}, {0, 2}, {1, 0}, {1, 1}}, // Z shape
	{{0, 0}, {0, 1}, {0, 2}, {1, 1}}, // T shape
}

func generateBlock() [][2]int {
	block := blocks[rand.IntN(len(blocks))]
	return copyBlock(block)
}

func copyBlock(block [][2]int) [][2]int {
	out := make([][2]int, len(block))
	copy(out, block)
	return out
}

func emptyGrid() [boardSize][boardSize]bool {
	return [boardSize][boardSize]bool{}
}

type Game struct {
	grid            [boardSize][boardSize]bool
	prevGrid        [boardSize][boardSize]bool
	ghostGrid       [boardSize][boardSize]bool
	currentBlock    [][2]int
	prevBlock       [][2]int
	nextBlock       [][2]int
	score           int
	prevScore       int
	currentPosition [2]int
	nextPosition    [2]int

	canvas  *ebiten.Image
	scale   float64
	offsetX float64
	offsetY float64
}

func newGame() *Game {
	g := &Game{
		currentPosition: [2]int{1, 10},
		nextPosition:    [2]int{1, 15},
	}
	g.currentBlock = generateBlock()
	g.prevBlock = copyBlock(g.currentBlock)
	g.nextBlock = generateBlock()
	return g
}

func (g *Game) canPlaceBlock(block [][2]int, position [2]int) bool {
	for _, b := range block {
		row := position[0] + b[0]
		col := position[1] + b[1]
		if row < 0 || row >= boardSize || col < 0 || col >= boardSize {
			return false
		}
		if g.grid[row][col] {
			return false
		}
	}
	return true
}

func (g *Game) addBlock(block [][2]int, position [2]int) {
	g.prevGrid = g.grid
	g.prevScore = g.score
	for _, b := range block {
		row := position[0] + b[0]
		col := position[1] + b[1]
		g.grid[row][col] = true
	}
}

// clearLines removes completely filled rows and columns, shifting the
// remaining cells toward the top and left. Returns the number cleared.
func (g *Game) clearLines() int {
	cleared := 0

	// Remove full rows, shifting the rest toward the bottom.
	var keptRows [][boardSize]bool
	for row := 0; row < boardSize; row++ {
		full := true
		for col := 0; col < boardSize; col++ {
			if !g.grid[row][col] {
				full = false
				break
			}
		}
		if full {
			cleared++
			g.score += 10
		} else {
			keptRows = append(keptRows, g.grid[row])
		}
	}
	grid := emptyGrid()
	for i, row := range keptRows {
		grid[boardSize-len(keptRows)+i] = row
	}

	// Remove full columns, shifting the rest toward the right.
	var keptCols [][boardSize]bool
	for col := 0; col < boardSize; col++ {
		full := true
		for row := 0; row < boardSize; row++ {
			if !grid[row][col] {
				full = false
				break
			}
		}
		if full {
			cleared++
			g.score += 10
		} else {
			var column [boardSize]bool
			for row := 0; row < boardSize; row++ {
				column[row] = grid[row][col]
			}
			keptCols = append(keptCols, column)
		}
	}
	grid = emptyGrid()
	for i, column := range keptCols {
		dst := boardSize - len(keptCols) + i
		for row := 0; row < boardSize; row++ {
			grid[row][dst] = column[row]
		}
	}

	g.grid = grid
	return cleared
}

func (g *Game) resetGame() {
	g.currentBlock = generateBlock()
	g.prevBlock = copyBlock(g.currentBlock)
	g.nextBlock = generateBlock()
	g.score = 0
	g.prevScore = 0
	g.grid = emptyGrid()
	g.prevGrid = emptyGrid()
	g.ghostGrid = emptyGrid()
}

func (g *Game) undoStep() {
	g.grid = g.prevGrid
	g.score = g.prevScore
	g.currentBlock = g.prevBlock
}

// sceneMouse returns the mouse position in design coordinates.
func (g *Game) sceneMouse() (float64, float64) {
	mx, my := ebiten.CursorPosition()
	x := (float64(mx) - g.offsetX) / g.scale
	y := (float64(my) - g.offsetY) / g.scale
	return x, y
}

// cellAt returns the board cell under the given scene coordinates.
func cellAt(x, y float64) (row, col int, ok bool) {
	if x < 0 || y < 0 || x >= float64(boardPx) || y >= float64(boardPx) {
		return 0, 0, false
	}
	return int(y) / cellStep, int(x) / cellStep, true
}

func (g *Game) updateGhost() {
	x, y := g.sceneMouse()
	g.ghostGrid = emptyGrid()
	row, col, ok := cellAt(x, y)
	if !ok {
		return
	}
	for _, b := range g.currentBlock {
		r := row + b[0]
		c := col + b[1]
		if r >= 0 && r < boardSize && c >= 0 && c < boardSize {
			g.ghostGrid[r][c] = true
		}
	}
}

func (g *Game) Update() error {
	g.updateGhost()

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := g.sceneMouse()
		pos := image.Pt(int(x), int(y))

		switch {
		case pos.In(resetRect):
			g.resetGame()
		case pos.In(undoRect):
			g.undoStep()
		default:
			row, col, ok := cellAt(x, y)
			if !ok {
				return nil
			}
			position := [2]int{row, col}
			if g.canPlaceBlock(g.currentBlock, position) {
				g.addBlock(g.currentBlock, position)
				g.clearLines()
				g.prevBlock = g.currentBlock
				g.currentBlock = g.nextBlock
				g.nextBlock = generateBlock()
				g.score += 4
			}
		}
	}

	return nil
}

func drawRect(screen *ebiten.Image, x, y, w, h int, c color.Color) {
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), c, false)
}

func drawText(screen *ebiten.Image, s string, x, y int, c color.Color) {
	face := basicfont.Face7x13
	w := font.MeasureString(face, s).Ceil()
	img := ebiten.NewImage(w, 13)
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  fixed.P(0, 12),
	}
	d.DrawString(s)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(2, 2)
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)
}

func drawTextCentered(screen *ebiten.Image, s string, r image.Rectangle, c color.Color) {
	face := basicfont.Face7x13
	w := font.MeasureString(face, s).Ceil()
	x := r.Min.X + (r.Dx()-w*2)/2
	y := r.Min.Y + (r.Dy()-13*2)/2
	drawText(screen, s, x, y, c)
}

func drawButton(screen *ebiten.Image, r image.Rectangle, label string) {
	drawRect(screen, r.Min.X, r.Min.Y, r.Dx(), r.Dy(), darkGrey)
	vector.StrokeRect(screen, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), 2, white, false)
	drawTextCentered(screen, label, r, white)
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.canvas == nil {
		g.canvas = ebiten.NewImage(designW, designH)
	}
	canvas := g.canvas
	canvas.Fill(black)

	// Draw the field: a dark-grey grid on a black background.
	drawRect(canvas, 0, 0, boardPx, boardPx, darkGrey)
	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			drawRect(canvas, cellStep*col+margin, cellStep*row+margin, cellSize, cellSize, black)
		}
	}
	// Draw placed shapes (white) and the ghost shape (light blue).
	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			var c color.Color
			switch {
			case g.grid[row][col]:
				c = white
			case g.ghostGrid[row][col]:
				c = ghostBlue
			default:
				continue
			}
			drawRect(canvas, cellStep*col+margin, cellStep*row+margin, cellSize, cellSize, c)
		}
	}

	// Draw the right-hand panel.
	drawRect(canvas, boardPx, 0, designW-boardPx, designH, darkGrey)

	// Draw the current block preview.
	drawText(canvas, "Current:", designW-400, designH-400, white)
	for _, b := range g.currentBlock {
		row := g.currentPosition[0] + b[0]
		col := g.currentPosition[1] + b[1]
		drawRect(canvas, cellStep*col+margin+5, cellStep*row+margin, cellSize, cellSize, white)
	}

	// Draw the next block preview.
	drawText(canvas, "Next:", designW-200, designH-400, white)
	for _, b := range g.nextBlock {
		row := g.nextPosition[0] + b[0]
		col := g.nextPosition[1] + b[1]
		drawRect(canvas, cellStep*col+margin+5, cellStep*row+margin, cellSize, cellSize, white)
	}

	// Draw the score.
	drawText(canvas, fmt.Sprintf("Score: %d", g.score), designW-400, designH-120, white)

	// Draw the buttons.
	drawButton(canvas, undoRect, "UNDO")
	drawButton(canvas, resetRect, "RESET")

	// Scale and center the canvas in the window.
	screen.Fill(black)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(g.scale, g.scale)
	op.GeoM.Translate(g.offsetX, g.offsetY)
	screen.DrawImage(canvas, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	scale := math.Min(float64(outsideWidth)/float64(designW), float64(outsideHeight)/float64(designH))
	if scale < 0.01 {
		scale = 0.01
	}
	g.scale = scale
	g.offsetX = (float64(outsideWidth) - float64(designW)*scale) / 2
	g.offsetY = (float64(outsideHeight) - float64(designH)*scale) / 2
	return outsideWidth, outsideHeight
}

func main() {
	ebiten.SetWindowSize(designW, designH)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Tactris")
	if err := ebiten.RunGame(newGame()); err != nil {
		log.Fatal(err)
	}
}
