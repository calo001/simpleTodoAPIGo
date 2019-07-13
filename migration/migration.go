package migration

import (
	"github.com/jinzhu/gorm"
	"todoAPI/model"
)

func Migrate(db *gorm.DB) {
	db.AutoMigrate(&model.Task{})
	db.AutoMigrate(&model.User{})
}