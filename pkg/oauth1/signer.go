package oauth1

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"net/url"
	"sort"
	"strings"
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

// signatureText returns a base string of the signing by Signer.
//
// The oauthParams is a set of the OAuth parameter (such as // oauth_consumer_key).
// The caller must exclude "realm" parameter.  The userParams is a set of the
// optional parameter specified in request body or URL query.
//
// See: RFC 5849 - 3.4.1  Signature Base String
func signatureText(req *http.Request, oauthParams map[string]string, userParams url.Values) string {
	params := make(map[string][]string)
	for k, v := range oauthParams {
		params[url.QueryEscape(k)] = append(params[k], v)
	}
	for k, vs := range userParams {
		params[url.QueryEscape(k)] = append(params[k], vs...)
	}

	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var kvs []string
	for _, k := range keys {
		for _, v := range params[k] {
			kvs = append(kvs, k+"="+url.QueryEscape(v))
		}
	}
	urlPart := url.QueryEscape(req.URL.Scheme + "://" + req.URL.Host + req.URL.Path)
	parameterPart := url.QueryEscape(strings.Join(kvs, "&"))
	return strings.Join([]string{req.Method, urlPart, parameterPart}, "&")
}
