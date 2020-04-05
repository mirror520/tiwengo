package model

// Visit ...
type Visit struct {
	Model
	GuestUserID          uint               `json:"-"`
	DepartmentEmployeeID uint               `json:"-"`
	LocationID           uint               `json:"-"`
	Guest                User               `json:"guest" gorm:"foreignkey:GuestUserID"`
	DepartmentEmployee   DepartmentEmployee `json:"department_employee" gorm:"DepartmentEmployeeID"`
	Location             Location           `json:"location" gorm:"foreignkey:LocationID"`
}
