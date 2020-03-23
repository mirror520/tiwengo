package model

// User ...
type User struct {
	Model
	Username string   `json:"username" gorm:"unique;not null"`
	Password string   `json:"password" gorm:"-"`
	Name     string   `json:"name"`
	Employee Employee `json:"employee" gorm:"foreignkey:UserID"`
	Guest    Guest    `json:"guest" gorm:"foreignkey:UserID"`
}

// Employee ...
type Employee struct {
	UserID      uint         `json:"user_id" gorm:"primary_key;auto_increment:false"`
	Name        string       `json:"name"`
	Account     string       `json:"account" gorm:"unique" json:"account"`
	Title       string       `json:"title"`
	Departments []Department `json:"departments" gorm:"many2many:department_employees"`
}

// Guest ...
type Guest struct {
	UserID             uint                 `json:"user_id" gorm:"primary_key;auto_increment:false"`
	Name               string               `json:"name"`
	Phone              string               `json:"phone" gorm:"unique"`
	PhoneVerify        bool                 `json:"phone_verify" gorm:"default:false"`
	PhoneToken         string               `json:"phone_token"`
	PhoneOTP           string               `json:"phone_otp" gorm:"-"`
	IDCard             string               `json:"idcard" gorm:"unique"`
	IDCardVerify       bool                 `json:"idcard_verify" gorm:"default:false"`
	VisitedDepartments []DepartmentEmployee `json:"visited_departments" gorm:"many2many:visits"`
}

// TccgUser ...
type TccgUser struct {
	Account        string         `json:"account"`
	Password       string         `json:"password"`
	Name           string         `json:"name"`
	Title          string         `json:"title"`
	Enabled        bool           `json:"enabled"`
	TccgDepartment TccgDepartment `json:"department"`
}

// TccgDepartment ...
type TccgDepartment struct {
	Department string `json:"department"`
	OU         string `json:"ou"`
}

// User ...
func (tccgUser *TccgUser) User() User {
	user := User{
		Username: tccgUser.Account,
		Name:     tccgUser.Name,
		Employee: Employee{
			Account: tccgUser.Account,
			Name:    tccgUser.Name,
			Title:   tccgUser.Title,
		},
	}

	return user
}

// User ...
func (guest *Guest) User() User {
	user := User{
		Username: guest.Phone,
		Name:     guest.Name,
		Guest: Guest{
			Name:         guest.Name,
			Phone:        guest.Phone,
			PhoneVerify:  false,
			IDCard:       guest.IDCard,
			IDCardVerify: false,
		},
	}

	return user
}
