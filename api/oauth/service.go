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
	log "github.com/sirupsen/logrus"
	"encoding/json"
	"github.com/machmum/depp/utl/logger"
	"reflect"
)

var (
	GrantAccess  = "password"
	GrantRefresh = "refresh"
	Count        int
	key          string
	oldkey       string
)

type (
	Service struct {
		Config NewConfig
		Log    *logger.Log
	}

	NewConfig struct {
		*config.Configuration
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
		Config: NewConfig{
			&config.Configuration{
				Debug: cfg.Debug,
				Redis: cfg.Redis,
				Conn:  cfg.Conn,
			},
		},
		Log: l,
	}
}

func NewHTTP(s Service, er *echo.Group) error {
	var err error

	g := er.Group("/oauth")
	g.POST("/token", s.RequestToken)

	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (s Service) RequestToken(c echo.Context) (err error) {
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

	lifetime := s.Config.Redis.Lifetime.Access
	raw := map[string]interface{}{
		"username": channel.Username,
		"password": channel.Password,
	}

	// set token
	if cred.GrantType == GrantAccess {
		//
		// do check refresh token first
		//

		// do get token
		token := secure.NewToken()

		log.Fatalln(reflect.TypeOf(token))

		err = token.SetToken2(lifetime, raw)
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
		err = s.setRedis(newtoken)
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

	// refresh token
	if cred.GrantType == GrantRefresh {
		token := secure.NewToken()

		// get token from redis
		oldkey = s.Config.Redis.Prefix.Refresh + cred.RefreshToken
		tkn, err := s.Config.Conn.Redis.HGetAll(oldkey).Result()
		if err != nil {
			return s.LogError(err)
		}

		if len(tkn) != 0 {
			if err := json.Unmarshal([]byte(tkn["token"]), &token); err != nil {
				return s.LogError(err)
			}

			if cred.Username == token.Username && cred.Password == token.Password {

				oldtoken := token.AccessToken

				// set new token
				err = token.SetToken2(lifetime, raw)
				if err != nil {
					return s.LogError(err)
				}

				// delete old access token
				key = s.Config.Redis.Prefix.Access + oldtoken
				if _, err = s.Config.Conn.Redis.HDel(key, "token").Result(); err != nil {
					return s.LogError(err)
				}

				// move refresh token
				key = s.Config.Redis.Prefix.Refresh + token.RefreshToken
				if _, err = s.Config.Conn.Redis.Rename(oldkey, key).Result(); err != nil {
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
				err = s.setRedis(newtoken)
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
				return errors.New("token invalid or expired")
			}

		} else {
			return errors.New("token not found")
		}
	}

	return errors.New("something error occurred")
}

func (s Service) setRedis(t NewToken) error {
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
	key = s.Config.Redis.Prefix.Refresh + t.RefreshToken
	expired := t.ExpiryRefresh - (3 * time.Second)

	if _, err = s.Config.Conn.Redis.HMSet(key, map[string]interface{}{"token": refresh}).Result(); err != nil {
		return err
	}

	if _, err := s.Config.Conn.Redis.Expire(key, expired).Result(); err != nil {
		return err
	}

	// write access
	key = s.Config.Redis.Prefix.Access + t.AccessToken
	expired = t.ExpiryAccess - (3 * time.Second)
	if _, err := s.Config.Conn.Redis.HMSet(key, map[string]interface{}{"token": access}).Result(); err != nil {
		return err
	}

	if _, err := s.Config.Conn.Redis.Expire(key, expired).Result(); err != nil {
		return err
	}

	return nil
}

func (s Service) setChannel(cred *depp.Credentials) (channel depp.ChannelVersion, err error) {

	var failedGetChannel = "failed to get channel"

	key = "channel_" + cred.Username
	rds, err := s.Config.Conn.Redis.HGetAll(s.Config.Redis.Prefix.Apps).Result()
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
		channel, err = cred.GetChannelVersion(s.Config.Conn.Mysql)
		if err != nil {
			return channel, s.LogError(err)
		}

		if channel.ApiChannelID == 0 {
			err = errors.New(failedGetChannel)
			return channel, err

		} else {
			bytes, err := json.Marshal(channel)
			if err != nil {
				return channel, s.LogError(err)
			}

			// write redis
			if _, err = s.Config.Conn.Redis.HMSet(s.Config.Redis.Prefix.Apps, map[string]interface{}{key: bytes}).Result(); err != nil {
				return channel, s.LogError(err)
			}

			// set expire redis
			duration := time.Duration(s.Config.Redis.Lifetime.Apps) * time.Second
			if _, err = s.Config.Conn.Redis.Expire(s.Config.Redis.Prefix.Apps, duration).Result(); err != nil {
				return channel, s.LogError(err)
			}
		}
	}

	return channel, nil
}

func (s Service) LogError(err error) error {
	s.Log.Log(
		"oauth",
		"failed in oauth/service",
		err,
		nil,
	)

	return err
}
