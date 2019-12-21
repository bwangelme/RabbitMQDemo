package utils_test

import (
	"testing"

	"github.com/bwangelme/RabbitMQDemo/utils"
)

func TestFib(t *testing.T) {
	var fibTests = []struct {
		n, res int
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{3, 2},
		{12, 144},
		{13, 233},
	}

	for _, tt := range fibTests {
		res := utils.Fib(tt.n)
		if tt.res != res {
			t.Errorf(`Fib(%d) = %d`, tt.n, res)
		}
	}
}
