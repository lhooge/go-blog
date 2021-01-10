// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//package crypt
package crypt

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

const bcryptRounds = 12

var (
	//AlphaUpper all upper alphas chars
	AlphaUpper = RandomSource("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	//AlphaLower all lowers alphas chars
	AlphaLower = RandomSource("abcdefghijklmnopqrstuvwxyz")
	//AlphaUpperLower all upper and lowers aplhas chars
	AlphaUpperLower = RandomSource("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	//AlphaUpperLowerNumeric all upper lowers alphas and numerics
	AlphaUpperLowerNumeric = RandomSource("ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyz")
	//AlphaUpperLowerNumericSpecial all upper lowers alphas, numerics and special chars
	AlphaUpperLowerNumericSpecial = RandomSource("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz123456890" +
		"!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~")
)

//RandomSource string containing which characters should be considered when generating random sequences
type RandomSource string

//RandomSequence returns random character with given length;
func (r RandomSource) RandomSequence(length int) []byte {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		char, _ := rand.Int(rand.Reader, big.NewInt(int64(len(r))))
		result[i] = r[int(char.Int64())]
	}
	fmt.Println(result)
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

//CryptPassword hashes a password with bcrypt and a given cost
func CryptPassword(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, bcryptRounds)
}

//GenerateSalt generates a random salt with alphanumerics and some special characters
func GenerateSalt() []byte {
	return AlphaUpperLowerNumericSpecial.RandomSequence(32)
}

func RandomHash(length int) string {
	hash := sha512.New()
	hash.Write(RandomSecureKey(length))

	return hex.EncodeToString(hash.Sum(nil))
}
