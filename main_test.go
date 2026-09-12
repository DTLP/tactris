package main

import "testing"

func setTarget(g *Game, i int, fig [][2]int) {
	g.targets[i] = target{figure: copyBlock(fig), refIndex: -1}
}

func drawCell(g *Game, r, c int) {
	g.grid[r][c] = stateActive
	g.activeCells = append(g.activeCells, [2]int{r, c})
	g.matchCheck()
}

func TestMatchCommit(t *testing.T) {
	g := newGame()
	setTarget(g, 0, [][2]int{{0, 0}, {0, 1}, {0, 2}, {1, 1}}) // T down
	// Draw the T one cell at a time, as if clicking.
	cells := [][2]int{{2, 3}, {2, 4}, {2, 5}, {3, 4}}
	g.drawing = true
	for _, c := range cells {
		drawCell(g, c[0], c[1])
	}
	if g.score != 0 {
		t.Fatalf("score = %d, want 0 before release", g.score)
	}
	if len(g.activeCells) != 4 {
		t.Fatalf("activeCells = %d, want 4", len(g.activeCells))
	}
	g.drawing = false
	g.matchCheck()
	if g.score != 4 {
		t.Fatalf("score = %d, want 4 after commit", g.score)
	}
	for _, c := range cells {
		if g.grid[c[0]][c[1]] != statePlaced {
			t.Fatalf("cell %v should be placed", c)
		}
	}
	if len(g.activeCells) != 0 {
		t.Fatal("expected activeCells cleared after commit")
	}
}

func TestAutoErase(t *testing.T) {
	g := newGame()
	setTarget(g, 0, [][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}}) // O
	g.drawing = true
	drawCell(g, 0, 0)
	drawCell(g, 0, 1)
	drawCell(g, 5, 5) // stray cell: cannot belong to any tetromino subset
	if len(g.activeCells) != 2 {
		t.Fatalf("activeCells = %d, want 2 after erase", len(g.activeCells))
	}
	if g.grid[0][0] != stateEmpty {
		t.Fatal("oldest active cell should be erased")
	}
}

func TestUndo(t *testing.T) {
	g := newGame()
	setTarget(g, 0, [][2]int{{0, 0}, {0, 1}, {0, 2}, {1, 1}}) // T down
	targetsBefore := copyTargets(g.targets)
	cells := [][2]int{{2, 3}, {2, 4}, {2, 5}, {3, 4}}
	g.drawing = true
	for _, c := range cells {
		drawCell(g, c[0], c[1])
	}
	g.drawing = false
	g.matchCheck()
	if g.score != 4 {
		t.Fatalf("score = %d, want 4", g.score)
	}
	g.undoStep()
	if g.score != 0 {
		t.Fatalf("score after undo = %d, want 0", g.score)
	}
	for _, c := range cells {
		if g.grid[c[0]][c[1]] != stateEmpty {
			t.Fatalf("cell %v should be empty after undo", c)
		}
	}
	for i := range g.targets {
		if g.targets[i].refIndex != targetsBefore[i].refIndex {
			t.Fatal("targets should be restored after undo")
		}
	}
}

func TestClearFullRow(t *testing.T) {
	g := newGame()
	for c := 0; c < boardSize; c++ {
		g.grid[7][c] = statePlaced
	}
	g.grid[9][0] = statePlaced // marker
	if n := g.clearLines(); n != 1 {
		t.Fatalf("clearLines() = %d, want 1", n)
	}
	if g.score != 10 {
		t.Fatalf("score = %d, want 10", g.score)
	}
	// The cleared row (bottom half) lands at the bottom; the marker shifts up.
	if g.grid[8][0] != statePlaced {
		t.Fatal("marker should shift up to row 8")
	}
	if g.grid[9][0] != stateEmpty {
		t.Fatal("cleared row should land at the bottom")
	}
}

func TestClearTopRow(t *testing.T) {
	g := newGame()
	for c := 0; c < boardSize; c++ {
		g.grid[2][c] = statePlaced
	}
	g.grid[0][0] = statePlaced // marker
	if n := g.clearLines(); n != 1 {
		t.Fatalf("clearLines() = %d, want 1", n)
	}
	// The cleared row (top half) lands at the top; the marker shifts down.
	if g.grid[1][0] != statePlaced {
		t.Fatal("marker should shift down to row 1")
	}
	if g.grid[0][0] != stateEmpty {
		t.Fatal("cleared row should land at the top")
	}
}

func TestClearFullColumn(t *testing.T) {
	g := newGame()
	for r := 0; r < boardSize; r++ {
		g.grid[r][2] = statePlaced
	}
	g.grid[0][0] = statePlaced // marker
	if n := g.clearLines(); n != 1 {
		t.Fatalf("clearLines() = %d, want 1", n)
	}
	if g.score != 10 {
		t.Fatalf("score = %d, want 10", g.score)
	}
	// The cleared column (left half) lands at the left; the marker shifts right.
	if g.grid[0][1] != statePlaced {
		t.Fatal("marker should shift right to col 1")
	}
	if g.grid[0][0] != stateEmpty {
		t.Fatal("column 0 should be the cleared column")
	}
}

func TestClearRightColumn(t *testing.T) {
	g := newGame()
	for r := 0; r < boardSize; r++ {
		g.grid[r][7] = statePlaced
	}
	g.grid[0][9] = statePlaced // marker
	if n := g.clearLines(); n != 1 {
		t.Fatalf("clearLines() = %d, want 1", n)
	}
	// The cleared column (right half) lands at the right; the marker shifts left.
	if g.grid[0][8] != statePlaced {
		t.Fatal("marker should shift left to col 8")
	}
	if g.grid[0][9] != stateEmpty {
		t.Fatal("column 9 should be the cleared column")
	}
}

func TestMultiLineScore(t *testing.T) {
	g := newGame()
	for c := 0; c < boardSize; c++ {
		g.grid[2][c] = statePlaced
		g.grid[5][c] = statePlaced
	}
	if n := g.clearLines(); n != 2 {
		t.Fatalf("clearLines() = %d, want 2", n)
	}
	if g.score != 40 {
		t.Fatalf("score = %d, want 40 (10 * 2^2)", g.score)
	}
}

func TestGameOver(t *testing.T) {
	g := newGame()
	for r := 0; r < boardSize; r++ {
		for c := 0; c < boardSize; c++ {
			g.grid[r][c] = statePlaced
		}
	}
	setTarget(g, 0, [][2]int{{0, 0}, {1, 0}, {2, 0}, {3, 0}})
	setTarget(g, 1, [][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}})
	if !g.checkGameOver() {
		t.Fatal("expected game over on a full board")
	}
	if !g.gameOver {
		t.Fatal("expected gameOver flag to be set")
	}
}

func TestNoGameOverWithRoom(t *testing.T) {
	g := newGame()
	setTarget(g, 0, [][2]int{{0, 0}, {1, 0}, {2, 0}, {3, 0}})
	setTarget(g, 1, [][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}})
	if g.checkGameOver() {
		t.Fatal("expected no game over on an empty board")
	}
	if g.gameOver {
		t.Fatal("expected gameOver flag to be false")
	}
}
