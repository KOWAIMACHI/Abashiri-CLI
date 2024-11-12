package helper

func RemoveDuplicates(elements []string) []string {
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
