package model

import "github.com/jinzhu/gorm"

// User ...
type User struct {
	gorm.Model
	Username string   `gorm:"unique;not null"`
	Password string   `gorm:"-"`
	Name     string   `gorm:"default:NULL"`
	Employee Employee `gorm:"foreignkey:UserID"`
	Guest    Guest    `gorm:"foreignkey:UserID"`
}

// Employee ...
type Employee struct {
	UserID      uint         `gorm:"primary_key;auto_increment:false"`
	Account     string       `gorm:"unique;default:NULL"`
	Departments []Department `gorm:"many2many:department_employees"`
	OU          string       `gorm:"-"`
}

// Guest ...
type Guest struct {
	UserID             uint                 `gorm:"primary_key;auto_increment:false"`
	Phone              string               `gorm:"unique;default:NULL"`
	PhoneVerify        bool                 `gorm:"default:false"`
	IDCard             string               `gorm:"unique;default:NULL"`
	IDCardVerify       bool                 `gorm:"default:false"`
	VisitedDepartments []DepartmentEmployee `gorm:"many2many:visits"`
}
