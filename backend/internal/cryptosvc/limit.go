package cryptosvc

// heavyConcurrency 是重操作的并发上限。
//
// 见 docs/standards/03-crypto.md §12：密钥生成、证书签发、PKCS#12 打包等
// 重操作必须经信号量限并发（≤ 2）。
//
// 这是**常量**而非环境变量：并发上限属于实现约束，不随部署环境变化。
const heavyConcurrency = 2

// semaphores 汇集一组信号量。
type semaphores struct {
	// heavy 是重操作许可池。
	heavy chan struct{}
}

// defaultSemaphores 是进程级的默认信号量集合。
var defaultSemaphores = &semaphores{
	heavy: make(chan struct{}, heavyConcurrency),
}
