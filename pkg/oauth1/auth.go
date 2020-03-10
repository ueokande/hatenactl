package oauth1

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// CallbackOOB means the client is unable to receive callbacks (out-of-band).
const CallbackOOB = "oob"

// Token represents an oauth token with its secret
type Token struct {
	Token  string
	Secret string
}

// Client is a OAuth 1.0 Client
type Client struct {
	Realm       string
	ConsumerKey string
	Signer      Signer

	TemporaryCredentialURI        string
	ResourceOwnerAuthorizationURI string
	TokenRequestURI               string
}

// Initiate gets a temporary credential from a service provider, and returns
// the token.
//
// The callbackURL is a URL to which the server will redirect the
// resource onwer when Resource Owner Authorization step is completed.
// The params is a service optional parameters submitted via request body in
// application/x-www-form-urlencoded.
//
// See: RFC 5849 - 2.1 Temporary Credentials
func (c *Client) Initiate(ctx context.Context, callbackURL string, params url.Values) (Token, error) {
	body := bytes.NewBufferString(params.Encode())
	req, err := http.NewRequest(http.MethodPost, c.TemporaryCredentialURI, body)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return Token{}, err
	}

	oauthParams := map[string]string{
		"oauth_consumer_key":     c.ConsumerKey,
		"oauth_nonce":            nonce(),
		"oauth_timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_signature_method": c.Signer.Method(),
		"oauth_version":          "1.0",
		"oauth_callback":         callbackURL,
	}
	oauthParams["oauth_signature"] = c.Signer.Sign("", c.signatureText(req, oauthParams, params))
	oauthParams["realm"] = c.Realm

	req.Header.Set("Authorization", oauthAuthorizationHeaderValue(oauthParams))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Token{}, err
	}
	if resp.StatusCode != 200 {
		return Token{}, fmt.Errorf("server returns %d (%s): %s", resp.StatusCode, resp.Status, respBody)
	}
	vals, err := url.ParseQuery(string(respBody))
	if err != nil {
		return Token{}, err
	}

	var token Token
	if v := vals.Get("oauth_token"); len(v) == 0 {
		return Token{}, errors.New(`response has no "oauth_token" field`)
	} else {
		token.Token = v
	}
	if v := vals.Get("oauth_token_secret"); len(v) == 0 {
		return Token{}, errors.New(`response has no "oauth_token_secret" field`)
	} else {
		token.Secret = v
	}
	return token, nil
}

// GetAuthorizeURL returns an URL to authorize an OAuth app.
//
// The token is returned value by Initiate() method.
//
// See: RFC 5849 2.2 Resource Owner Authorization.
func (c *Client) GetAuthorizeURL(ctx context.Context, token string) (string, error) {
	u, err := url.Parse(c.ResourceOwnerAuthorizationURI)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Add("oauth_token", token)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// GetAccessToken get a set of token credentials from the server.

// The token parameter is the returned value by Initiate() method, and the
// verifier is presented by Resource Owner Authorization step.
//
// See: RFC 5849 - 2.3 Token Credentials
func (c *Client) GetAccessToken(ctx context.Context, token Token, verifier string) (Token, error) {
	req, err := http.NewRequest(http.MethodPost, c.TokenRequestURI, nil)
	if err != nil {
		return Token{}, err
	}

	oauthParams := map[string]string{
		"oauth_consumer_key":     c.ConsumerKey,
		"oauth_nonce":            nonce(),
		"oauth_timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_token":            token.Token,
		"oauth_verifier":         verifier,
		"oauth_signature_method": c.Signer.Method(),
		"oauth_version":          "1.0",
	}
	oauthParams["oauth_signature"] = c.Signer.Sign(token.Secret, c.signatureText(req, oauthParams, nil))
	oauthParams["realm"] = c.Realm

	req.Header.Set("Authorization", oauthAuthorizationHeaderValue(oauthParams))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Token{}, err
	}
	if resp.StatusCode != 200 {
		return Token{}, fmt.Errorf("server returns %d (%s): %s", resp.StatusCode, resp.Status, respBody)
	}

	vals, err := url.ParseQuery(string(respBody))
	if err != nil {
		return Token{}, err
	}

	if v := vals.Get("oauth_token"); len(v) == 0 {
		return Token{}, errors.New(`response has no "oauth_token" field`)
	} else {
		token.Token = v
	}
	if v := vals.Get("oauth_token_secret"); len(v) == 0 {
		return Token{}, errors.New(`response has no "oauth_token_secret" field`)
	} else {
		token.Secret = v
	}
	return token, nil
}

// signatureText returns a base string of the signing by Signer.
//
// The oauthParams is a set of the OAuth parameter (such as // oauth_consumer_key).
// The caller must exclude "realm" parameter.  The userParams is a set of the
// optional parameter specified in request body or URL query.
//
// See: RFC 5849 - 3.4.1  Signature Base String
func (c *Client) signatureText(req *http.Request, oauthParams map[string]string, userParams url.Values) string {
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

// nonce generate a random bytes with length of 16 bytes with HEX-encoded.
func nonce() string {
	var nonce [16]byte
	_, err := rand.Read(nonce[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", nonce)
}

// oauthAuthorizationHeaderValue presents a value of "Authorization" header
// from the params.
func oauthAuthorizationHeaderValue(params map[string]string) string {
	var kvs []string
	for k, v := range params {
		kvs = append(kvs, url.QueryEscape(k)+"=\""+url.QueryEscape(v)+"\"")
	}
	return "OAuth " + strings.Join(kvs, ", ")
}
