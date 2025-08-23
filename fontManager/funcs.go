package fontManager

func makeSubsetRange(end int) map[int]int {
	answer := make(map[int]int)
	for i := 0; i < end; i++ {
		answer[i] = 0
	}
	return answer
}
