package system

// containsString 判断切片是否包含某值。
func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

// itoa 把整数转成字符串。
func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := []byte{}
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}

// firewalldService 把端口映射到 firewalld 的服务名。
//
// 端口不在已知列表时返回空串，调用方据此改用 --add-port。
func firewalldService(port string) string {
	switch port {
	case "80":
		return "http"
	case "443":
		return "https"
	case "22":
		return "ssh"
	case "25":
		return "smtp"
	case "53":
		return "dns"
	case "3306":
		return "mysql"
	case "5432":
		return "postgresql"
	case "6379":
		return "redis"
	default:
		return ""
	}
}
