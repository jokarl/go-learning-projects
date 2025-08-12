package math

import "math"

// NextPow2 calculates the next power of 2
// that is greater than or equal to c.
func NextPow2(c int) int {
	if c <= 1 {
		return 0
	}
	// In mathematics, the binary logarithm is the power to
	// which the number 2 must be raised to obtain the value c.
	// E.g. trying to divide into 8 subnets requires 3 bits (2^3 = 8).
	// E.g. trying to divide into 9 subnets requires 4 bits because:
	// 2^3 = 8 (3 borrowed bits) can fit at most 8 subnets, so we need to borrow one more bit:
	// 2^4 = 16 (4 borrowed bits) can fit 9 subnets, but also 16 subnets.
	return int(math.Ceil(math.Log2(float64(c))))
}
