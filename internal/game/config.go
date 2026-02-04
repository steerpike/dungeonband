package game

// Config holds game configuration options.
type Config struct {
	// Seed for random number generation. Used for reproducible dungeon generation.
	// A seed of 0 means a random seed will be generated.
	Seed int64
}
