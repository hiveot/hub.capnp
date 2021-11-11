package idprovclient

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

// Create an HMAC signature from a JSON message with the out of band secret
// This generates the HMAC of the message with SHA256 hash of the secret
//  message is the JSON message
//  secret is the secret shared with the receiver
// This returns the base64 encoded HMAC
func Sign(message string, secret string) (base64Encoded string, err error) {
	hmac256 := hmac.New(sha256.New, []byte(secret))
	hmac256.Write([]byte(message))
	hmacMessage := hmac256.Sum(nil)
	base64Signature := base64.StdEncoding.EncodeToString(hmacMessage)
	return base64Signature, nil
}

// Verify a JSON message signature with the out of band secret
//  message is the JSON message
//  secret is the secret shared with the receiver
//  signature is the message signature Base64(HMAC(message,SHA256(secret)))
// This returns an error if verification failed
func Verify(message string, secret string, base64Signature string) error {
	hmacMessage, err := base64.StdEncoding.DecodeString(base64Signature)
	if err != nil {
		return err
	}
	// generate a new HMAC using our secret
	hmac256 := hmac.New(sha256.New, []byte(secret))
	hmac256.Write([]byte(message))
	myHmac := hmac256.Sum(nil)

	equal := hmac.Equal(hmacMessage, myHmac)
	if !equal {
		return errors.New("Signature.Verify: message verification failed. Shared secrets do not match")
	}
	return nil
}
