package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ITestController interface {
	CreateItem(ctx *gin.Context)
}

type testController struct{}

func NewTestController() ITestController {
	return &testController{}
}

func (c *testController) CreateItem(ctx *gin.Context) {
	type Person struct {
		name string
	}

	xd := Person{
		name: "Arcangel",
	}

	ctx.JSON(http.StatusOK, xd)
}
