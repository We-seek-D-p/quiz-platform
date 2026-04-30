package redis

import (
    "math/rand"
    "strconv"
    "time"
)

type RandomRoomCodeGenerator struct {
    rnd *rand.Rand
}

func NewRandomRoomCodeGenerator() *RandomRoomCodeGenerator {
    return &RandomRoomCodeGenerator{
        rnd: rand.New(rand.NewSource(time.Now().UnixNano())),
    }
}

func (g *RandomRoomCodeGenerator) Generate() string {
    code := g.rnd.Intn(99_999_999) + 1
    return strconv.Itoa(code)
}
