package cryptosvc

import "runtime"

// Zero 清零一段缓冲区。
//
// Go 编译器**可能**消除「看似无副作用」的清零循环，
// 因此必须配合 runtime.KeepAlive 保证不被优化掉
// （见 docs/standards/03-crypto.md §10）。
//
// b 为 nil 时是空操作。
func Zero(b []byte) {
	if len(b) == 0 {
		return
	}
	for i := range b {
		b[i] = 0
	}
	runtime.KeepAlive(b)
}
