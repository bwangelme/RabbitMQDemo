package utils

import (
	"math/rand"
	"strings"
)

func RandomString(l int) string {
	var buf strings.Builder
	for i := 0; i < l; i++ {
		buf.WriteByte(byte(ranInt(65, 90)))
	}
	return buf.String()
}

func ranInt(min, max int) int {
	return min + rand.Intn(max-min)
}
