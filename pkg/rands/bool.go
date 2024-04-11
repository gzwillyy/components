package rands

// Bool 生成真假随机值
func Bool() bool {
	return Int(0, 1) == 0
}
