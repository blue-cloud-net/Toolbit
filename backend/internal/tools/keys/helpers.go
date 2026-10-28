package keys

// containsString 判断字符串切片是否包含目标值。
func containsString(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

// containsInt 判断整数切片是否包含目标值。
func containsInt(list []int, target int) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}
