package gateway

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterUIRoutes(r *gin.Engine) {

	r.LoadHTMLGlob("web/templates/*")

	r.Static("/static","./web/static")

	r.GET("/",func(c *gin.Context) {
		token,err := c.Cookie("token")
		if err != nil || token == "" {
			c.Redirect(http.StatusFound,"/login")
			return 
		}
		c.HTML(http.StatusOK,"dashboard.html",nil)
	})

	r.GET("/login",func(c *gin.Context){
		token,err := c.Cookie("token")
		if err == nil || token != "" {
			c.Redirect(http.StatusFound,"/")
			return
		}
		c.HTML(http.StatusOK,"login.html",nil)
	})

	r.GET("/admin",func(c *gin.Context) {
		c.HTML(http.StatusOK,"admin_login.html",nil)
	})

	r.GET("/payment-success",func(c *gin.Context) {
		c.HTML(http.StatusOK,"payment_success.html",nil)
	})

	r.GET("/payment-cancel",func(c *gin.Context) {
		c.HTML(http.StatusOK,"payment_cancel.html",nil)
	})
}