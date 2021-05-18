package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/mirror520/tiwengo/model"

	log "github.com/sirupsen/logrus"
)

// UserVisitHandler ...
func UserVisitHandler(ctx *gin.Context) {
	logger := log.WithFields(log.Fields{
		"controller": "Visit",
		"event":      "UserVisit",
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

	var input model.Visit
	var location model.Location
	var lastVisit model.Visit
	var isLeave bool = false
	ctx.ShouldBind(&input)
	if input.LocationID != 0 {
		db.Where("id = ?", input.LocationID).First(&location)
		logger = logger.WithFields(log.Fields{"location": location.Location})

		if location.Capacity > 0 {
			now := time.Now()
			todayStart := fmt.Sprintf("%s 00:00:00", now.Format("2006-01-02"))
			db.Where("guest_user_id = ? AND location_id = ? AND created_at > ?", guest.ID, location.ID, todayStart).
				Preload("Followers").
				Last(&lastVisit)

			if lastVisit.ID != 0 && !lastVisit.Leave {
				location.Current -= uint(len(lastVisit.Followers)) + 1
				isLeave = true
			} else {
				location.Current++
				isLeave = false
			}
		}
	}

	visit := model.Visit{
		GuestUserID:          guest.ID,
		DepartmentEmployeeID: departmentEmployee.ID,
		LocationID:           location.ID,
		Leave:                isLeave,
	}

	db.Create(&visit)

	if input.Followers != nil {
		followers := ""
		for _, follower := range input.Followers {
			db.Create(&model.Follower{
				VisitID: visit.ID,
				Name:    follower.Name,
			})

			followers += fmt.Sprintf("[%s]", follower.Name)
		}

		if location.ID != 0 && location.Capacity > 0 {
			if lastVisit.ID == 0 || lastVisit.Leave {
				location.Current += uint(len(input.Followers))
			}
		}

		logger = logger.WithFields(log.Fields{"followrs": followers})
	}

	if location.ID != 0 && location.Capacity > 0 {
		db.Save(location)

		model.MQTTClient.Publish(
			fmt.Sprintf("location/%d", location.ID),
			0,
			true,
			strconv.Itoa(int(location.Current)),
		).Wait()

		logger.WithFields(log.Fields{
			"service":  "mqtt",
			"location": location.Location,
		}).Infoln("發送地點更新人數")
	}

	db.Where("id = ?", visit.ID).
		Preload("Guest").
		Preload("DepartmentEmployee.Employee").
		Preload("DepartmentEmployee.Department").
		Preload("Location").
		Preload("Location.Building").
		Preload("Followers").
		First(&visit)

	visit.Mask()

	result = model.NewSuccessResult().SetLogger(logger)
	result.AddInfo("已登記訪客洽公")
	result.SetData(visit)

	ctx.JSON(http.StatusOK, result)
}

// GetBuildingsHandler ...
func GetBuildingsHandler(ctx *gin.Context) {
	logger := log.WithFields(log.Fields{
		"controller": "Visit",
		"event":      "GetBuildings",
	})

	var db *gorm.DB = model.DB

	var buildings []model.Building
	db.Preload("Locations").
		Find(&buildings)

	logger.Infoln("成功取得所有建物地點")

	ctx.JSON(http.StatusOK, &buildings)
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
		Preload("Location.Building").
		Find(&visits)

	result = model.NewSuccessResult().SetLogger(logger)
	result.AddInfo("成功取得所有拜訪紀錄")
	result.SetData(visits)

	ctx.JSON(http.StatusOK, result)
}

// TcpassUserVisitHandler ...
func TcpassUserVisitHandler(ctx *gin.Context) {
	logger := log.WithFields(log.Fields{
		"controller": "Visit",
		"event":      "TcpassUserVisit",
	})

	var db *gorm.DB = model.DB
	var result *model.Result

	uuid := ctx.Param("uuid")
	logger = logger.WithFields(log.Fields{"uuid": uuid})

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

	var input model.TcpassVisit
	var location model.Location
	var lastVisit model.TcpassVisit
	var isLeave bool = false
	ctx.ShouldBind(&input)
	if input.LocationID != 0 {
		db.Where("id = ?", input.LocationID).First(&location)
		logger = logger.WithFields(log.Fields{"location": location.Location})

		if location.Capacity > 0 {
			now := time.Now()
			todayStart := fmt.Sprintf("%s 00:00:00", now.Format("2006-01-02"))
			db.Where("uuid = ? AND location_id = ? AND created_at > ?", uuid, location.ID, todayStart).
				Last(&lastVisit)

			if lastVisit.ID != 0 && !lastVisit.Leave {
				location.Current--
				isLeave = true
			} else {
				location.Current++
				isLeave = false
			}
		}
	}

	visit := model.TcpassVisit{
		DepartmentEmployeeID: departmentEmployee.ID,
		LocationID:           location.ID,
		UUID:                 uuid,
		Leave:                isLeave,
	}

	db.Create(&visit)

	if location.ID != 0 && location.Capacity > 0 {
		db.Save(location)

		model.MQTTClient.Publish(
			fmt.Sprintf("location/%d", location.ID),
			0,
			true,
			strconv.Itoa(int(location.Current)),
		).Wait()

		logger.WithFields(log.Fields{
			"service":  "mqtt",
			"location": location.Location,
		}).Infoln("發送地點更新人數")
	}

	db.Where("id = ?", visit.ID).
		Preload("DepartmentEmployee.Employee").
		Preload("DepartmentEmployee.Department").
		Preload("Location").
		Preload("Location.Building").
		First(&visit)

	visit.Mask()

	result = model.NewSuccessResult().SetLogger(logger)
	result.AddInfo("已登記訪客洽公")
	result.SetData(visit)

	ctx.JSON(http.StatusOK, result)
}
