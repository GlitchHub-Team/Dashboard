package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var ErrPageNotFound = errors.New("page not found")

func RequestError(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, gin.H{
		"error": err.Error(),
	})
}

func RequestOk(ctx *gin.Context, obj any) {
	ctx.JSON(http.StatusOK, obj)
}

func RequestUnauthorized(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusUnauthorized, gin.H{
		"error": err.Error(),
	})
}

func RequestNotFound(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusNotFound, gin.H{
		"error": err.Error(),
	})
}

func RequestServerError(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusInternalServerError, gin.H{
		"error": err.Error(),
	})
}

func ValidationError(ctx *gin.Context, err error) bool {
	errorFields := gin.H{}
	if validationErr, ok := errors.AsType[validator.ValidationErrors](err); ok {
		for _, fieldError := range validationErr {
			errorFields[fieldError.Field()] = fieldError.Tag()
		}
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":  "invalid format",
			"fields": errorFields,
		})
		return true
	}
	return false
}
