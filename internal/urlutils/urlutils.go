package urlutils

import (
	"errors"
	randMath "math/rand/v2"
	"regexp"
)

var urlStorage = make(map[string]string)

func DecodeURL(token string) (string, error) {
	if url, found := urlStorage[token]; found {
		return url, nil
	} else {
		return "", errors.New("haven't found the URL")
	}
}

func randRange(a int, b int) int {
	return randMath.IntN(b-a) + a
}

func generateToken(length int) string {
	var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	token := make([]byte, length)
	for i := range token {
		token[i] = letters[randMath.IntN(len(letters))]
	}
	return string(token)
}

// Encode URL to a range number from 6 to 9 of random characters
func EncodeURL(url string) string {
	// Creating the random length (from 6 to including 9) slice of bytes
	length := randRange(6, 10)
	token := generateToken(length)
	urlStorage[string(token)] = url
	return string(token)
}

// Check the validity of provided URL address
func URLCheck(url string) bool {
	var regExpURLPattern = regexp.MustCompile(`^((http|https):\/\/)?([a-zA-Z0-9.-]+(\.[a-zA-Z]{2,})+)(\/[a-zA-Z0-9-._~:?#@!$&'()*+,;=]*)?$`)
	return regExpURLPattern.MatchString(url)
}
