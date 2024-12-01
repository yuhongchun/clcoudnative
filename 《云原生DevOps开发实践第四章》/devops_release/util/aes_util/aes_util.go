package aesutil

import (
	"crypto/aes"
	"crypto/cipher"

	"devops_release/config"
)

var commonIV = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
var c *cipher.Block

func EncryptString(text string) ([]byte, error) {

	key := config.ApplicationConfig.AesKey
	textBytes := []byte(text)
	if c == nil {
		c1, err := aes.NewCipher([]byte(key))
		if err != nil {
			return nil, err
		}
		c = &c1
	}
	//加密字符串
	cfb := cipher.NewCFBEncrypter(*c, commonIV)
	ciphertext := make([]byte, len(text))
	cfb.XORKeyStream(ciphertext, textBytes)
	return ciphertext, nil
}
func DecryptString(textBytes []byte) (string, error) {
	key := config.ApplicationConfig.AesKey
	if c == nil {
		c1, err := aes.NewCipher([]byte(key))
		if err != nil {
			return "", err
		}
		c = &c1
	}
	//解密字符串
	cfbdec := cipher.NewCFBDecrypter(*c, commonIV)
	plaintextCopy := make([]byte, len(textBytes))
	ciphertext := textBytes
	cfbdec.XORKeyStream(plaintextCopy, ciphertext)
	return string(plaintextCopy), nil
}
