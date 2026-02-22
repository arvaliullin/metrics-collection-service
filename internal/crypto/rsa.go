package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

// Encrypt шифрует данные публичным RSA-ключом.
func Encrypt(publicKey *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, plaintext, nil)
}

// Decrypt расшифровывает данные приватным RSA-ключом.
func Decrypt(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, nil)
}
