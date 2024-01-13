package middlewares

import "github.com/gin-gonic/gin"

const (
	UsernameKey = "username"
	UserIdKey   = "userid"
	UserIPKey   = "userip"
	// 还可添加角色或者权限信息（放入验证token中）
)

func Context() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 可自定义扩展添加一些逻辑
		//c.Set(UserIPKey,ip)
		c.Next()
	}
}
