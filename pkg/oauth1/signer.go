package oauth1

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/url"
)

// Signer is an interface to generate a signature.
//
// RFC 5849 - 3.4 Signature
type Signer interface {
	Method() string

	Sign(key string, message string) string
}

// HMACSHA1 is an implementation of Signer
//
// RFC 5849 - 3.4.2 HMAC-SHA1
type HMACSHA1 struct {
	ConsumerSecret string
}

func (s *HMACSHA1) Method() string {
	return "HMAC-SHA1"
}

func (s *HMACSHA1) Sign(tokenSecret string, text string) string {
	key := url.QueryEscape(s.ConsumerSecret) + "&" + url.QueryEscape(tokenSecret)
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(text))
	bytes := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(bytes)
}
