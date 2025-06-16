// Package game provides utilities for game mechanics
package game

import (
	"math/rand"
	"sync"
	"time"
)

// Random provides a thread-safe random number generator for game mechanics
// This uses math/rand which is suitable for game features but NOT for security
type Random struct {
	rng *rand.Rand
	mu  sync.Mutex
}

// NewRandom creates a new game random number generator
func NewRandom() *Random {
	return &Random{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NewSeededRandom creates a new game random number generator with a specific seed
// This is useful for reproducible game mechanics (e.g., testing, replays)
func NewSeededRandom(seed int64) *Random {
	return &Random{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// Intn returns a random int in [0,n)
func (r *Random) Intn(n int) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.Intn(n)
}

// Int63n returns a random int64 in [0,n)
func (r *Random) Int63n(n int64) int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.Int63n(n)
}

// Float64 returns a random float64 in [0.0,1.0)
func (r *Random) Float64() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.Float64()
}

// RollDice simulates rolling a die with the given number of sides
func (r *Random) RollDice(sides int) int {
	if sides <= 0 {
		return 0
	}
	return r.Intn(sides) + 1
}

// RollMultipleDice simulates rolling multiple dice
func (r *Random) RollMultipleDice(count, sides int) []int {
	results := make([]int, count)
	for i := 0; i < count; i++ {
		results[i] = r.RollDice(sides)
	}
	return results
}

// Choose randomly selects one item from a slice
func (r *Random) Choose(n int) int {
	if n <= 0 {
		return 0
	}
	return r.Intn(n)
}

// Shuffle randomly shuffles a slice of integers
func (r *Random) Shuffle(slice []int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rng.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}

// DefaultRandom is a global instance for convenience
var DefaultRandom = NewRandom()

// Package-level convenience functions using the default random

// RollDice rolls a die with the given number of sides
func RollDice(sides int) int {
	return DefaultRandom.RollDice(sides)
}

// Choose randomly selects an index from 0 to n-1
func Choose(n int) int {
	return DefaultRandom.Choose(n)
}