package oauth

import (
	"github.com/labstack/echo"
	"github.com/machmum/depp/config"
	"net/http"
	"github.com/machmum/depp/utl/server"
	"github.com/machmum/depp/utl/models"
)

type (
	//Credentials struct {
	//	GrantType    string `json:"grant_type" validate:"required,oneof=password refresh"`
	//	Username     string `json:"username" validate:"required"`
	//	Password     string `json:"password" validate:"required"`
	//	RefreshToken string `json:"refresh_token" validate:"omitempty"`
	//}

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

		// set to redis
		//err := redis.SetRedis("", result, 0)
		//if err != nil {
		//	return err
		//}

		c.JSON(http.StatusOK, server.NewResponse{
			Meta: server.ResponseMeta{
				Code:    http.StatusOK,
				Message: "",
			},
			Data: result,
		})
	}

	return nil
}
