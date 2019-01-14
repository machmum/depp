package secure

import (
	"time"
	"strings"
	"bytes"
	"github.com/gofrs/uuid"
	"sync"
)

var (
	tokenLen = 6
	tokeType = "Bearer"
)

type TokenSource interface {
	SetToken(int, map[string]interface{}) error
}

type (
	// Used when build token
	tokenize struct {
		token string
		err   error
	}

	// New object token
	token struct {
		// AccessToken is the token that authorizes and authenticates
		// the requests.
		AccessToken string `json:"access_token"`

		// TokenType is the type of token.
		// The Type method returns either this or "Bearer", the default.
		TokenType string `json:"token_type,omitempty"`

		// RefreshToken is a token that's used by the application
		// (as opposed to the user) to refresh the access token
		// if it expires.
		RefreshToken string `json:"refresh_token,omitempty"`

		// Expiry is the optional expiration time of the access token.
		//
		// If zero, TokenSource implementations will reuse the same
		// token forever and RefreshToken or equivalent
		// mechanisms for that TokenSource will not be used.
		ExpiryAccess time.Duration `json:"expiry_access,omitempty"`

		// Expiry is the optional expiration time of the access token.
		//
		// If zero, TokenSource implementations will reuse the same
		// token forever and RefreshToken or equivalent
		// mechanisms for that TokenSource will not be used.
		ExpiryRefresh time.Duration `json:"expiry_refresh,omitempty"`

		Username string
		Password string
	}
)

func NewToken() (*token) {
	return new(token)
}

// Type returns t.TokenType if non-empty, else "Bearer".
func (t *token) Type() string {
	if strings.EqualFold(t.TokenType, "bearer") {
		return "Bearer"
	}
	if strings.EqualFold(t.TokenType, "mac") {
		return "MAC"
	}
	if strings.EqualFold(t.TokenType, "basic") {
		return "Basic"
	}
	if t.TokenType != "" {
		return t.TokenType
	}
	return "Bearer"
}

func (t *token) SetToken(exp int, channel map[string]interface{}) error {
	var err error
	var wg sync.WaitGroup

	tk := make(chan *tokenize)

	go func() {
		wg.Wait()
		close(tk)
	}()

	for i := 0; i < 2; i++ {

		wg.Add(1)
		go t.doToken(&wg, tk)
	}

	// check error
	i := 0
	for data := range tk {
		if data.err != nil {
			return data.err
		}

		if i == 0 {
			t.AccessToken = data.token
			t.ExpiryAccess = time.Duration(exp) * time.Second
		} else {
			t.RefreshToken = data.token
			t.ExpiryRefresh = time.Duration(exp) * time.Second
		}
		i++
	}

	// set token type
	t.TokenType = tokeType

	// set raw data
	if val, ok := channel["username"]; ok {
		t.Username = val.(string)
	}

	if val, ok := channel["password"]; ok {
		t.Password = val.(string)
	}

	return err
}

func (t *token) ResetToken(exp int, channel map[string]interface{}) error {
	var err error

	err = t.SetToken(exp, channel)
	if err != nil {
		return err
	}

	return nil
}

func (t *token) doToken(wg *sync.WaitGroup, token chan<- *tokenize) {

	defer wg.Done()

	u, err := uuid.NewV4()
	if err != nil {
		token <- &tokenize{err: err}
		return
	}

	var buffer bytes.Buffer
	for j := 0; j < tokenLen; j++ {
		buffer.WriteString(uuid.Must(u, err).String() + "-")
	}

	time.Sleep(100 * time.Millisecond)

	str := strings.TrimSuffix(buffer.String(), "-")

	token <- &tokenize{token: str}
}
