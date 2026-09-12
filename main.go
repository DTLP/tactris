package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
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
	black = color.RGBA{0, 0, 0, 255}
	white = color.RGBA{255, 255, 255, 255}
	grey  = color.RGBA{100, 100, 150, 255}
	lgrey = color.RGBA{215, 215, 230, 255}
	green = color.RGBA{0, 255, 0, 255}
)

var (
	windowW = (margin+cellSize)*20 + margin
	windowH = (margin+cellSize)*10 + margin
)

var undoRect = image.Rect(windowW-350, windowH-80, windowW-350+140, windowH-80+60)
var resetRect = image.Rect(windowW-200, windowH-80, windowW-200+140, windowH-80+60)

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

func (g *Game) updateGhost() {
	x, y := ebiten.CursorPosition()
	g.ghostGrid = emptyGrid()
	row := y / (cellSize + margin)
	col := x / (cellSize + margin)
	if row < 0 || row >= boardSize || col < 0 || col >= boardSize {
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
		x, y := ebiten.CursorPosition()
		pos := image.Pt(x, y)

		switch {
		case pos.In(resetRect):
			g.resetGame()
		case pos.In(undoRect):
			g.undoStep()
		default:
			row := y / (cellSize + margin)
			col := x / (cellSize + margin)
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
	drawRect(screen, r.Min.X, r.Min.Y, r.Dx(), r.Dy(), green)
	drawRect(screen, r.Min.X-10, r.Min.Y-10, r.Dx()+20, r.Dy()+20, black)
	vector.StrokeRect(screen, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), 2, white, false)
	drawTextCentered(screen, label, r, white)
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(black)

	// Draw the board.
	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			c := white
			if g.ghostGrid[row][col] {
				c = lgrey
			}
			if g.grid[row][col] {
				c = grey
			}
			drawRect(screen, (margin+cellSize)*col+margin, (margin+cellSize)*row+margin, cellSize, cellSize, c)
		}
	}

	// Draw the current block preview.
	drawText(screen, "Current:", windowW-400, windowH-400, white)
	for _, b := range g.currentBlock {
		row := g.currentPosition[0] + b[0]
		col := g.currentPosition[1] + b[1]
		drawRect(screen, (margin+cellSize)*col+margin+5, (margin+cellSize)*row+margin, cellSize, cellSize, white)
	}

	// Draw the next block preview.
	drawText(screen, "Next:", windowW-200, windowH-400, white)
	for _, b := range g.nextBlock {
		row := g.nextPosition[0] + b[0]
		col := g.nextPosition[1] + b[1]
		drawRect(screen, (margin+cellSize)*col+margin+5, (margin+cellSize)*row+margin, cellSize, cellSize, white)
	}

	// Draw the score.
	drawText(screen, fmt.Sprintf("Score: %d", g.score), windowW-400, windowH-120, white)

	// Draw the buttons.
	drawButton(screen, undoRect, "UNDO")
	drawButton(screen, resetRect, "RESET")
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return windowW, windowH
}

func main() {
	ebiten.SetWindowSize(windowW, windowH)
	ebiten.SetWindowTitle("Tactris")
	if err := ebiten.RunGame(newGame()); err != nil {
		log.Fatal(err)
	}
}
