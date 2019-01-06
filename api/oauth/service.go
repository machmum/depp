package oauth

import (
	"github.com/labstack/echo"
	"github.com/machmum/depp/config"
	"net/http"
	"github.com/machmum/depp/utl/server"
	"github.com/machmum/depp/utl/models"
	"time"
	"github.com/gin-gonic/gin/json"
	"errors"
	"github.com/machmum/depp/utl/secure"
	log "github.com/sirupsen/logrus"
)

var (
	GrantAccess  = "password"
	GrantRefresh = "refresh"
)

type (
	NewConfig struct {
		*config.Configuration
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

//type NewToken struct {
//	*secure.Token
//}

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

	//result, err := depp.GetChannelVersion(cred, cfg.Cfg.Conn)
	result, err := cred.GetChannelVersion2(cfg.Conn.Mysql)
	if err != nil {
		return err
	}

	// set to redis
	if result.ApiChannelID != 0 {

		bytes, err := json.Marshal(result)
		if err != nil {
			return err
		}

		// write redis
		if _, err = cfg.Conn.Redis.HMSet(cfg.Redis.Prefix.Apps, map[string]interface{}{"channel_" + result.Username: bytes}).Result(); err != nil {
			return err
		}

		// set expire redis
		duration := time.Duration(cfg.Redis.Lifetime.Apps) * time.Second
		if _, err = cfg.Conn.Redis.Expire(cfg.Redis.Prefix.Apps, duration).Result(); err != nil {
			return err
		}

		if cred.GrantType == GrantAccess {
			//
			// do check refresh token first
			//

			// do get token
			//var token NewToken
			//log.Println(token)
			token := secure.SetToken()
			if err != nil {
				return err
			}

			log.Fatalln(token)

		}

		c.JSON(http.StatusOK, server.NewResponse{
			Meta: server.ResponseMeta{
				Code:    http.StatusOK,
				Message: "",
			},
			Data: result,
		})
	}

	return errors.New("something error occurred")
}

//func generateToken(p map[string]string) (result interface{}, err error) {
//	// generate token
//	at := common.GenerateRandomUuid()
//	rt := common.GenerateRandomUuid()
//	start := time.Now().Unix()
//
//	// prepare access token for redis
//	atPrefix := cons.AccessTokenPrefix + at
//	dataAt := map[string]interface{}{
//		"token": map[string]interface{}{
//			"refresh_token": rt,
//			"token_type":    cons.TokenType,
//			"expire_in":     cons.AccessTokenExpire,
//			"start_time":    start,
//		},
//	}
//
//	// prepare refresh token for redis
//	rtPrefix := cons.RefreshTokenPrefix + rt
//	dataRt := map[string]interface{}{
//		"token": map[string]interface{}{
//			"access_token": at,
//			"token_type":   cons.TokenType,
//			"username":     r.Username,
//			"password":     r.Password,
//			"expire_in":    cons.RefreshTokenExpire,
//			"start_time":   start,
//		},
//	}
//
//	// delete old access_token
//	if p["at"] != "" {
//		common.DeleteRedis(p["ct"], "csrf_token")
//		common.DeleteRedis(p["at"], "token")
//		common.RenameRedis(p["rt"], rtPrefix)
//	}
//
//	// insert redis
//	if err = common.SetRedis(atPrefix, dataAt, cons.AccessTokenExpire); err != nil {
//		return
//	}
//	if err = common.SetRedis(rtPrefix, dataRt, cons.RefreshTokenExpire); err != nil {
//		return
//	}
//
//	// populate data response
//	result = GenerateTokenStruct{
//		AccessToken:  at,
//		TokenType:    cons.TokenType,
//		ExpireIn:     cons.AccessTokenExpire,
//		RefreshToken: rt,
//	}
//
//	return
//}
