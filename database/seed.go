package database

import (
	"encoding/json"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/mirror520/tiwengo/model"
)

// Seed ...
func Seed(db *gorm.DB) {
	var count int
	db.Model(&model.User{}).Count(&count)

	if count != 3 {
		db.Create(&model.User{
			Username: "0911111111",
			Name:     "用手機驗證的訪客",
			Guest: model.Guest{
				Phone:       "0911111111",
				PhoneVerify: true,
			},
		})

		db.Create(&model.User{
			Username: "employee",
			Name:     "員工",
			Employee: model.Employee{
				Account: "employee",
				OU:      "010003",
			},
		})

		db.Create(&model.User{
			Username: "0922222222",
			Name:     "用身分證驗證的訪客",
			Guest: model.Guest{
				Phone:        "0922222222",
				PhoneVerify:  false,
				IDCard:       "L123850660",
				IDCardVerify: true,
			},
		})
	}

	institutions := []model.Institution{
		{
			Institution: "秘書處",
			OU:          "387010000A",
			Departments: []model.Department{
				{Department: "處本部", OU: "010001"},
				{Department: "處長", OU: "010001A"},
				{Department: "副處長", OU: "010001B"},
				{Department: "主任秘書", OU: "010001E"},
				{Department: "專門委員", OU: "010001F"},
				{Department: "文檔科", OU: "010002"},
				{Department: "文檔科第一股", OU: "010003"},
				{Department: "文檔科第二股", OU: "010004"},
				{Department: "文檔科第三股", OU: "010005"},
				{Department: "總務科", OU: "010006"},
				{Department: "總務科第一股", OU: "010007"},
				{Department: "總務科第二股", OU: "010008"},
				{Department: "總務科第三股", OU: "010029"},
				{Department: "公共關係科", OU: "010009"},
				{Department: "公共關係科第一股", OU: "010010"},
				{Department: "公共關係科第二股", OU: "010011"},
				{Department: "國際事務科", OU: "010012"},
				{Department: "國際事務科第一股", OU: "010013"},
				{Department: "國際事務科第二股", OU: "010014"},
				{Department: "國際事務科第三股", OU: "010030"},
				{Department: "機要科", OU: "010015"},
				{Department: "機要科第一股", OU: "010016"},
				{Department: "機要科第二股", OU: "010017"},
				{Department: "廳舍管理科", OU: "010018"},
				{Department: "廳舍管理科第一股", OU: "010019"},
				{Department: "廳舍管理科第二股", OU: "010020"},
				{Department: "廳舍管理科第三股", OU: "010021"},
				{Department: "採購管理科", OU: "010022"},
				{Department: "採購管理科第一股", OU: "010023"},
				{Department: "採購管理科第二股", OU: "010024"},
				{Department: "採購管理科第三股", OU: "010025"},
				{Department: "人事室", OU: "010026"},
				{Department: "會計室", OU: "010027"},
				{Department: "政風室", OU: "010028"},
			},
		},
	}

	if true {
		for _, institution := range institutions {
			db.FirstOrCreate(&institution)
		}
	}

	var targetDepartment model.Department
	db.Where("ou = ?", "010003").First(&targetDepartment)

	var targetEmployee model.User
	db.Set("gorm:auto_preload", true).Where("username = ?", "employee").First(&targetEmployee)

	// departments := targetEmployee.Employee.Departments
	// currentDepartment := departments[len(departments)-1]
	var currentDepartmentEmployee model.DepartmentEmployee
	db.Where("department_id = ? AND employee_user_id = ?", targetDepartment.ID, targetEmployee.ID).Last(&currentDepartmentEmployee)
	if targetDepartment.ID != currentDepartmentEmployee.DepartmentID {
		db.Create(&model.DepartmentEmployee{
			DepartmentID:   targetDepartment.ID,
			EmployeeUserID: targetEmployee.ID,
		})
	}
	db.Where("department_id = ? AND employee_user_id = ?", targetDepartment.ID, targetEmployee.ID).Last(&currentDepartmentEmployee)

	var targetGuest model.User
	db.Set("gorm:auto_preload", true).Where("username = ?", "0911111111").First(&targetGuest)
	// db.Model(&targetGuest.Guest).Association("VisitedDepartments").Append(currentDepartmentEmployee)
	db.Create(&model.Visit{
		GuestUserID:          targetGuest.ID,
		DepartmentEmployeeID: currentDepartmentEmployee.ID,
	})

	db.Set("gorm:auto_preload", true).
		Where("username = ?", "0911111111").
		Preload("Guest.VisitedDepartments.Department").
		Preload("Guest.VisitedDepartments.Employee").
		First(&targetGuest)

	b, _ := json.MarshalIndent(targetGuest, "", " ")
	fmt.Printf("%s", b)
}
