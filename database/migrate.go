package database

import (
	"github.com/jinzhu/gorm"
	"github.com/mirror520/tiwengo/model"
)

// Migrate ...
func Migrate(db *gorm.DB) {
	db.AutoMigrate(
		&model.Visit{}, &model.DepartmentEmployee{}, &model.Follower{},
		&model.User{}, &model.Employee{}, &model.Guest{},
		&model.Department{}, &model.Institution{},
		&model.Location{}, &model.Building{},
		&model.TcpassVisit{},
	)
}
