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

	guestUsername := ctx.Param("username")

	var guest model.User
	db.Set("gorm:auto_preload", true).Where("username = ?", guestUsername).First(&guest)
	if guest.ID == 0 {
		// 這裡應該要再做一次身分證驗證...，先依賴前端跳過

		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("您使用身分證驗證，第一次需要登錄資料")
		result.SetData(guestUsername)
		ctx.AbortWithStatusJSON(http.StatusConflict, result)
		return
	}
	logger = logger.WithFields(log.Fields{"guest": guest.Username})

	var employee model.User
	employeeUsername, ok := ctx.Get("username")
	if !ok {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("找不到員工使用者")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}
	db.Set("gorm:auto_preload", true).Where("username = ?", employeeUsername).First(&employee)
	logger = logger.WithFields(log.Fields{"employee": employee.Username})

	departments := employee.Employee.Departments
	department := departments[len(departments)-1]
	var departmentEmployee model.DepartmentEmployee
	db.Where("department_id = ? AND employee_user_id = ?", department.ID, employee.ID).Last(&departmentEmployee)

	var input model.Location
	var location model.Location
	ctx.ShouldBind(&input)
	if input.ID != 0 {
		db.Where("id = ?", input.ID).First(&location)
		logger = logger.WithFields(log.Fields{"location": location.Location})
	}

	visit := model.Visit{
		GuestUserID:          guest.ID,
		DepartmentEmployeeID: departmentEmployee.ID,
		LocationID:           location.ID,
	}
	db.Create(&visit)
	db.Where("id =?", visit.ID).
		Preload("Guest").
		Preload("DepartmentEmployee.Employee").
		Preload("DepartmentEmployee.Department").
		Preload("Location").
		First(&visit)

	var msg string
	if location.ID == 0 {
		msg = fmt.Sprintf("「%s」已登記訪客「%s」到「%s」洽公", employee.Name, guest.Name, department.Department)
	} else {
		msg = fmt.Sprintf("「%s」已登記訪客「%s」到「%s」洽公", employee.Name, guest.Name, location.Location)
	}
	result = model.NewSuccessResult().SetLogger(logger)
	result.AddInfo(msg)
	result.SetData(visit)

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
	db.Preload("Building").Find(&locations)

	logger.Infoln("成功取得所有地點")

	ctx.JSON(http.StatusOK, &locations)
}

// ListAllGuestVisitRecordHandler ...
func ListAllGuestVisitRecordHandler(ctx *gin.Context) {
	logger := log.WithFields(log.Fields{
		"controller": "Visit",
		"event":      "ListAllGuestVisitRecord",
	})

	var db *gorm.DB = model.DB
	var result *model.Result

	var visits []model.Visit
	db.Preload("Guest").
		Preload("DepartmentEmployee.Employee").
		Preload("DepartmentEmployee.Department").
		Preload("Location").
		Find(&visits)

	result = model.NewSuccessResult().SetLogger(logger)
	result.AddInfo("成功取得所有拜訪紀錄")
	result.SetData(visits)

	ctx.JSON(http.StatusOK, result)
}
