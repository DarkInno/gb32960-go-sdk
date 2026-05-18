package gb32960

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

var (
	ErrKeyTooShort = errors.New("gb32960: AES key too short")
	ErrDecrypt     = errors.New("gb32960: AES decrypt failed")
)

func DeriveAESKey(token []byte) []byte {
	key := make([]byte, aes.BlockSize)
	n := copy(key, token)
	for i := n; i < aes.BlockSize; i++ {
		key[i] = 0
	}
	return key
}

func EncryptAES128(data, key []byte) ([]byte, error) {
	if len(key) < aes.BlockSize {
		return nil, ErrKeyTooShort
	}

	block, err := aes.NewCipher(key[:aes.BlockSize])
	if err != nil {
		return nil, err
	}

	padded := pkcs7Pad(data, aes.BlockSize)
	encrypted := make([]byte, len(padded))

	iv := key[:aes.BlockSize]
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted, padded)

	return encrypted, nil
}

func DecryptAES128(data, key []byte) ([]byte, error) {
	if len(key) < aes.BlockSize {
		return nil, ErrKeyTooShort
	}
	if len(data)%aes.BlockSize != 0 {
		return nil, ErrDecrypt
	}

	block, err := aes.NewCipher(key[:aes.BlockSize])
	if err != nil {
		return nil, err
	}

	decrypted := make([]byte, len(data))
	iv := key[:aes.BlockSize]
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decrypted, data)

	unpadded, err := pkcs7Unpad(decrypted)
	if err != nil {
		return nil, ErrDecrypt
	}

	return unpadded, nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padLen := blockSize - len(data)%blockSize
	pad := bytes.Repeat([]byte{byte(padLen)}, padLen)
	return append(data, pad...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}
	padLen := int(data[len(data)-1])
	if padLen > len(data) || padLen == 0 || padLen > aes.BlockSize {
		return nil, errors.New("invalid padding")
	}
	return data[:len(data)-padLen], nil
}
