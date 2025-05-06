package net

// IsValidPort 检查端口是否合法。 0 被视为无效端口.
func IsValidPort(port int) bool {
	return port > 0 && port < 65535
}
