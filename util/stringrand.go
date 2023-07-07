package util

import (
	"math/rand"
	"time"
)

// 随机字符串生成器
type StringRand interface {
	RandString(lens int) string
}

type stringRand struct {
	base string
	r    *rand.Rand
}

func (sr *stringRand) RandString(lens int) string {
	b := make([]byte, lens)
	for i := 0; i < len(b); i++ {
		b[i] = sr.base[sr.r.Intn(len(sr.base))]
	}
	return string(b)
}

func NewStringRand() StringRand {
	var (
		Numbers      = "0123456789"
		LettersUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		LettersLower = "abcdefghijklmnopqrstuvwxyz"
	)
	base := Numbers + LettersUpper + LettersLower
	sr := &stringRand{
		r:    rand.New(rand.NewSource(time.Now().UnixNano())),
		base: base,
	}
	return sr
}
