package model

// Location ...
type Location struct {
	Model
	Location   string `json:"location"`
	BuildingID uint   `json:"-"`
}

// Building ...
type Building struct {
	Model
	Building  string     `json:"building"`
	Locations []Location `json:"locations" gorm:"foreignkey:BuildingID"`
}
