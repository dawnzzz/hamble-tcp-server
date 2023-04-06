package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"
)

// GenerateCrtAndKeyFile 生成证书和私钥文件
func GenerateCrtAndKeyFile(crtFileName, KeyFileName string) (err error) {
	defer func() {
		if err != nil {
			// 如果期间发生错误，删除以及生成的证书和私钥文件
			_ = os.Remove(crtFileName)
			_ = os.Remove(KeyFileName)
		}
	}()
	// 生成私钥
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	// 创建证书模板
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Beijing University of Post and Telecommunication"},
		},

		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(24 * time.Hour * 365 * 10), // 证书十年之内有效

		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// 创建证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	// 序列化证书文件
	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if pemCert == nil {
		return err
	}
	if err := os.WriteFile(crtFileName, pemCert, 0644); err != nil {
		return err
	}

	// 生成私钥文件
	privateBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return err
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateBytes})
	if pemKey == nil {
		return err
	}
	if err := os.WriteFile(KeyFileName, pemKey, 0600); err != nil {
		return err
	}

	return nil
}
