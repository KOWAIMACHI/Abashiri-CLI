package helper

// Removes duplicate elements from a single array and returns an array containing only unique elements.
func RemoveDuplicatesFromArray(elements []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for _, element := range elements {
		if !encountered[element] {
			encountered[element] = true
			result = append(result, element)
		}
	}
	return result
}

// Removes duplicate elements from two arrays, a and b, and returns an array with elements remaining only from a.
func RemoveDuplicatesBetweenArrays(a []string, b []string) []string {
	bMap := make(map[string]struct{})
	for _, val := range b {
		bMap[val] = struct{}{}
	}

	var result []string
	for _, val := range a {
		if _, found := bMap[val]; !found {
			result = append(result, val)
		}
	}
	return result
}
