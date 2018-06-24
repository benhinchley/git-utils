package diff

// Strings returns the unique values from a,b
func Strings(a, b []string) []string {
	var diff []string

	for i := 0; i < 2; i++ {
		for _, v1 := range a {
			found := false

			for _, v2 := range b {
				if v1 == v2 {
					found = true
					break
				}
			}

			// string not found append to return slice
			if !found {
				diff = append(diff, v1)
			}
		}

		// swap the slices after the first interation
		if i == 0 {
			a, b = b, a
		}
	}

	return diff
}
