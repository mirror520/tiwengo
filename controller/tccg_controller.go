package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/jinzhu/gorm"
	"github.com/mirror520/tiwengo/environment"
	"github.com/mirror520/tiwengo/model"

	log "github.com/sirupsen/logrus"
)

var tccgBaseURL = environment.TCCGBaseURL

// LoginTccgUserHandler ...
func LoginTccgUserHandler(input *model.TccgUser) (*model.User, error) {
	logger := log.WithFields(log.Fields{
		"controller": "Tccg",
		"event":      "LoginTccgUser",
		"username":   input.Account,
	})

	var db *gorm.DB = model.DB
	var enforcer *casbin.Enforcer = model.Enforcer

	tccgUser, err := login(input)
	if err != nil {
		return nil, err
	}

	// if !tccgUser.Enabled {
	// 	return nil, errors.New("不合法的使用者")
	// }
	logger.Infoln("成功取得使用者公務帳號資料")

	var user model.User
	db.Set("gorm:auto_preload", true).Where("username = ?", tccgUser.Account).First(&user)
	if user.Username != tccgUser.Account {
		user = tccgUser.User()
		db.Create(&user)
		logger.Infoln("使用者第一次登入系統，建立使用者")

		authPath := fmt.Sprintf("/api/v1/guests/%d/qr", user.ID)
		enforcer.AddNamedPolicy("p", user.Username, authPath, "GET")
		enforcer.AddRoleForUser(user.Username, "tccg_user")
		logger.Infoln("新增使用者權限")
	}

	var targetDepartment model.Department
	db.Where("ou = ?", tccgUser.OU).First(&targetDepartment)

	var currentDepartmentEmployee model.DepartmentEmployee
	db.Where("department_id = ? AND employee_user_id = ?", targetDepartment.ID, user.ID).Last(&currentDepartmentEmployee)
	if targetDepartment.ID != currentDepartmentEmployee.DepartmentID {
		db.Create(&model.DepartmentEmployee{
			DepartmentID:   targetDepartment.ID,
			EmployeeUserID: user.ID,
		})
		logger.Infoln("使用者變更所屬單位，加入使用者新的單位")

		db.Set("gorm:auto_preload", true).Where("id = ?", user.ID).First(&user)
	}
	logger.Infoln("使用者登入成功")

	return &user, nil
}

func login(user *model.TccgUser) (*model.TccgUser, error) {
	client := &http.Client{}
	b, _ := json.Marshal(user)

	req, err := http.NewRequest("PATCH", tccgBaseURL+"/login", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var result model.Result
		json.NewDecoder(resp.Body).Decode(&result)
		return nil, errors.New(result.Info[0])
	}

	user = &model.TccgUser{}
	json.NewDecoder(resp.Body).Decode(&user)

	return user, nil
}
