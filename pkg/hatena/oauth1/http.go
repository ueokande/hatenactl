package oauth1

import (
	"net/http"
	"strconv"
	"time"
)

type Transport struct {
	Realm       string
	ConsumerKey string
	OAuthToken  Token
	Signer      Signer

	http.RoundTripper
}

func (t Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	oauthParams := map[string]string{
		"oauth_consumer_key":     t.ConsumerKey,
		"oauth_nonce":            nonce(),
		"oauth_timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_token":            t.OAuthToken.Token,
		"oauth_signature_method": t.Signer.Method(),
		"oauth_version":          "1.0",
	}
	oauthParams["oauth_signature"] = t.Signer.Sign(t.OAuthToken.Secret, signatureText(r, oauthParams, r.URL.Query()))
	oauthParams["realm"] = t.Realm

	r.Header.Set("Authorization", oauthAuthorizationHeaderValue(oauthParams))

	if t.RoundTripper == nil {
		return http.DefaultTransport.RoundTrip(r)
	}
	return t.RoundTripper.RoundTrip(r)
}

func NewHTTPClient(consumerKey, consumerSecret, token, tokenSecret string) *http.Client {
	return &http.Client{
		Transport: &Transport{
			ConsumerKey: consumerKey,
			Signer:      &HMACSHA1{ConsumerSecret: consumerSecret},
			OAuthToken: Token{
				Token:  token,
				Secret: tokenSecret,
			},
		},
	}
}
