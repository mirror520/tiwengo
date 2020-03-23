package model

// Department ...
type Department struct {
	Model
	Department    string `json:"department"`
	OU            string `json:"ou" gorm:"unique"`
	InstitutionID uint   `json:"-"`
}

// Institution ...
type Institution struct {
	Model
	Institution string       `json:"institution"`
	OU          string       `json:"ou" gorm:"unique"`
	Departments []Department `json:"departments" gorm:"foreignkey:InstitutionID"`
}

// DepartmentEmployee ...
type DepartmentEmployee struct {
	Model
	EmployeeUserID uint       `json:"-"`
	DepartmentID   uint       `json:"-"`
	Department     Department `json:"department"`
	Employee       User       `json:"employee" gorm:"foreignkey:EmployeeUserID"`
}
