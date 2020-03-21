package model

import "github.com/jinzhu/gorm"

// Visit ...
type Visit struct {
	gorm.Model
	GuestUserID          uint
	DepartmentEmployeeID uint
}
