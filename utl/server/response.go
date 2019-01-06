package server

import (
	"github.com/labstack/echo"
	"net/http"
)

type ResponseMeta struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type NewResponse struct {
	Meta ResponseMeta `json:"meta"`
	Data interface{}  `json:"data,omitempty"`
}

// not used
func ResponseFail(c echo.Context, err error) error {
	return c.JSON(http.StatusOK, NewResponse{
		Meta: ResponseMeta{
			Code:    http.StatusNotFound,
			Message: err.Error(),
		},
	})
}

//type NewResponse struct {
//	Code    int    `json:"code"`
//	Message string `json:"message"`
//}
