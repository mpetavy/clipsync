package main

import (
	"github.com/mpetavy/common"
	"github.com/mpetavy/common/sqldb"
	"gorm.io/gorm"
)

type Bookmark struct {
	ID        sqldb.FieldInt64  `json:"id" desc:"Unique database ID"`
	CreatedAt sqldb.FieldTime   `json:"createdAt" gorm:"index" desc:"Timestamp this DB record has been created"`
	UpdatedAt sqldb.FieldTime   `json:"updatedAt" gorm:"index" desc:"Timestamp this DB record has been updated the last time"`
	Username  sqldb.FieldString `json:"username" gorm:"index" desc:"Username"`
	Password  sqldb.FieldString `json:"password" gorm:"index" desc:"Password"`
	Payload   sqldb.FieldString `json:"payload" gorm:"index" desc:"Payload"`
}

var BookmarkSchema = struct {
	TableName string
	ID        string
	CreatedAt string
	UpdatedAt string
	Username  string
	Password  string
	Payload   string
}{}

type EventSync struct {
	Sync *Bookmark
}

func (sync *Bookmark) AfterSave(tx *gorm.DB) error {
	// GORM updates call this func with empty Job
	if !sync.ID.Valid {
		return nil
	}

	common.Events.Emit(EventSync{
		Sync: sync,
	}, false)

	return nil
}

func NewBookmark() (*Bookmark, error) {
	sync := &Bookmark{}

	return sync, nil
}
