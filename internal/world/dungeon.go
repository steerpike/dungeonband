package world

import (
	"context"
	"math/rand"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/samdwyer/dungeonband/internal/telemetry"
)

const (
	// Default dungeon dimensions
	DefaultWidth  = 80
	DefaultHeight = 24

	// BSP parameters
	minRoomSize = 8  // Minimum room dimension (accommodates party formation)
	maxRoomSize = 15 // Maximum room dimension
	minLeafSize = 10 // Minimum BSP leaf size before stopping split
)

// Dungeon represents the game map.
type Dungeon struct {
	Width  int
	Height int
	Tiles  [][]Tile
	Rooms  []Room
	rng    *rand.Rand
}

// NewDungeon creates a new dungeon filled with walls.
func NewDungeon(width, height int) *Dungeon {
	tiles := make([][]Tile, height)
	for y := range tiles {
		tiles[y] = make([]Tile, width)
		for x := range tiles[y] {
			tiles[y][x] = TileWall
		}
	}

	return &Dungeon{
		Width:  width,
		Height: height,
		Tiles:  tiles,
		Rooms:  make([]Room, 0),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Generate creates the dungeon layout using BSP algorithm.
func (d *Dungeon) Generate(ctx context.Context) {
	tracer := telemetry.Tracer("world")
	ctx, span := tracer.Start(ctx, "dungeon.generate")
	defer span.End()

	startTime := time.Now()

	// Start BSP with the entire dungeon as root
	root := &bspNode{
		x:      1,
		y:      1,
		width:  d.Width - 2,
		height: d.Height - 2,
	}

	// Recursively split the dungeon
	d.splitNode(root)

	// Create rooms in leaf nodes
	d.createRooms(root)

	// Connect rooms with corridors
	d.connectRooms(root)

	// Record telemetry
	span.SetAttributes(
		attribute.Int("dungeon.width", d.Width),
		attribute.Int("dungeon.height", d.Height),
		attribute.Int("dungeon.room_count", len(d.Rooms)),
		attribute.Int64("dungeon.generation_ms", time.Since(startTime).Milliseconds()),
	)
}

// IsPassable returns true if the given position can be walked on.
func (d *Dungeon) IsPassable(x, y int) bool {
	if x < 0 || x >= d.Width || y < 0 || y >= d.Height {
		return false
	}
	return d.Tiles[y][x].IsPassable()
}

// GetTile returns the tile at the given position.
func (d *Dungeon) GetTile(x, y int) Tile {
	if x < 0 || x >= d.Width || y < 0 || y >= d.Height {
		return TileWall
	}
	return d.Tiles[y][x]
}

// RoomIndexAt returns the index of the room containing the position, or -1 if not in a room.
func (d *Dungeon) RoomIndexAt(x, y int) int {
	for i, room := range d.Rooms {
		if room.Contains(x, y) {
			return i
		}
	}
	return -1
}

// RandomPointInRoom returns a random passable point within the specified room.
func (d *Dungeon) RandomPointInRoom(roomIndex int) (int, int) {
	if roomIndex < 0 || roomIndex >= len(d.Rooms) {
		return -1, -1
	}
	room := d.Rooms[roomIndex]

	// Try random points until we find a passable one (max 100 attempts)
	for i := 0; i < 100; i++ {
		x := room.X + d.rng.Intn(room.Width)
		y := room.Y + d.rng.Intn(room.Height)
		if d.IsPassable(x, y) {
			return x, y
		}
	}

	// Fallback to room center
	return room.Center()
}

// bspNode represents a node in the BSP tree.
type bspNode struct {
	x, y          int
	width, height int
	left, right   *bspNode
	room          *Room
}

// isLeaf returns true if this node has no children.
func (n *bspNode) isLeaf() bool {
	return n.left == nil && n.right == nil
}

// splitNode recursively splits a BSP node.
func (d *Dungeon) splitNode(node *bspNode) {
	// Stop if too small to split
	if node.width < minLeafSize*2 && node.height < minLeafSize*2 {
		return
	}

	// Determine split direction
	var splitHorizontally bool
	if node.width > node.height && node.width >= minLeafSize*2 {
		splitHorizontally = false // Split vertically (left/right)
	} else if node.height >= minLeafSize*2 {
		splitHorizontally = true // Split horizontally (top/bottom)
	} else if node.width >= minLeafSize*2 {
		splitHorizontally = false
	} else {
		return // Can't split
	}

	// Calculate split position (between 45% and 55% for variety)
	var splitPos int
	if splitHorizontally {
		min := minLeafSize
		max := node.height - minLeafSize
		if max <= min {
			return
		}
		splitPos = min + d.rng.Intn(max-min+1)
	} else {
		min := minLeafSize
		max := node.width - minLeafSize
		if max <= min {
			return
		}
		splitPos = min + d.rng.Intn(max-min+1)
	}

	// Create child nodes
	if splitHorizontally {
		node.left = &bspNode{
			x:      node.x,
			y:      node.y,
			width:  node.width,
			height: splitPos,
		}
		node.right = &bspNode{
			x:      node.x,
			y:      node.y + splitPos,
			width:  node.width,
			height: node.height - splitPos,
		}
	} else {
		node.left = &bspNode{
			x:      node.x,
			y:      node.y,
			width:  splitPos,
			height: node.height,
		}
		node.right = &bspNode{
			x:      node.x + splitPos,
			y:      node.y,
			width:  node.width - splitPos,
			height: node.height,
		}
	}

	// Recursively split children
	d.splitNode(node.left)
	d.splitNode(node.right)
}

// createRooms creates rooms in leaf nodes of the BSP tree.
func (d *Dungeon) createRooms(node *bspNode) {
	if node == nil {
		return
	}

	if node.isLeaf() {
		// Create a room within this leaf
		roomWidth := minRoomSize + d.rng.Intn(min(maxRoomSize-minRoomSize+1, node.width-minRoomSize+1))
		roomHeight := minRoomSize + d.rng.Intn(min(maxRoomSize-minRoomSize+1, node.height-minRoomSize+1))

		// Ensure room fits within leaf
		if roomWidth > node.width-2 {
			roomWidth = node.width - 2
		}
		if roomHeight > node.height-2 {
			roomHeight = node.height - 2
		}
		if roomWidth < minRoomSize || roomHeight < minRoomSize {
			return // Skip if too small
		}

		// Random position within leaf
		roomX := node.x + 1 + d.rng.Intn(node.width-roomWidth-1)
		roomY := node.y + 1 + d.rng.Intn(node.height-roomHeight-1)

		room := Room{
			X:      roomX,
			Y:      roomY,
			Width:  roomWidth,
			Height: roomHeight,
		}
		node.room = &room
		d.Rooms = append(d.Rooms, room)

		// Carve out the room
		d.carveRoom(room)
	} else {
		d.createRooms(node.left)
		d.createRooms(node.right)
	}
}

// carveRoom sets all tiles within the room to floor.
func (d *Dungeon) carveRoom(room Room) {
	for y := room.Y; y < room.Y+room.Height; y++ {
		for x := room.X; x < room.X+room.Width; x++ {
			if x > 0 && x < d.Width-1 && y > 0 && y < d.Height-1 {
				d.Tiles[y][x] = TileFloor
			}
		}
	}
}

// connectRooms connects rooms with corridors.
func (d *Dungeon) connectRooms(node *bspNode) {
	if node == nil || node.isLeaf() {
		return
	}

	// Connect children first
	d.connectRooms(node.left)
	d.connectRooms(node.right)

	// Get a room from each subtree and connect them
	leftRoom := d.getRoom(node.left)
	rightRoom := d.getRoom(node.right)

	if leftRoom != nil && rightRoom != nil {
		d.carveCorridor(*leftRoom, *rightRoom)
	}
}

// getRoom returns a room from a subtree (any room will do).
func (d *Dungeon) getRoom(node *bspNode) *Room {
	if node == nil {
		return nil
	}

	if node.room != nil {
		return node.room
	}

	// Try left subtree first
	if room := d.getRoom(node.left); room != nil {
		return room
	}
	return d.getRoom(node.right)
}

// carveCorridor creates a corridor between two rooms.
func (d *Dungeon) carveCorridor(room1, room2 Room) {
	x1, y1 := room1.Center()
	x2, y2 := room2.Center()

	// Randomly choose to go horizontal-then-vertical or vertical-then-horizontal
	if d.rng.Intn(2) == 0 {
		d.carveHorizontalTunnel(x1, x2, y1)
		d.carveVerticalTunnel(y1, y2, x2)
	} else {
		d.carveVerticalTunnel(y1, y2, x1)
		d.carveHorizontalTunnel(x1, x2, y2)
	}
}

// carveHorizontalTunnel carves a horizontal tunnel.
func (d *Dungeon) carveHorizontalTunnel(x1, x2, y int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		if x > 0 && x < d.Width-1 && y > 0 && y < d.Height-1 {
			d.Tiles[y][x] = TileFloor
		}
	}
}

// carveVerticalTunnel carves a vertical tunnel.
func (d *Dungeon) carveVerticalTunnel(y1, y2, x int) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		if x > 0 && x < d.Width-1 && y > 0 && y < d.Height-1 {
			d.Tiles[y][x] = TileFloor
		}
	}
}
