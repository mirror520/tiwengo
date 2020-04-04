package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/mirror520/tiwengo/model"

	log "github.com/sirupsen/logrus"
)

// UserVisitHandler ...
func UserVisitHandler(ctx *gin.Context) {
	logger := log.WithFields(log.Fields{
		"controller": "Visit",
		"event":      "GuestVisit",
	})

	var db *gorm.DB = model.DB
	var result *model.Result

	guestID := ctx.Param("user")

	var guest model.User
	db.Set("gorm:auto_preload", true).Where("id = ?", guestID).First(&guest)
	if guest.ID == 0 {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("找不到訪客使用者")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}
	logger = logger.WithFields(log.Fields{"guest_user_id": guest.ID})

	var employee model.User
	username, ok := ctx.Get("username")
	if !ok {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("找不到員工使用者")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}
	db.Set("gorm:auto_preload", true).Where("username = ?", username).First(&employee)
	logger = logger.WithFields(log.Fields{"employee_user_id": employee.ID})

	departments := employee.Employee.Departments
	department := departments[len(departments)-1]
	var departmentEmployee model.DepartmentEmployee
	db.Where("department_id = ? AND employee_user_id = ?", department.ID, employee.ID).Last(&departmentEmployee)

	var input model.Location
	var location model.Location
	ctx.ShouldBind(&input)
	if input.ID != 0 {
		db.Where("id = ?", input.ID).First(&location)
		logger = logger.WithFields(log.Fields{"location_id": location.ID})
	}

	db.Create(&model.Visit{
		GuestUserID:          guest.ID,
		DepartmentEmployeeID: departmentEmployee.ID,
		LocationID:           location.ID,
	})

	var msg string
	if location.ID == 0 {
		msg = fmt.Sprintf("「%s」已登記訪客「%s」到「%s」洽公", employee.Name, guest.Name, department.Department)
	} else {
		msg = fmt.Sprintf("「%s」已登記訪客「%s」到「%s」洽公", employee.Name, guest.Name, location.Location)
	}
	result = model.NewSuccessResult().SetLogger(logger)
	result.AddInfo(msg)

	ctx.JSON(http.StatusOK, result)
}

// GetLocationsHandler ...
func GetLocationsHandler(ctx *gin.Context) {
	logger := log.WithFields(log.Fields{
		"controller": "Visit",
		"event":      "GetLocations",
	})

	var db *gorm.DB = model.DB

	var locations []model.Location
	db.Find(&locations)

	logger.Infoln("成功取得所有地點")

	ctx.JSON(http.StatusOK, &locations)
}
