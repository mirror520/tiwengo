package model

// TcpassVisit ...
type TcpassVisit struct {
	Model
	DepartmentEmployeeID uint               `json:"-"`
	LocationID           uint               `json:"location_id"`
	DepartmentEmployee   DepartmentEmployee `json:"department_employee" gorm:"DepartmentEmployeeID"`
	Location             Location           `json:"location" gorm:"foreignkey:LocationID"`
	UUID                 string             `json:"uuid"`
}

// Mask ...
func (visit *TcpassVisit) Mask() {
	visit.DepartmentEmployee.Employee.Mask(VisitMask)
}
