// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package utils

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"io"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

const (
	//AlphaUpper all upper alphas chars
	AlphaUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	//AlphaLower all lowers alphas chars
	AlphaLower = "abcdefghijklmnopqrstuvwxyz"
	//AlphaUpperLower all upper and lowers aplhas chars
	AlphaUpperLower = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	//AlphaUpperLowerNumeric all upper lowers alphas and numerics
	AlphaUpperLowerNumeric = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyz"
	//AlphaUpperLowerNumericSpecial all upper lowers alphas, numerics and special chas
	AlphaUpperLowerNumericSpecial = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz123456890" +
		"!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
)

//RandomSource express which chars should be considered
type RandomSource struct {
	CharsToGen string
}

//RandomSequence returns random character with given length;
//random source express which chars should be considered
func (r RandomSource) RandomSequence(length int) []byte {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		char, _ := rand.Int(rand.Reader, big.NewInt(int64(len(r.CharsToGen))))
		result[i] = r.CharsToGen[int(char.Int64())]
	}
	return result
}

//RandomSecureKey returns random character with given length
func RandomSecureKey(length int) []byte {
	k := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil
	}
	return k
}

//CryptPassword bcrypts a password at given costs
func CryptPassword(password []byte, cost int) ([]byte, error) {
	s, err := bcrypt.GenerateFromPassword(password, cost)

	if err != nil {
		return nil, err
	}

	return s, nil
}

//GenerateSalt generates a random salt with alphanumerics and some special characters
func GenerateSalt() []byte {
	r := RandomSource{
		CharsToGen: AlphaUpperLowerNumericSpecial,
	}
	return r.RandomSequence(32)
}

//EncodeBase64 encodes a string to base64
func EncodeBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

//DecodeBase64 descodes a string to base64
func DecodeBase64(b64 string) (string, error) {
	out, err := base64.StdEncoding.DecodeString(b64)
	return string(out), err
}

func RandomHash(length int) string {
	hash := sha512.New()
	hash.Write(RandomSecureKey(length))

	return hex.EncodeToString(hash.Sum(nil))
}
