package world

// Room represents a rectangular room in the dungeon.
type Room struct {
	X, Y          int // Top-left corner position
	Width, Height int // Dimensions of the room
}

// Center returns the center coordinates of the room.
func (r Room) Center() (int, int) {
	return r.X + r.Width/2, r.Y + r.Height/2
}

// Contains returns true if the given point is inside the room.
func (r Room) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.Width && y >= r.Y && y < r.Y+r.Height
}

// Intersects returns true if this room overlaps with another room.
func (r Room) Intersects(other Room) bool {
	return r.X < other.X+other.Width &&
		r.X+r.Width > other.X &&
		r.Y < other.Y+other.Height &&
		r.Y+r.Height > other.Y
}
