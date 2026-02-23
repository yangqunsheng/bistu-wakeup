package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"math/rand"
)

// CAS 加密使用的字符集（与 encrypt.js 中 $aes_chars 一致）
const aesChars = "ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678"

// randomString 生成指定长度的随机字符串
func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = aesChars[rand.Intn(len(aesChars))]
	}
	return string(b)
}

// EncryptPassword 使用 AES-CBC-128 加密密码（与 CAS encrypt.js 完全对齐）
// 流程：randomString(64) + password → AES-CBC(key=salt, iv=randomString(16)) → Base64(密文)
func EncryptPassword(password, salt string) (string, error) {
	key := []byte(salt)
	if len(key) != 16 {
		return "", fmt.Errorf("salt 长度必须为 16 字节，当前: %d", len(key))
	}

	// 64 字符随机前缀 + 密码
	plaintext := []byte(randomString(64) + password)

	// PKCS7 填充
	blockSize := aes.BlockSize
	padding := blockSize - len(plaintext)%blockSize
	for i := 0; i < padding; i++ {
		plaintext = append(plaintext, byte(padding))
	}

	// 16 字符随机 IV（ASCII 字符，与 JS 端一致）
	iv := []byte(randomString(16))

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建 AES cipher 失败: %w", err)
	}

	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)

	// 仅返回 Base64(密文)，不包含 IV
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
