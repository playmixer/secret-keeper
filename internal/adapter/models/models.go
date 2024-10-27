package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Login        string `gorm:"index:,unique"`
	PasswordHash string
}

type DataType string

const (
	CARD     DataType = "CARD"
	PASSWORD DataType = "PASSWORD"
	TEXT     DataType = "TEXT"
	BINARY   DataType = "BINARY"
)

type Secret struct {
	User User
	gorm.Model
	Title     string
	DataType  DataType `sql:"type:ENUM('CARD', 'PASSWORD', 'TEXT', 'BINARY')" gorm:"data_type"`
	MetaName  string
	Data      []byte
	UserID    uint
	UpdateDT  int64
	IsDeleted bool
}

type Text struct {
	Title string `json:"Title"`
	Text  string `json:"Text"`
}

type Card struct {
	Title  string `json:"Title"`
	Number string `json:"Number"`
	PIN    string `json:"PIN"`
	CVV    string `json:"CVV"`
	Expiry string `json:"Expiry"`
}

type Password struct {
	Title    string `json:"Title"`
	Site     string `json:"Site"`
	Login    string `json:"Login"`
	Password string `json:"Password"`
}

type Binary struct {
	Title    string `json:"Title"`
	Filename string `json:"Filename"`
}

type MetaDataItem struct {
	Data      *[]byte
	Title     string
	DataType  DataType
	ID        uint
	UpdatedDT int64
	IsDeleted bool
}

type FileMetaDataItem struct {
	Title        string   `json:"title"`
	OriginalPath string   `json:"original_path"`
	DataType     DataType `json:"type"`
	Filename     string   `json:"filename"`
	ID           int64    `json:"id"`
	ExternalID   uint     `json:"external_id"`
	UpdateDT     int64    `json:"update_dt"`
	IsDeleted    bool     `json:"is_deleted"`
	IsUpdated    bool     `json:"is_updated"`
}
