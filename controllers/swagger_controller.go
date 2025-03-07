package controllers

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func GetSwaggerIndex() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ginSwagger.WrapHandler(
			swaggerfiles.Handler,
			ginSwagger.DocExpansion("none"),
			ginSwagger.DefaultModelsExpandDepth(-1),
		)
	}
}
