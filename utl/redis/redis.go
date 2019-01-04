package redis

import (
	"time"
	"github.com/go-redis/redis"
	"github.com/machmum/depp/config"
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

var client *redis.Client

type (
	//Client struct{}

	NewConfig struct {
		*config.Configuration
	}

	NewClient struct {
		*redis.Client
	}
)

func NewRedisConfig(cfg *config.Configuration) NewConfig {
	return NewConfig{
		&config.Configuration{
			Debug: cfg.Debug,
			Redis: cfg.Redis,
		},
	}
}

func Open(cfg NewConfig) (c NewClient, err error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Host + cfg.Redis.Port,
		Password:     cfg.Redis.Password,
		PoolTimeout:  30 * time.Second,
		IdleTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	_, err = rdb.Ping().Result()
	if err != nil {
		return c, err
	}

	c = NewClient{rdb}

	return c, nil
}

func Open2(cfg NewConfig) (client *redis.Client, err error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Host + cfg.Redis.Port,
		Password:     cfg.Redis.Password,
		PoolTimeout:  30 * time.Second,
		IdleTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	_, err = rdb.Ping().Result()
	if err != nil {
		return client, err
	}

	return rdb, nil
}

// Set redis
func SetRedis(key string, data interface{}, expiration int, cfg *redis.Client) (err error) {

	//for i := range data {
	//	if data[i], err = json.Marshal(data[i]); err != nil {
	//		err = handleError(err)
	//		return
	//	}
	//}

	// write redis
	//if _, err = client.HMSet(key, data).Result(); err != nil {
	//	err = handleError(err)
	//	return
	//}
	//
	//// set expire redis
	//if _, err = client.Expire(key, exp).Result(); err != nil {
	//	err = handleError(err)
	//}

	return nil
}

func (client NewClient) SetRedis2(key string, data map[string]interface{}, expiration int) (err error) {
	exp := time.Duration(expiration) * time.Second
	for i := range data {
		if data[i], err = json.Marshal(data[i]); err != nil {
			err = handleError(err)
			return
		}
	}

	// write redis
	if _, err = client.HMSet(key, data).Result(); err != nil {
		err = handleError(err)
		return
	}

	// set expire redis
	if _, err = client.Expire(key, exp).Result(); err != nil {
		err = handleError(err)
	}

	return
}

// handle error redis
func handleError(err error) (e error) {
	//if conf.GetBool(`debug`) != true {
	//	e = errors.New("redis went wrong [gateway]")
	//	return
	//}
	log.Println("something")

	return err
}

/*

// Get redis
func GetRedis(key string) (result map[string]string, err error) {
	if result, err = client.HGetAll(key).Result(); err != nil {
		err = handleError(err)
	}

	return
}

// Delete redis
func DeleteRedis(key string, index string) (err error) {
	if _, err = client.HDel(key, index).Result(); err != nil {
		err = handleError(err)
	}

	return
}

// Delete all redis hash
func DeleteAllRedis(key string) (err error) {
	if _, err = client.Del(key).Result(); err != nil {
		err = handleError(err)
	}

	return
}

// Rename redis
func RenameRedis(key string, newkey string) (err error) {
	if _, err = client.Rename(key, newkey).Result(); err != nil {
		err = handleError(err)
	}

	return
}
*/

//func GetRedisData(key string, accessToken string) (redis Redisdata, err error) {
//
//	var val string
//	var exist bool
//
//	// get access token
//	get, err := GetRedis(accessToken)
//	if err != nil {
//		return
//	}
//
//	if val, exist = get["token"]; !exist {
//		err = errors.New("something wrong, invalid or expired access token [get]")
//		return
//
//	} else {
//
//		err = json.Unmarshal([]byte(val), &redis.Refresh)
//
//		// get value access token
//		redis.Refresh.RefreshToken = cons.RefreshTokenPrefix + redis.Refresh.RefreshToken
//
//		get, err := GetRedis(redis.Refresh.RefreshToken)
//		if err != nil {
//			return redis, err
//		}
//
//		// split key
//		scopes := strings.Split(key, "+")
//
//		for _, scope := range scopes {
//
//			if val, exist = get[scope]; !exist {
//
//				// only flag error if cant find profile in redis
//				if key == "profile" {
//					err = errors.New("something wrong, " + key + " not found [get]")
//				}
//
//			} else {
//
//				// get profile
//				if scope == "profile" {
//					err = json.Unmarshal([]byte(val), &redis.Profile)
//					if err != nil {
//						return redis, err
//					}
//				}
//
//				// get menu
//				if scope == "menu" {
//					err = json.Unmarshal([]byte(val), &redis.Menu)
//					if err != nil {
//						return redis, err
//					}
//				}
//
//				// get role
//				if scope == "role" {
//					err = json.Unmarshal([]byte(val), &redis.Role)
//					if err != nil {
//						return redis, err
//					}
//				}
//			}
//		}
//
//		return redis, err
//	}
//}
