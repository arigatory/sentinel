// Package hash реализует подпись передаваемых данных по алгоритму HMAC-SHA256.
package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Header — имя HTTP-заголовка, в котором передаётся подпись тела запроса/ответа.
const Header = "HashSHA256"

// Sign возвращает HMAC-SHA256 от data с ключом key в hex-представлении.
func Sign(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// Valid сообщает, соответствует ли полученная подпись got расчётной для data.
// Сравнение выполняется за постоянное время.
func Valid(data []byte, key, got string) bool {
	expected, err := hex.DecodeString(got)
	if err != nil {
		return false
	}

	actual, err := hex.DecodeString(Sign(data, key))
	if err != nil {
		return false
	}

	return hmac.Equal(actual, expected)
}
