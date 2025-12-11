package util

func StatusToValue(status string, statuses []string) float64 {
	for idx, s := range statuses {
		if status == s {
			return float64(idx)
		}
	}

	return -1
}
