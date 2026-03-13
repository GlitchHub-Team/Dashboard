package common

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func RequestError(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, gin.H{
		"error": err.Error(),
	})
}

func RequestOk(ctx *gin.Context, obj any) {
	ctx.JSON(http.StatusOK, obj)
}
