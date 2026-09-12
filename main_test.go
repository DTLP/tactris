package main

import "testing"

func TestClearFullRow(t *testing.T) {
	g := newGame()
	for c := 0; c < boardSize; c++ {
		g.grid[5][c] = true
	}
	g.grid[0][0] = true
	g.score = 0
	if got := g.clearLines(); got != 1 {
		t.Fatalf("clearLines() = %d, want 1", got)
	}
	if g.score != 10 {
		t.Fatalf("score = %d, want 10", g.score)
	}
	// The full row is removed and kept rows shift down by one.
	if g.grid[0][0] {
		t.Fatal("full row should have been cleared")
	}
	if !g.grid[1][0] {
		t.Fatal("marker cell should have shifted down one row")
	}
}

func TestClearFullColumn(t *testing.T) {
	g := newGame()
	for r := 0; r < boardSize; r++ {
		g.grid[r][2] = true
	}
	g.grid[0][0] = true
	g.score = 0
	if got := g.clearLines(); got != 1 {
		t.Fatalf("clearLines() = %d, want 1", got)
	}
	if g.score != 10 {
		t.Fatalf("score = %d, want 10", g.score)
	}
	// The full column is removed and kept columns shift right by one.
	if g.grid[0][2] {
		t.Fatal("full column should have been cleared")
	}
	if !g.grid[0][1] {
		t.Fatal("marker cell should have shifted right one column")
	}
}

func TestPlaceAndUndo(t *testing.T) {
	g := newGame()
	block := [][2]int{{0, 0}}
	if !g.canPlaceBlock(block, [2]int{0, 0}) {
		t.Fatal("expected block to be placeable on empty board")
	}
	g.addBlock(block, [2]int{0, 0})
	if !g.grid[0][0] {
		t.Fatal("expected cell to be filled after addBlock")
	}
	g.undoStep()
	if g.grid[0][0] {
		t.Fatal("expected cell to be empty after undo")
	}
}

func TestCannotPlaceOutOfBounds(t *testing.T) {
	g := newGame()
	block := [][2]int{{0, 0}}
	if g.canPlaceBlock(block, [2]int{10, 0}) {
		t.Fatal("expected block off-board to be unplaceable")
	}
	// A block that extends past the board edge must be rejected too.
	tall := [][2]int{{0, 0}, {1, 0}}
	if g.canPlaceBlock(tall, [2]int{9, 9}) {
		t.Fatal("expected block extending past the board edge to be unplaceable")
	}
}
