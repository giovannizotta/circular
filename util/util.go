package util

func All(v []bool) bool {
	for _, b := range v {
		if !b {
			return false
		}
	}
	return true
}

func GetDirection(from, to string) string {
	if from < to {
		return "1"
	}
	return "0"
}
