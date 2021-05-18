package model

// Location ...
type Location struct {
	Model
	Location   string   `json:"location"`
	BuildingID uint     `json:"-"`
	Building   Building `json:"building" gorm:"foreignkey:BuildingID"`
	Capacity   uint     `json:"capacity" gorm:"default:0"`
	Current    uint     `json:"current" gorm:"default:0"`
}

// Building ...
type Building struct {
	Model
	Building  string     `json:"building" gorm:"unique"`
	Locations []Location `json:"locations" gorm:"foreignkey:BuildingID"`
}
