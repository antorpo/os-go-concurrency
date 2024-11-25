package application

import (
	"net/http"

	"github.com/antorpo/os-go-concurrency/pkg/otel"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(app *Application) {
	// OTel middleware
	app.Router.Use(otel.Middleware())

	// Pprof endpoints
	pprof.Register(app.Router)

	// Application endpoints
	app.Router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	app.Router.POST("/products", app.ProductController.ProcessProducts)
}
