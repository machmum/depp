package auth

import (
	"github.com/labstack/echo"
	"github.com/machmum/depp/config"
	"net/http"
	"github.com/machmum/depp/utl/server"
)

type (
	Credentials struct {
		GrantType    string `json:"grant_type" validate:"required,oneof=password refresh"`
		Username     string `json:"username" validate:"required"`
		Password     string `json:"password" validate:"required"`
		RefreshToken string `json:"refresh_token" validate:"omitempty"`
	}

	NewConfig struct {
		Cfg *config.Configuration
	}
)

func NewService(cfg *config.Configuration) NewConfig {
	return NewConfig{
		Cfg: &config.Configuration{
			Debug:  cfg.Debug,
			Server: cfg.Server,
			DB:     cfg.DB,
			App:    cfg.App,
			Conn:   cfg.Conn,
		},
	}
}

func NewHTTP(cfg NewConfig, er *echo.Group) error {
	//var err error

	g := er.Group("/user")
	g.POST("/token", cfg.UserAuth)

	return nil
}

func (cfg NewConfig) UserAuth(c echo.Context) (err error) {
	c.JSON(http.StatusOK, server.NewResponse{
		Meta: server.ResponseMeta{
			Code:    http.StatusOK,
			Message: "routes found",
		},
	})

	//cred := new(conn.Credentials)
	//
	//err = c.Bind(&cred)
	//if err != nil {
	//	log.Fatalf("err 1 %v", err.Error())
	//	return
	//}
	//
	//err = c.Validate(cred)
	//if err != nil {
	//	log.Fatalf("err 2 %v", err.Error())
	//	return
	//}
	//
	//log.Printf("got password, is %v", cred.Password)
	//log.Printf("got username, is %v", cred.Username)
	//
	//result := depp.GetChannelVersion2(cred, cfg.Cfg.Conn)
	//
	//if result.ApiChannelID != 0 {
	//	c.JSON(http.StatusOK, server.NewResponse{
	//		Meta: server.ResponseMeta{
	//			Code:    http.StatusOK,
	//			Message: "",
	//		},
	//		Data: result,
	//	})
	//}

	return nil
}
