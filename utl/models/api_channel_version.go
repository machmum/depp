package depp

import (
	"github.com/jinzhu/gorm"
	"time"
)

type (
	Credentials struct {
		GrantType    string `json:"grant_type" validate:"required,oneof=password refresh"`
		Username     string `json:"username" validate:"required"`
		Password     string `json:"password" validate:"required"`
		RefreshToken string `json:"refresh_token" validate:"omitempty"`
	}

	ChannelVersion struct {
		ID           uint       `json:"id,omitempty" gorm:"primary_key"`
		ApiChannelID int        `json:"api_channel_id,omitempty" gorm:"column:api_channel_id"`
		Version      string     `json:"version,omitempty" gorm:"omitempty"`
		Username     string     `json:"username,omitempty"`
		Password     string     `json:"password,omitempty"`
		CreatedAt    *time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
		UpdatedAt    *time.Time `json:"updated_at,omitempty" gorm:"column:updated_at"`
		DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"column:deleted_at,omitempty"`
	}
)

func GetChannelVersion(cred Credentials, db *gorm.DB) (result ChannelVersion, err error) {

	if err := db.Raw("select * from api_channel_version where username = ?", cred.Username).Scan(&result).Error; err != nil {
		//fmt.Println(reflect.TypeOf(err))
		return result, err
	}

	return result, nil
	//db.Debug().Raw("select * from api_channel_version where username = ?", cred.Username).Scan(&result)
	//db.LogMode(true)

	//return result

}

func (cred Credentials) GetChannelVersion2(db *gorm.DB) (result ChannelVersion, err error) {

	if err := db.Raw("select api_channel_id, username, password from api_channel_version where username = ?", cred.Username).Scan(&result).Error; err != nil {
		//fmt.Println(reflect.TypeOf(err))
		return result, err
	}

	return result, nil
	//db.Debug().Raw("select * from api_channel_version where username = ?", cred.Username).Scan(&result)
	//db.LogMode(true)

	//return result

}
