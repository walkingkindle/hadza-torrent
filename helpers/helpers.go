// Package helpers are just some helper functions
package helpers

func CountTrue(bools []bool) int {
	count := 0

	for _, b := range bools {
		if b {
			count++
		}
	}

	return count
}
