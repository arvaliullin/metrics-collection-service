package crypto

import (
	"crypto/rand"
	"crypto/rsa"
)

// Encrypt шифрует данные публичным RSA-ключом.
func Encrypt(publicKey *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, publicKey, plaintext)
}

// Decrypt расшифровывает данные приватным RSA-ключом.
func Decrypt(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, privateKey, ciphertext)
}
