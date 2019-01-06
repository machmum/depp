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
	"fmt"
)

var (
	GrantAccess  = "password"
	GrantRefresh = "refresh"
)

type (
	NewConfig struct {
		*config.Configuration
	}

	NewToken struct {
		AccessToken   string        `json:"access_token"`
		TokenType     string        `json:"token_type,omitempty"`
		RefreshToken  string        `json:"refresh_token,omitempty"`
		ExpiryAccess  time.Duration `json:"expiry_access,omitempty"`
		ExpiryRefresh time.Duration `json:"expiry_refresh,omitempty"`
	}
)

func NewService(cfg *config.Configuration) NewConfig {
	return NewConfig{
		&config.Configuration{
			Debug:  cfg.Debug,
			Server: cfg.Server,
			DB:     cfg.DB,
			App:    cfg.App,
			Conn:   cfg.Conn,
			Redis:  cfg.Redis,
		},
	}

}

func NewHTTP(cfg NewConfig, er *echo.Group) error {
	//var err error

	g := er.Group("/oauth")
	g.POST("/token", cfg.RequestToken)

	return nil
}

func zz() error {
	log.Fatal("test")
	return nil
}

func (cfg NewConfig) RequestToken(c echo.Context) (err error) {
	cred := new(depp.Credentials)

	err = c.Bind(cred)
	if err != nil {
		return
	}

	err = c.Validate(cred)
	if err != nil {
		return
	}

	// check token
	var channel depp.ChannelVersion

	key := "channel_" + cred.Username
	rds, err := cfg.Conn.Redis.HGetAll(cfg.Redis.Prefix.Apps).Result()
	if err != nil {
		return err
	}

	if len(rds) != 0 {
		if err := json.Unmarshal([]byte(rds[key]), &channel); err != nil {
			return err
		}

		fmt.Println(channel)

	} else {
		channel, err = cred.GetChannelVersion(cfg.Conn.Mysql)
		if err != nil {
			return err
		}

		if channel.ApiChannelID != 0 {
			bytes, err := json.Marshal(channel)
			if err != nil {
				return err
			}

			// write redis
			if _, err = cfg.Conn.Redis.HMSet(cfg.Redis.Prefix.Apps, map[string]interface{}{key: bytes}).Result(); err != nil {
				return err
			}

			// set expire redis
			duration := time.Duration(cfg.Redis.Lifetime.Apps) * time.Second
			if _, err = cfg.Conn.Redis.Expire(cfg.Redis.Prefix.Apps, duration).Result(); err != nil {
				return err
			}
		}
	}

	// set token
	if cred.GrantType == GrantAccess {
		//
		// do check refresh token first
		//

		// do get token
		token, err := secure.SetToken(cfg.Redis.Lifetime.Access)
		if err != nil {
			return err
		}

		newtoken := NewToken{
			AccessToken:   token.AccessToken,
			RefreshToken:  token.RefreshToken,
			TokenType:     token.TokenType,
			ExpiryRefresh: token.ExpiryRefresh,
			ExpiryAccess:  token.ExpiryAccess,
		}

		// save token to redis
		err = setRedis(cfg, newtoken)
		if err != nil {
			return err
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

	return errors.New("something error occurred")
}

func setRedis(cfg NewConfig, t NewToken) error {
	var err error

	refresh, err := json.Marshal(map[string]interface{}{
		"access_token": t.AccessToken,
		"token_type":   t.TokenType,
		"expires_in":   t.ExpiryRefresh / time.Second,
	})
	if err != nil {
		return err
	}

	access, err := json.Marshal(map[string]interface{}{
		"refresh_token": t.RefreshToken,
		"token_type":    t.TokenType,
		"expires_in":    t.ExpiryAccess / time.Second,
	})
	if err != nil {
		return err
	}

	// write redis
	key := cfg.Redis.Prefix.Refresh + "_" + t.RefreshToken
	expired := t.ExpiryRefresh - (3 * time.Second)
	if _, err := cfg.Conn.Redis.HMSet(key, map[string]interface{}{
		"token": refresh,
	}).Result(); err != nil {
		return err
	}

	if _, err := cfg.Conn.Redis.Expire(key, expired).Result(); err != nil {
		log.Fatal("masuk")
		return err
	}

	key = cfg.Redis.Prefix.Access + "_" + t.AccessToken
	expired = t.ExpiryAccess - (3 * time.Second)
	if _, err := cfg.Conn.Redis.HMSet(key, map[string]interface{}{
		"token": access,
	}).Result(); err != nil {
		return err
	}

	// set expire redis

	if _, err := cfg.Conn.Redis.Expire(key, expired).Result(); err != nil {
		log.Fatal("masuk")
		return err
	}

	return nil
}
