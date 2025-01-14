package v1

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"zertuserver/internal/app/controllers"
	"zertuserver/internal/app/models"
)

func WifiSetting(c *gin.Context) {
	infos := models.WifiInfos{}
	if err := c.ShouldBindJSON(&infos); err != nil {
		controllers.Response(c, 500, "error", "")
		return
	}

	log.Info(infos)

	controllers.Response(c, 200, "Success", "")
}
