package v1

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"zertuserver/internal/app/controllers"
	"zertuserver/internal/app/models"
)

func HydrologyDeviceDirect(c *gin.Context) {
	direct := c.DefaultQuery("direct", "")

	log.Info("direction info : ", direct)

	//if err := services.RtuService().Direction(direct); err != nil {
	//	log.Error("adjust direct error: ", err)
	//	controllers.Response(c, 500, "Success", "")
	//	return
	//}

	controllers.Response(c, 200, "Success", "")
}

func HydrologyBaseData(c *gin.Context) {
	data := models.HydrologyBaseData{
		Distance: 10,
		Depth:    10,
		Surface:  12,
		Bottom:   13,
		Speed:    0,
	}

	log.Info("direction info : ", data)

	controllers.Response(c, 200, "Success", data)
}
