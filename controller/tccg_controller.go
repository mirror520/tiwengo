package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/mirror520/tiwengo/model"

	log "github.com/sirupsen/logrus"
)

const baseURL = "https://api.secret.taichung.gov.tw/v1.0/tccg/users"

// LoginTccgUserHandler ...
func LoginTccgUserHandler(ctx *gin.Context) {
	logger := log.WithFields(log.Fields{
		"controller": "Tccg",
		"event":      "LoginTccgUser",
	})

	var db *gorm.DB = model.DB
	var result *model.Result

	var input model.TccgUser
	err := ctx.ShouldBind(&input)
	if err != nil {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("使用者輸入資料格式錯誤")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}
	logger = logger.WithFields(log.Fields{"account": input.Account})

	tccgUser, err := login(&input)
	if err != nil {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo(err.Error())
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}

	if !tccgUser.Enabled {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("不合法的使用者")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}
	logger.Infoln("成功取得使用者公務帳號資料")

	var user model.User
	db.Set("gorm:auto_preload", true).Where("username = ?", tccgUser.Account).First(&user)
	if user.Username != tccgUser.Account {
		user = tccgUser.User()
		db.Create(&user)

		logger.Infoln("使用者第一次登入系統，建立使用者")
	}

	var targetDepartment model.Department
	db.Where("ou = ?", tccgUser.TccgDepartment.OU).First(&targetDepartment)

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

	result = model.NewSuccessResult().SetLogger(logger)
	result.AddInfo("使用者登入成功")
	result.SetData(&user)

	ctx.JSON(http.StatusOK, result)
}

func login(user *model.TccgUser) (*model.TccgUser, error) {
	client := &http.Client{}
	b, _ := json.Marshal(user)

	req, err := http.NewRequest("PATCH", baseURL+"/login", bytes.NewBuffer(b))
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
