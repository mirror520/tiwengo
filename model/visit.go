package model

// Visit ...
type Visit struct {
	Model
	GuestUserID          uint               `json:"-"`
	DepartmentEmployeeID uint               `json:"-"`
	LocationID           uint               `json:"location_id"`
	Leave                bool               `json:"leave"`
	Guest                User               `json:"guest" gorm:"foreignkey:GuestUserID"`
	DepartmentEmployee   DepartmentEmployee `json:"department_employee" gorm:"DepartmentEmployeeID"`
	Location             Location           `json:"location" gorm:"foreignkey:LocationID"`
	Followers            []Follower         `json:"followers" gorm:"foreignkey:VisitID"`
}

// Follower ...
type Follower struct {
	VisitID uint   `json:"visit_id" gorm:"primary_key;auto_increment:false"`
	Name    string `json:"name" gorm:"primary_key"`
}

// Mask ...
func (visit *Visit) Mask() {
	visit.Guest.Mask(VisitMask)
	visit.DepartmentEmployee.Employee.Mask(VisitMask)
}
