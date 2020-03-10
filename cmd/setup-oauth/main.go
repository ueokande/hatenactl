package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/ueokande/hatenactl/pkg/oauth1"
)

var (
	OAuthConsumerKey    = os.Getenv("OAUTH_CONSUMER_KEY")
	OAuthConsumerSecret = os.Getenv("OAUTH_CONSUMER_SECRET")
)

func openURL(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	}
	return errors.New("unknown platform: " + runtime.GOOS)
}

func run(ctx context.Context) error {
	if len(OAuthConsumerKey) == 0 {
		return errors.New("OAUTH_CONSUMER_KEY not set")
	}
	if len(OAuthConsumerSecret) == 0 {
		return errors.New("OAUTH_CONSUMER_SECRET not set")
	}

	client := oauth1.Client{
		Signer:      &oauth1.HMACSHA1{ConsumerSecret: OAuthConsumerSecret},
		ConsumerKey: OAuthConsumerKey,

		TemporaryCredentialURI:        "https://www.hatena.com/oauth/initiate",
		ResourceOwnerAuthorizationURI: "https://www.hatena.ne.jp/oauth/authorize",
		TokenRequestURI:               "https://www.hatena.com/oauth/token",
	}

	q := url.Values{}
	q.Set("scope", "read_public")
	token, err := client.Initiate(ctx, oauth1.CallbackOOB, q)
	if err != nil {
		return fmt.Errorf("unable to initiate oauth app: %w", err)
	}

	authzURL, err := client.GetAuthorizeURL(ctx, token.Token)
	if err != nil {
		return fmt.Errorf("unable to authorize oauth app: %w", err)
	}
	err = openURL(authzURL)
	if err != nil {
		return fmt.Errorf("unable to open URL: %w", err)
	}
	fmt.Print("Enter verification code: ")

	r := bufio.NewReader(os.Stdin)
	verifier, err := r.ReadString('\n')
	if err != nil {
		return err
	}

	token, err = client.GetAccessToken(ctx, token, strings.TrimSpace(verifier))
	if err != nil {
		return fmt.Errorf("unable to get access toke: %w", err)
	}
	fmt.Println("Verification succeeded")
	fmt.Println("oauth_token: " + token.Token)
	fmt.Println("oauth_token_secret: " + token.Secret)
	return nil
}

func main() {
	err := run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
