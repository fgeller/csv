package main

import "math/rand"
import "time"
import "strings"

type parameters struct {
	fields        int
	maxWordLength int
	minWordLength int
	lineCount     int
}

func randomASCIIByte() byte {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	return byte(rand.Intn(95) + 32)
}

func randomWord(minLength int, maxLength int) string {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	length := rand.Intn(maxLength-minLength) + minLength

	characters := make([]byte, length)
	for index := 0; index < length; index += 1 {
		characters[index] = randomASCIIByte()
	}

	return string(characters)
}

func randomLine(parameters *parameters) string {
	randomWords := make([]string, parameters.fields)
	for index := 0; index < parameters.fields; index += 1 {
		randomWords[index] = randomWord(parameters.minWordLength, parameters.maxWordLength)
	}
	return strings.Join(randomWords, ",")
}
