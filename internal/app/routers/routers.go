package routers

import (
	"net/http"
	"zertuserver/internal/app/controllers"

	"sync"

	"github.com/gin-gonic/gin"
)

var apiOnce sync.Once
var g *gin.Engine

func SetUp() *gin.Engine {
	apiOnce.Do(func() {
		g = gin.Default()

		g.Static("/assets", "./dist/assets")
		// 加载HTML模板
		//g.LoadHTMLGlob("dist/*")
		g.LoadHTMLFiles("./dist/index.html")
		g.GET("/", func(context *gin.Context) {
			context.HTML(http.StatusOK, "index.html", gin.H{})
		})

		// 跨域中间件
		// g.Use(corsMiddleware())

		mainGroup := g.Group("/zertu")
		mainGroup.GET("/health", controllers.Health)

	})

	return g
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Origin")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	}
}
