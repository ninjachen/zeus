package middleware

import (
	"go.yym.plus/zeus/pkg/log"
	"github.com/gin-gonic/gin"
	"time"
)

// Logger is logger  middleware
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		ip := c.ClientIP()
		req := c.Request
		path := req.URL.Path
		c.Next()
		form := req.Form
		var err error

		err, _ = c.Value("error").(error)

		dt := time.Since(now)

		lf := log.Infow
		errMsg := ""
		isSlow := dt >= (time.Millisecond * 500)
		if err != nil {
			errMsg = err.Error()
			lf = log.Errorw
		} else {
			if isSlow {
				lf = log.Warnw
			}
		}
		data, _ := c.Get("data")
		lf("http request",
			"method", req.Method,
			"ip", ip,
			"path", path,
			"query", req.URL.Query().Encode(),
			"form", form.Encode(),
			"data", data,
			"status", c.Writer.Status(),
			"err", errMsg,
			"ts", dt.Seconds(),
		)
	}
}
