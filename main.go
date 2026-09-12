package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	cellSize    = 40
	margin      = 1
	boardSize   = 10
	panelWidth  = 230
	previewCell = 20
	previewStep = 21

	highScoreFile = "highscore.txt"
)

var (
	cellStep = cellSize + margin
	boardPx  = cellStep*boardSize + margin
	designW  = boardPx + panelWidth
	designH  = boardPx
)

var (
	black     = color.RGBA{0, 0, 0, 255}
	white     = color.RGBA{255, 255, 255, 255}
	darkGrey  = color.RGBA{25, 25, 30, 255}
	ghostBlue = color.RGBA{0xAA, 0xDD, 0xFF, 255}
)

var (
	panelX     = boardPx
	scoreX     = panelX + 12
	shape1X    = panelX + 12
	shape2X    = panelX + 105
	scoreY     = 11
	highScoreY = 40
	shapeY     = 95
	undoRect   = image.Rect(panelX+12, 300, panelX+12+96, 300+36)
	resetRect  = image.Rect(panelX+12, 346, panelX+12+96, 346+36)
)

type cellState int

const (
	stateEmpty cellState = iota
	stateActive
	statePlaced
)

type target struct {
	figure   [][2]int
	refIndex int
}

// The 19 one-sided tetrominoes, normalized so the minimum row and column are 0.
var blocks = [][][2]int{
	{{0, 0}, {1, 0}, {2, 0}, {3, 0}}, // I vertical
	{{0, 0}, {0, 1}, {0, 2}, {0, 3}}, // I horizontal
	{{0, 0}, {0, 1}, {1, 0}, {1, 1}}, // O
	{{0, 1}, {1, 0}, {1, 1}, {1, 2}}, // T up
	{{0, 1}, {1, 1}, {1, 2}, {2, 1}}, // T right
	{{0, 0}, {0, 1}, {0, 2}, {1, 1}}, // T down
	{{0, 1}, {1, 0}, {1, 1}, {2, 1}}, // T left
	{{0, 1}, {0, 2}, {1, 0}, {1, 1}}, // S horizontal
	{{0, 0}, {1, 0}, {1, 1}, {2, 1}}, // S vertical
	{{0, 0}, {0, 1}, {1, 1}, {1, 2}}, // Z horizontal
	{{0, 1}, {1, 0}, {1, 1}, {2, 0}}, // Z vertical
	{{0, 0}, {1, 0}, {1, 1}, {1, 2}}, // J 0
	{{0, 0}, {0, 1}, {1, 0}, {2, 0}}, // J 90
	{{0, 0}, {0, 1}, {0, 2}, {1, 2}}, // J 180
	{{0, 1}, {1, 1}, {2, 0}, {2, 1}}, // J 270
	{{0, 2}, {1, 0}, {1, 1}, {1, 2}}, // L 0
	{{0, 0}, {1, 0}, {2, 0}, {2, 1}}, // L 90
	{{0, 0}, {0, 1}, {0, 2}, {1, 0}}, // L 180
	{{0, 0}, {0, 1}, {1, 1}, {2, 1}}, // L 270
}

func generateTarget(existingRefs []int) target {
	for {
		idx := rand.IntN(len(blocks))
		if !containsInt(existingRefs, idx) {
			return target{figure: copyBlock(blocks[idx]), refIndex: idx}
		}
	}
}

func containsInt(s []int, v int) bool {
	for _, e := range s {
		if e == v {
			return true
		}
	}
	return false
}

func copyBlock(block [][2]int) [][2]int {
	out := make([][2]int, len(block))
	copy(out, block)
	return out
}

func emptyGrid() [][boardSize]cellState {
	return make([][boardSize]cellState, boardSize)
}

func copyGrid(src [][boardSize]cellState) [][boardSize]cellState {
	dst := make([][boardSize]cellState, boardSize)
	copy(dst, src)
	return dst
}

func copyTargets(t [2]target) [2]target {
	var out [2]target
	for i := range t {
		out[i] = target{figure: copyBlock(t[i].figure), refIndex: t[i].refIndex}
	}
	return out
}

type Game struct {
	grid        [][boardSize]cellState
	prevGrid    [][boardSize]cellState
	targets     [2]target
	prevTargets [2]target
	activeCells [][2]int
	drawing     bool
	score       int
	prevScore   int
	highScore   int
	gameOver    bool

	canvas  *ebiten.Image
	scale   float64
	offsetX float64
	offsetY float64
	lastW   int
	lastH   int
	stable  int
}

func newGame() *Game {
	g := &Game{lastW: designW, lastH: designH}
	g.highScore = loadHighScore()
	g.resetGame()
	return g
}

func (g *Game) resetGame() {
	g.recordHighScore()
	g.grid = emptyGrid()
	g.prevGrid = emptyGrid()
	g.targets[0] = generateTarget(nil)
	g.targets[1] = generateTarget([]int{g.targets[0].refIndex})
	g.prevTargets = copyTargets(g.targets)
	g.activeCells = nil
	g.drawing = false
	g.score = 0
	g.prevScore = 0
	g.gameOver = false
}

// recordHighScore keeps the running high score up to date and persists it.
func (g *Game) recordHighScore() {
	if g.score > g.highScore {
		g.highScore = g.score
	}
	if err := os.WriteFile(highScoreFile, []byte(strconv.Itoa(g.highScore)), 0o644); err != nil {
		log.Printf("could not save high score: %v", err)
	}
}

func loadHighScore() int {
	data, err := os.ReadFile(highScoreFile)
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func (g *Game) undoStep() {
	g.grid = copyGrid(g.prevGrid)
	g.score = g.prevScore
	g.targets = copyTargets(g.prevTargets)
	g.activeCells = nil
	g.drawing = false
	g.gameOver = false
}

// matchCheck inspects the drawn cells. If they form a translated subset of a
// target shape they are kept (and committed once a full 4-cell match is drawn
// and the mouse is released). Otherwise the oldest drawn cell is erased.
func (g *Game) matchCheck() {
	n := len(g.activeCells)
	if n < 2 {
		return
	}
	minRow, minCol := boardSize, boardSize
	for _, c := range g.activeCells {
		if c[0] < minRow {
			minRow = c[0]
		}
		if c[1] < minCol {
			minCol = c[1]
		}
	}
	offsets := [][2]int{
		{minRow, minCol},
		{minRow, minCol - 1},
		{minRow - 1, minCol},
		{minRow, minCol - 2},
		{minRow - 2, minCol},
	}
	for _, off := range offsets {
		for fi := range g.targets {
			if subsetMatch(g.activeCells, g.targets[fi].figure, off) {
				if !g.drawing && n == 4 {
					g.commit(fi)
				}
				return
			}
		}
	}
	oldest := g.activeCells[0]
	g.activeCells = g.activeCells[1:]
	g.grid[oldest[0]][oldest[1]] = stateEmpty
}

func subsetMatch(active, figure [][2]int, off [2]int) bool {
	count := 0
	for _, a := range active {
		for _, fc := range figure {
			if a[0]-off[0] == fc[0] && a[1]-off[1] == fc[1] {
				count++
				break
			}
		}
	}
	return count == len(active)
}

func (g *Game) commit(matched int) {
	g.prevGrid = copyGrid(g.grid)
	for _, c := range g.activeCells {
		g.prevGrid[c[0]][c[1]] = stateEmpty
	}
	g.prevScore = g.score
	g.prevTargets = copyTargets(g.targets)

	g.score += 4
	for _, c := range g.activeCells {
		g.grid[c[0]][c[1]] = statePlaced
	}
	g.activeCells = nil
	g.drawing = false

	g.clearLines()
	g.targets[matched] = generateTarget([]int{g.targets[1-matched].refIndex})
	if g.checkGameOver() {
		g.recordHighScore()
	}
}

// clearLines empties every full row and column, moving each emptied line to
// the nearest edge. Returns the number of lines cleared.
func (g *Game) clearLines() int {
	var rows, cols []int
	for t := 0; t < boardSize; t++ {
		full := true
		for c := 0; c < boardSize; c++ {
			if g.grid[t][c] != statePlaced {
				full = false
				break
			}
		}
		if full {
			rows = append(rows, t)
		}
		full = true
		for r := 0; r < boardSize; r++ {
			if g.grid[r][t] != statePlaced {
				full = false
				break
			}
		}
		if full {
			cols = append(cols, t)
		}
	}
	n := len(rows) + len(cols)

	for _, r := range rows {
		for c := 0; c < boardSize; c++ {
			g.grid[r][c] = stateEmpty
		}
		row := g.grid[r]
		g.grid = append(g.grid[:r], g.grid[r+1:]...)
		if r >= boardSize/2 {
			g.grid = append(g.grid, row)
			for i := range rows {
				if rows[i] > r {
					rows[i]--
				}
			}
		} else {
			g.grid = append([][boardSize]cellState{row}, g.grid...)
			for i := range rows {
				if rows[i] != r {
					rows[i]++
				}
			}
		}
	}

	for _, ci := range cols {
		for r := 0; r < boardSize; r++ {
			g.grid[r][ci] = stateEmpty
		}
		var col [boardSize]cellState
		for r := 0; r < boardSize; r++ {
			col[r] = g.grid[r][ci]
		}
		for r := 0; r < boardSize; r++ {
			var newRow [boardSize]cellState
			if ci >= boardSize/2 {
				copy(newRow[:], g.grid[r][:ci])
				copy(newRow[ci:], g.grid[r][ci+1:])
				newRow[boardSize-1] = col[r]
			} else {
				newRow[0] = col[r]
				copy(newRow[1:1+ci], g.grid[r][:ci])
				copy(newRow[1+ci:], g.grid[r][ci+1:])
			}
			g.grid[r] = newRow
		}
		if ci >= boardSize/2 {
			for i := range cols {
				if cols[i] > ci {
					cols[i]--
				}
			}
		} else {
			for i := range cols {
				if cols[i] != ci {
					cols[i]++
				}
			}
		}
	}

	g.score += 10 * n * n
	return n
}

// checkGameOver reports whether neither target shape fits anywhere on the board.
func (g *Game) checkGameOver() bool {
	for _, tgt := range g.targets {
		for row := 0; row < boardSize; row++ {
			for col := 0; col < boardSize; col++ {
				fits := true
				for _, cell := range tgt.figure {
					r := row + cell[0]
					c := col + cell[1]
					if r >= boardSize || c >= boardSize || g.grid[r][c] == statePlaced {
						fits = false
						break
					}
				}
				if fits {
					return false
				}
			}
		}
	}
	g.gameOver = true
	return true
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
	limit := float64(cellStep * boardSize)
	if x < 0 || y < 0 || x >= limit || y >= limit {
		return 0, 0, false
	}
	return int(y) / cellStep, int(x) / cellStep, true
}

// enforceAspectRatio keeps the window close to the design aspect ratio. The
// window resizes freely while being dragged; once it stays the same size for a
// few frames it snaps to the design ratio to avoid letterboxing.
func (g *Game) enforceAspectRatio() {
	w, h := ebiten.WindowSize()
	if w <= 0 || h <= 0 {
		return
	}
	if w != g.lastW || h != g.lastH {
		g.lastW, g.lastH = w, h
		g.stable = 0
		return
	}
	g.stable++
	if g.stable < 15 {
		return
	}
	scale := math.Min(float64(w)/float64(designW), float64(h)/float64(designH))
	wantW := int(math.Round(float64(designW) * scale))
	wantH := int(math.Round(float64(designH) * scale))
	if wantW < 1 || wantH < 1 {
		return
	}
	if wantW != w || wantH != h {
		ebiten.SetWindowSize(wantW, wantH)
	}
}

func (g *Game) Update() error {
	g.enforceAspectRatio()

	x, y := g.sceneMouse()
	pos := image.Pt(int(x), int(y))
	justPressed := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	released := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)

	if g.gameOver {
		if justPressed {
			g.resetGame()
		}
		return nil
	}

	if justPressed {
		switch {
		case pos.In(resetRect):
			g.resetGame()
			return nil
		case pos.In(undoRect):
			g.undoStep()
			return nil
		}
	}

	pressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if justPressed || (pressed && g.drawing) {
		if row, col, ok := cellAt(x, y); ok {
			if g.grid[row][col] == stateEmpty {
				g.grid[row][col] = stateActive
				g.activeCells = append(g.activeCells, [2]int{row, col})
				g.drawing = true
				g.matchCheck()
			}
		}
	}
	if released && g.drawing {
		g.drawing = false
		g.matchCheck()
	}
	return nil
}

func drawRect(screen *ebiten.Image, x, y, w, h int, c color.Color) {
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), c, false)
}

func drawText(screen *ebiten.Image, s string, x, y int, scale float64, c color.Color) {
	face := basicfont.Face7x13
	w := font.MeasureString(face, s).Ceil()
	img := ebiten.NewImage(w, 14)
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  fixed.P(0, 12),
	}
	d.DrawString(s)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)
}

func drawTextCentered(screen *ebiten.Image, s string, r image.Rectangle, c color.Color) {
	face := basicfont.Face7x13
	w := font.MeasureString(face, s).Ceil()
	x := r.Min.X + (r.Dx()-w*2)/2
	y := r.Min.Y + (r.Dy()-13*2)/2
	drawText(screen, s, x, y, 2, c)
}

func drawButton(screen *ebiten.Image, r image.Rectangle, label string) {
	drawRect(screen, r.Min.X, r.Min.Y, r.Dx(), r.Dy(), darkGrey)
	vector.StrokeRect(screen, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), 2, white, false)
	drawTextCentered(screen, label, r, white)
}

func drawShape(screen *ebiten.Image, fig [][2]int, ox, oy int, c color.Color) {
	for _, cell := range fig {
		drawRect(screen, ox+previewStep*cell[1], oy+previewStep*cell[0], previewCell, previewCell, c)
	}
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
			c := black
			switch g.grid[row][col] {
			case stateActive:
				c = ghostBlue
			case statePlaced:
				c = white
			}
			drawRect(canvas, cellStep*col+margin, cellStep*row+margin, cellSize, cellSize, c)
		}
	}

	// Draw the right-hand panel.
	drawRect(canvas, boardPx, 0, designW-boardPx, designH, darkGrey)

	// Draw the score and high score.
	drawText(canvas, fmt.Sprintf("Score: %d", g.score), scoreX, scoreY, 1.5, white)
	drawText(canvas, fmt.Sprintf("High score: %d", g.highScore), scoreX, highScoreY, 1.5, white)

	// Draw the two target shapes.
	drawShape(canvas, g.targets[0].figure, shape1X, shapeY, ghostBlue)
	drawShape(canvas, g.targets[1].figure, shape2X, shapeY, ghostBlue)

	// Draw the buttons.
	drawButton(canvas, undoRect, "UNDO")
	drawButton(canvas, resetRect, "RESET")

	// Game over overlay.
	if g.gameOver {
		drawRect(canvas, 0, 0, boardPx, boardPx, color.RGBA{0, 0, 0, 200})
		drawTextCentered(canvas, "No moves left", image.Rect(0, 170, boardPx, 200), white)
		drawTextCentered(canvas, "Click to restart", image.Rect(0, 205, boardPx, 235), white)
	}

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
	g := newGame()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
	g.recordHighScore()
}
