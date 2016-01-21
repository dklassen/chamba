package encrypt

import (
	"strings"
	"testing"
)

func TestAESEncryptDecrypt(t *testing.T) {
	key := "a very very very very secret key"
	expected := "the pig is in the punch"

	cryptoText, _ := AESEncrypt(key, expected)
	result, _ := AESDecrypt(key, cryptoText)
	if strings.Compare(result, expected) != 0 {
		t.Error("expected:", string(expected), "got:", string(result))
	}
}
