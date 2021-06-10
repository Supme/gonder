package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

type CryptoAlg string

const (
	CryptoAlgAES  = CryptoAlg("aes_")
	CryptoAlgNone = CryptoAlg("")
)

// Encrypt encrypt bytes to base64 data
func Encrypt(data []byte) (string, error) {
	if Config.secretString != "" {
		return encrypt(CryptoAlgAES, data)
	}
	return encrypt(CryptoAlgNone, data)
}

func encrypt(alg CryptoAlg, data []byte) (string, error) {
	switch alg {
	case CryptoAlgAES:
		encrypted, err := encryptAES(hashTo32Bytes(Config.secretString), data)
		if err != nil {
			return "", err
		}
		return string(CryptoAlgAES) + base64.URLEncoding.EncodeToString(encrypted), nil

	default:
		return base64.URLEncoding.EncodeToString(data), nil
	}
}

// Decrypt decrypt base64 data to bytes
func Decrypt(data string) ([]byte, error) {
	switch {
	case strings.HasPrefix(data, string(CryptoAlgAES)):
		encrypted, err := base64.URLEncoding.DecodeString(strings.TrimPrefix(data, string(CryptoAlgAES)))
		if err != nil {
			return nil, err
		}
		return decryptAES(hashTo32Bytes(Config.secretString), encrypted)

	default:
		return base64.URLEncoding.DecodeString(data)
	}
}

func encryptAES(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	output := make([]byte, aes.BlockSize+len(data))
	iv := output[:aes.BlockSize]
	encrypted := output[aes.BlockSize:]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encrypted, data)
	return output, nil
}

func decryptAES(key, cryptoData []byte) ([]byte, error) {
	if len(cryptoData) < aes.BlockSize {
		return nil, fmt.Errorf("cipherText too short. It decodes to %v bytes but the minimum length is %d", len(cryptoData), aes.BlockSize)
	}
	iv := cryptoData[:aes.BlockSize]
	cryptoData = cryptoData[aes.BlockSize:]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cryptoData, cryptoData)
	return cryptoData, nil
}

// hashTo32Bytes will compute a cryptographically useful hash of the input string.
func hashTo32Bytes(input string) []byte {
	data := sha256.Sum256([]byte(input))
	return data[0:]
}
