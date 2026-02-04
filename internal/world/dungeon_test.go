package world

import (
	"context"
	"math/rand"
	"testing"
)

func TestDungeonReproducibility(t *testing.T) {
	// Generate two dungeons with the same seed
	seed := int64(12345)

	rng1 := rand.New(rand.NewSource(seed))
	rng2 := rand.New(rand.NewSource(seed))

	d1 := NewDungeon(DefaultWidth, DefaultHeight, rng1)
	d2 := NewDungeon(DefaultWidth, DefaultHeight, rng2)

	ctx := context.Background()
	d1.Generate(ctx)
	d2.Generate(ctx)

	// Verify same number of rooms
	if len(d1.Rooms) != len(d2.Rooms) {
		t.Fatalf("Room count mismatch: %d != %d", len(d1.Rooms), len(d2.Rooms))
	}

	// Verify rooms are in same positions
	for i := range d1.Rooms {
		r1, r2 := d1.Rooms[i], d2.Rooms[i]
		if r1.X != r2.X || r1.Y != r2.Y || r1.Width != r2.Width || r1.Height != r2.Height {
			t.Errorf("Room %d mismatch: (%d,%d,%d,%d) != (%d,%d,%d,%d)",
				i, r1.X, r1.Y, r1.Width, r1.Height,
				r2.X, r2.Y, r2.Width, r2.Height)
		}
	}

	// Verify tiles are identical
	for y := 0; y < d1.Height; y++ {
		for x := 0; x < d1.Width; x++ {
			if d1.Tiles[y][x] != d2.Tiles[y][x] {
				t.Errorf("Tile mismatch at (%d,%d): %v != %v", x, y, d1.Tiles[y][x], d2.Tiles[y][x])
			}
		}
	}
}

func TestDungeonDifferentSeeds(t *testing.T) {
	// Generate two dungeons with different seeds - they should be different
	rng1 := rand.New(rand.NewSource(12345))
	rng2 := rand.New(rand.NewSource(54321))

	d1 := NewDungeon(DefaultWidth, DefaultHeight, rng1)
	d2 := NewDungeon(DefaultWidth, DefaultHeight, rng2)

	ctx := context.Background()
	d1.Generate(ctx)
	d2.Generate(ctx)

	// With different seeds, at least room positions should differ
	// (very unlikely to be identical by chance)
	identical := true
	for i := range d1.Rooms {
		if i >= len(d2.Rooms) {
			identical = false
			break
		}
		r1, r2 := d1.Rooms[i], d2.Rooms[i]
		if r1.X != r2.X || r1.Y != r2.Y {
			identical = false
			break
		}
	}

	if len(d1.Rooms) != len(d2.Rooms) {
		identical = false
	}

	if identical {
		t.Error("Dungeons with different seeds should not be identical")
	}
}
