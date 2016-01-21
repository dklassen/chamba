package encrypt

// encrypt string to base64 crypto using AES
import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
)

func AESEncrypt(keyText, text string) (encryptedText string, err error) {
	key := []byte(keyText)
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Panicln(err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		log.Println(err)
		return
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	encryptedText = base64.URLEncoding.EncodeToString(ciphertext)
	return
}

func AESDecrypt(keystring, cryptotext string) (decryptedText string, err error) {
	key := []byte(keystring)
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptotext)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println(err)
		return
	}

	if len(ciphertext) < aes.BlockSize {
		log.Println("ciphertext too short")
		return
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	decryptedText = string(ciphertext)

	return
}
