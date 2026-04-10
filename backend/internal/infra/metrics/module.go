package metrics

import (
	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"metrics",
	fx.Invoke(RegisterPrometheus),
)

func RegisterPrometheus(router *gin.Engine) {
	p := ginprometheus.NewPrometheus("gin")
	p.ReqCntURLLabelMappingFn = func(ctx *gin.Context) string {
		if path := ctx.FullPath(); path != "" {
			return path
		}
		return ctx.Request.URL.Path
	}

	p.Use(router)
}
