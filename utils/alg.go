package utils

func Fib(n int) int {
	if n <= 0 {
		return 0
	}
	if n < 3 {
		return 1
	}

	var (
		a   = 1
		b   = 1
		res = 2
	)
	for i := 4; i <= n; i++ {
		a = b
		b = res
		res = a + b
	}

	return res
}
