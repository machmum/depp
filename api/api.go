package api

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/go-playground/validator"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/machmum/depp/config"
	"github.com/machmum/depp/middleware/logging"
	"github.com/machmum/depp/middleware/secure"

	"net/http"
	"time"
	"encoding/json"
	"errors"
	"github.com/machmum/depp/api/oauth"
	"github.com/machmum/depp/utl/server"
	"github.com/jinzhu/gorm"
	"github.com/machmum/depp/utl/redis"
	log "github.com/sirupsen/logrus"
	"github.com/machmum/depp/utl/logger"
)

// Custom errors
var (
	ErrFailedConnMysql = errors.New("Failed connecting to database")
	ErrFailedConnRedis = errors.New("Failed connecting to redis")
	ErrMethodNotFound  = errors.New("method not found !")
)

// use new model
func Start(cfg *config.Configuration) {
	// Open a connection
	// store *DB to config
	db, err := gorm.Open("mysql", cfg.DB.Mysql)
	if err != nil {
		customLogDebug(cfg, err) // log real error
		panic(ErrFailedConnMysql)
	}
	defer db.Close()

	// configure custom log for gorm
	db.SetLogger(gorm.Logger{log.New()})
	db.LogMode(cfg.Debug)

	// set db
	cfg.Conn.Mysql = db

	// Open redis connection
	// don't need to close connection, go-redis will do
	redisConfig := redis.NewRedisConfig(cfg)
	rdb, err := redis.Open(redisConfig)
	if err != nil {
		customLogDebug(cfg, err) // log real error
		panic(ErrFailedConnRedis)
	}

	cfg.Conn.Redis = rdb.Client

	// Initialize echo
	// begin
	e := echo.New()
	e.Debug = cfg.Debug // read debug from config

	// register
	// middleware
	e.Use(
		middleware.Recover(),
		secure.CORS(),
		secure.Headers(),
		logging.MiddlewareLogging, // access_log
	)

	// register
	// new http error handler
	e.HTTPErrorHandler = customHTTPErrorHandler

	// register
	// request validator
	e.Validator = &server.CustomValidator{V: validator.New()}

	// register
	// new logger
	logger := logger.New()

	// register routes
	// group
	v1 := e.Group("/v1")

	// execute routes
	//oauth.NewHTTP(a, v1)
	oauth.NewHTTP(oauth.NewService(cfg, logger), v1)

	// prepare server
	s := &http.Server{
		Addr:         cfg.Server.Port,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}

	e.Logger.Fatal(e.StartServer(s))
}

// handle default error
func customHTTPErrorHandler(err error, c echo.Context) {

	logging.MakeLogEntry(c).Error(err)

	// handle http error
	switch err.(type) {
	default:
		// do nothing
	case validator.ValidationErrors:
		err = errors.New(err.Error())
	case *echo.HTTPError:
		parseError, ok := err.(*echo.HTTPError).Internal.(*json.UnmarshalTypeError)
		if ok {
			err = errors.New(parseError.Error())
		} else {
			err = ErrMethodNotFound
		}
	}

	server.ResponseFail(c, err)
}

// handle debug log
func customLogDebug(cfg *config.Configuration, err error) {
	if cfg.Debug != false {
		log.Printf("error %v", err.Error())
	}
}
