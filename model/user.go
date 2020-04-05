package model

// UserType ...
type UserType int

const (
	// EmployeeUser ...
	EmployeeUser UserType = iota

	// GuestUser ...
	GuestUser
)

// User ...
type User struct {
	Model
	Username string   `json:"username" gorm:"unique;not null" binding:"required"`
	Password string   `json:"password" gorm:"-"`
	Name     string   `json:"name"`
	Type     UserType `json:"type"`
	Employee Employee `json:"employee" gorm:"foreignkey:UserID"`
	Guest    Guest    `json:"guest" gorm:"foreignkey:UserID"`
	Token    Token    `json:"token" gorm:"-"`
}

// Employee ...
type Employee struct {
	UserID      uint         `json:"user_id" gorm:"primary_key;auto_increment:false"`
	Name        string       `json:"name"`
	Account     string       `json:"account" gorm:"unique"`
	Title       string       `json:"title"`
	Departments []Department `json:"departments" gorm:"many2many:department_employees"`
}

// Guest ...
type Guest struct {
	UserID             uint                 `json:"user_id" gorm:"primary_key;auto_increment:false"`
	Name               string               `json:"name"`
	Phone              string               `json:"phone"`
	PhoneVerify        bool                 `json:"phone_verify" gorm:"default:false"`
	PhoneToken         string               `json:"phone_token"`
	PhoneOTP           string               `json:"phone_otp" gorm:"-"`
	IDCard             string               `json:"id_card"`
	IDCardVerify       bool                 `json:"id_card_verify" gorm:"default:false"`
	VisitedDepartments []DepartmentEmployee `json:"visited_departments" gorm:"many2many:visits"`
}

// TccgUser ...
type TccgUser struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name"`
	Title    string `json:"title"`
	Enabled  bool   `json:"enabled"`
	OU       string `json:"ou"`
}

// User ...
func (tccgUser *TccgUser) User() User {
	user := User{
		Username: tccgUser.Account,
		Name:     tccgUser.Name,
		Type:     EmployeeUser,
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
	var username string
	if guest.IDCard != "" {
		username = guest.IDCard
	} else {
		username = guest.Phone
	}

	user := User{
		Username: username,
		Name:     guest.Name,
		Type:     GuestUser,
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
