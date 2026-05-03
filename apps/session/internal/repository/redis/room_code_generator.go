package redis

import (
	"fmt"
	"math"
	"math/rand/v2"
)

const defaultCodeLength = 8

type RandomRoomCodeGenerator struct {
	length int
	limit  uint64
}

func NewRandomRoomCodeGenerator() *RandomRoomCodeGenerator {
	return &RandomRoomCodeGenerator{
		length: defaultCodeLength,
		limit:  uint64(math.Pow10(defaultCodeLength)),
	}
}

func (g *RandomRoomCodeGenerator) Generate() string {
	val := rand.Uint64N(g.limit) //nolint:gosec
	return fmt.Sprintf("%0*d", g.length, val)
}
