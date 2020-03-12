package wsse

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

const iso8601 = "2006-01-02T15:04:05-0700"

// A Transport is an implementation of the http.RoundTripper for WSSE
// authentication.
//
// https://www.xml.com/pub/a/2003/12/17/dive.html
type Transport struct {
	// Username is a user name for the authentication
	Username string
	// Password is a password for the authentication
	Password string

	http.RoundTripper
}

func (t Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	var nonce [16]byte
	_, err := rand.Read(nonce[:])
	if err != nil {
		return nil, err
	}

	now := time.Now().Format(iso8601)

	token := bytes.Join([][]byte{nonce[:], []byte(now), []byte(t.Password)}, []byte{})
	digest := sha1.Sum(token)
	wsse := fmt.Sprintf("UsernameToken Username=%q, PasswordDigest=%q, Nonce=%q, Created=%q",
		t.Username,
		base64.StdEncoding.EncodeToString(digest[:]),
		base64.StdEncoding.EncodeToString(nonce[:]),
		now)
	r.Header.Set("X-WSSE", wsse)

	if t.RoundTripper == nil {
		return http.DefaultTransport.RoundTrip(r)
	}
	return t.RoundTripper.RoundTrip(r)
}
