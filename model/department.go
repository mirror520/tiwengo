package model

import "github.com/jinzhu/gorm"

// Department ...
type Department struct {
	gorm.Model
	Department    string
	OU            string `gorm:"unique"`
	InstitutionID uint
}

// Institution ...
type Institution struct {
	gorm.Model
	Institution string
	OU          string       `gorm:"unique"`
	Departments []Department `gorm:"foreignkey:InstitutionID"`
}

// DepartmentEmployee ...
type DepartmentEmployee struct {
	gorm.Model
	EmployeeUserID uint
	DepartmentID   uint
	Department     Department
	Employee       User `gorm:"foreignkey:EmployeeUserID"`
}
