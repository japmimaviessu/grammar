package grammar

import (
	"math/rand"
	"time"
)

var rnd *rand.Rand

func init() {
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func random(low int, high int) int {
	return low + rnd.Intn(high - low + 1)
}

func next(i *int) int {
	*i += 1
	return *i
}
