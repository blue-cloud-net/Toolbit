package pkisvc

import (
	"github.com/blue-cloud-net/tongsuo-go/x509"
)

// Store 是证书信任库句柄。
//
// 它持有原生句柄，**必须**显式 Close（幂等）。
//
// ⚠️ 每次验证**必须**新建 Store：`SetTime` 会影响该 store 后续的**所有**
// 验证（见 docs/standards/03-crypto.md §8.3）。
type Store struct {
	// store 是底层信任库；Close 之后置为 nil。
	store *x509.Store
}

// NewStore 创建空的信任库。
//
// 与 v0.5.0 的链验证保持一致：**不**做任何隐式系统信任根加载
// （上游未提供该入口，见 I-17）。
func NewStore() *Store {
	return &Store{store: x509.NewStore()}
}

// Close 释放信任库句柄。
//
// **幂等**：重复调用只返回首次的结果。
func (s *Store) Close() error {
	if s == nil || s.store == nil {
		return nil
	}
	err := s.store.Close()
	s.store = nil
	return err
}

// AddCertificate 把一张证书加入信任库。
func (s *Store) AddCertificate(cert *Certificate) error {
	if s == nil || s.store == nil {
		return toolerrInvalidCert("trust store already closed")
	}
	if cert == nil || cert.cert == nil {
		return toolerrInvalidCert("certificate unavailable")
	}
	if err := s.store.AddCert(cert.cert); err != nil {
		return toolerrInvalidCert("cannot add certificate to trust store")
	}
	return nil
}

// addAll 把多张证书加入信任库。
func (s *Store) addAll(certs []*Certificate) error {
	for _, cert := range certs {
		if err := s.AddCertificate(cert); err != nil {
			return err
		}
	}
	return nil
}
