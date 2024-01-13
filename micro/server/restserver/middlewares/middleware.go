package middlewares

import "github.com/gin-gonic/gin"

var Middlewares = defaultMiddlewares()

func defaultMiddlewares() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"recovery": gin.Recovery(),
		"cors":     Cors(),
		"context":  Context(),
	}
}
