package main

import (
	"math/rand"
	"strconv"
	"strings"
)

var randomCharTable = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
var randomCharTableSize = len(randomCharTable)

func randomChar() byte {
	return randomCharTable[rand.Intn(randomCharTableSize)]
}

// encode a password.
//
// This is still plain text and every 7th char is the password
// in reverse. The rest is just random characters.
func encode(in string) string {
	var out strings.Builder

	j := len(in)

	for i := 1; i <= (320 - j); i++ {
		switch {
		case 0 == i%7 && j > 0:
			j--
			out.WriteByte(in[j])
		case i == 123:
			out.WriteString(strconv.Itoa(len(in) / 10))
		case i == 289:
			out.WriteString(strconv.Itoa(len(in) % 10))
		default:
			out.WriteByte(randomChar())
		}
	}

	return out.String()
}
