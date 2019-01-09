package oauth

import (
	"github.com/labstack/echo"
	"github.com/machmum/depp/config"
	"net/http"
	"github.com/machmum/depp/utl/server"
	"github.com/machmum/depp/utl/models"
	"time"
	"errors"
	"github.com/machmum/depp/utl/secure"
	"encoding/json"
	"github.com/machmum/depp/utl/logger"
)

var (
	// return error in processing service
	ErrInvalidToken    = errors.New("token invalid or expired")
	ErrNotFoundToken   = errors.New("token not found")
	ErrNotFoundChannel = errors.New("channel not found")

	// return error for log
	LogSrc = "oauth"
	LogMsg = "failed occurred in oauth/service"

	// token grant_type
	GrantAccess  = "password"
	GrantRefresh = "refresh"

	// token redis
	key    string
	oldkey string
)

type (
	Service struct {
		*config.Configuration
		*logger.Log
	}

	NewToken struct {
		AccessToken   string        `json:"access_token"`
		TokenType     string        `json:"token_type,omitempty"`
		RefreshToken  string        `json:"refresh_token,omitempty"`
		ExpiryAccess  time.Duration `json:"expiry_access,omitempty"`
		ExpiryRefresh time.Duration `json:"expiry_refresh,omitempty"`
		Username      string        `json:"username,omitempty"`
		Password      string        `json:"password,omitempty"`
	}
)

func NewService(cfg *config.Configuration, l *logger.Log) Service {
	return Service{
		Configuration: cfg,
		Log:           l,
	}
}

func NewHTTP(s Service, er *echo.Group) {

	g := er.Group("/oauth")

	g.POST("/token", s.Tokenize)
}

func (s Service) Tokenize(c echo.Context) (err error) {
	cred := new(depp.Credentials)

	err = c.Bind(cred)
	if err != nil {
		return s.LogError(err)
	}

	err = c.Validate(cred)
	if err != nil {
		return s.LogError(err)
	}

	// get channel
	channel, err := s.setChannel(cred)
	if err != nil {
		return err
	}

	raw := map[string]interface{}{
		"username": channel.Username,
		"password": channel.Password,
	}

	// initiate token
	token := secure.NewToken()

	// get access token
	if cred.GrantType == GrantAccess {

		err = token.SetToken(s.Redis.Lifetime.Access, raw)
		if err != nil {
			return s.LogError(err)
		}

		newtoken := NewToken{
			AccessToken:   token.AccessToken,
			RefreshToken:  token.RefreshToken,
			TokenType:     token.TokenType,
			ExpiryRefresh: token.ExpiryRefresh,
			ExpiryAccess:  token.ExpiryAccess,
			Username:      token.Username,
			Password:      token.Password,
		}

		// save token to redis
		err = s.setTokenRedis(newtoken)
		if err != nil {
			return s.LogError(err)
		}

		c.JSON(http.StatusOK, server.NewResponse{
			Meta: server.ResponseMeta{
				Code:    http.StatusOK,
				Message: "",
			},
			Data: map[string]interface{}{
				"access_token":  newtoken.AccessToken,
				"refresh_token": newtoken.RefreshToken,
				"token_type":    newtoken.TokenType,
				"expire_in":     newtoken.ExpiryAccess / time.Second,
			},
		})

		return nil
	}

	// get refresh token
	if cred.GrantType == GrantRefresh {

		// get token from redis
		oldkey = s.Redis.Prefix.Refresh + cred.RefreshToken
		trds, err := s.Conn.Redis.HGetAll(oldkey).Result()
		if err != nil {
			return s.LogError(err)
		}

		if len(trds) == 0 {
			return ErrNotFoundToken

		} else {
			if err := json.Unmarshal([]byte(trds["token"]), &token); err != nil {
				return s.LogError(err)
			}

			if cred.Username == token.Username && cred.Password == token.Password {

				oldtoken := token.AccessToken

				// set new token
				err = token.SetToken(s.Redis.Lifetime.Access, raw)
				if err != nil {
					return s.LogError(err)
				}

				// delete old access token
				key = s.Redis.Prefix.Access + oldtoken
				if _, err = s.Conn.Redis.HDel(key, "token").Result(); err != nil {
					return s.LogError(err)
				}

				// move refresh token
				key = s.Redis.Prefix.Refresh + token.RefreshToken
				if _, err = s.Conn.Redis.Rename(oldkey, key).Result(); err != nil {
					return s.LogError(err)
				}

				newtoken := NewToken{
					AccessToken:   token.AccessToken,
					RefreshToken:  token.RefreshToken,
					TokenType:     token.TokenType,
					ExpiryRefresh: token.ExpiryRefresh,
					ExpiryAccess:  token.ExpiryAccess,
					Username:      token.Username,
					Password:      token.Password,
				}

				// save token to redis
				err = s.setTokenRedis(newtoken)
				if err != nil {
					return s.LogError(err)
				}

				c.JSON(http.StatusOK, server.NewResponse{
					Meta: server.ResponseMeta{
						Code:    http.StatusOK,
						Message: "",
					},
					Data: map[string]interface{}{
						"access_token":  newtoken.AccessToken,
						"refresh_token": newtoken.RefreshToken,
						"token_type":    newtoken.TokenType,
						"expire_in":     newtoken.ExpiryAccess / time.Second,
					},
				})

				return nil

			} else {
				return ErrInvalidToken
			}
		}
	}

	return nil
}

func (s Service) setTokenRedis(t NewToken) error {
	var err error

	refresh, err := json.Marshal(map[string]interface{}{
		"access_token": t.AccessToken,
		"token_type":   t.TokenType,
		"expires_in":   t.ExpiryRefresh / time.Second,
		"username":     t.Username,
		"password":     t.Password,
	})
	if err != nil {
		return err
	}

	access, err := json.Marshal(map[string]interface{}{
		"refresh_token": t.RefreshToken,
		"token_type":    t.TokenType,
		"expires_in":    t.ExpiryAccess / time.Second,
		"channel":       t.Username,
	})
	if err != nil {
		return err
	}

	// write refresh
	key = s.Redis.Prefix.Refresh + t.RefreshToken
	expired := t.ExpiryRefresh - (3 * time.Second)

	if _, err = s.Conn.Redis.HMSet(key, map[string]interface{}{"token": refresh}).Result(); err != nil {
		return err
	}

	if _, err := s.Conn.Redis.Expire(key, expired).Result(); err != nil {
		return err
	}

	// write access
	key = s.Redis.Prefix.Access + t.AccessToken
	expired = t.ExpiryAccess - (3 * time.Second)
	if _, err := s.Conn.Redis.HMSet(key, map[string]interface{}{"token": access}).Result(); err != nil {
		return err
	}

	if _, err := s.Conn.Redis.Expire(key, expired).Result(); err != nil {
		return err
	}

	return nil
}

func (s Service) setChannel(cred *depp.Credentials) (channel depp.ChannelVersion, err error) {

	key = "channel_" + cred.Username
	rds, err := s.Conn.Redis.HGetAll(s.Redis.Prefix.Apps).Result()
	if err != nil {
		return channel, s.LogError(err)
	}

	if len(rds) != 0 {
		err = json.Unmarshal([]byte(rds[key]), &channel)
		if err != nil {
			return channel, s.LogError(err)
		}

	} else {

		// get channel
		// then set to redis
		channel, err = cred.GetChannelVersion(s.Conn.Mysql)
		if err != nil {
			return channel, s.LogError(err)
		}

		if channel.ApiChannelID == 0 {
			return channel, ErrNotFoundChannel

		} else {
			bytes, err := json.Marshal(channel)
			if err != nil {
				return channel, s.LogError(err)
			}

			// write redis
			if _, err = s.Conn.Redis.HMSet(s.Redis.Prefix.Apps, map[string]interface{}{key: bytes}).Result(); err != nil {
				return channel, s.LogError(err)
			}

			// set expire redis
			duration := time.Duration(s.Redis.Lifetime.Apps) * time.Second
			if _, err = s.Conn.Redis.Expire(s.Redis.Prefix.Apps, duration).Result(); err != nil {
				return channel, s.LogError(err)
			}
		}
	}

	return channel, nil
}

func (s Service) LogError(err error) error {
	s.Log.Log(
		LogSrc,
		LogMsg,
		err,
	)

	return err
}
