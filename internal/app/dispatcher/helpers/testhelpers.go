package helpers

import (
	"math/rand"
	"time"
)

var lettersLower = []rune("abcdefghijklmnopqrstuvwxyz")

// RandomName random letter sequence
func RandomName(n int) string {
	return randFromSelection(n, lettersLower)
}

func randFromSelection(length int, choices []rune) string {
	b := make([]rune, length)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = choices[rand.Intn(len(choices))]
	}
	return string(b)
}
