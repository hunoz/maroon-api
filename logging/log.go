package logging

import (
	"encoding/json"
	"io"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type UTCFormatter struct {
	logrus.Formatter
}

func (u UTCFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	return u.Formatter.Format(e)
}

// GetClientIP gets the correct IP for the end client instead of the proxy
func getClientIP(c *gin.Context) string {
	// first check the X-Forwarded-For header
	requester := c.Request.Header.Get("X-Forwarded-For")
	// if empty, check the Real-IP header
	if len(requester) == 0 {
		requester = c.Request.Header.Get("X-Real-IP")
	}
	// if the requester is still empty, use the hard-coded address from the socket
	if len(requester) == 0 {
		requester = c.Request.RemoteAddr
	}

	// if requester is a comma delimited list, take the first one
	// (this happens when proxied via elastic load balancer then again through nginx)
	if strings.Contains(requester, ",") {
		requester = strings.Split(requester, ",")[0]
	}

	return requester
}

// GetUserID gets the current_user ID as a string
func getUserID(c *gin.Context) string {
	userID, exists := c.Get("userID")
	if exists {
		return userID.(string)
	}
	return ""
}

func SetLogMode(stage string) {
	if strings.ToLower(stage) == "prod" {
		logrus.SetFormatter(UTCFormatter{&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05Z",
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				s := strings.Split(f.Function, ".")
				funcname := s[len(s)-1]
				// Locations such as init have different endings in their split strings
				if funcname == "func1" || funcname == "0" {
					funcname = s[len(s)-2]
				}
				_, filename := path.Split(f.File)
				return funcname, filename
			},
			DisableHTMLEscape: true,
		}})
	} else {
		logrus.SetFormatter(UTCFormatter{&logrus.TextFormatter{
			TimestampFormat: "2006-01-02T15:04:05Z",
			DisableColors:   true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				s := strings.Split(f.Function, ".")
				funcname := s[len(s)-1]
				// Locations such as init have different endings in their split strings
				if funcname == "func1" || funcname == "0" {
					funcname = s[len(s)-2]
				}
				_, filename := path.Split(f.File)
				return funcname, filename
			},
			ForceQuote: true,
		}})
	}
	logrus.SetReportCaller(true)
}

// JSONLogMiddleware logs a gin HTTP request in JSON format, with some additional custom key/values
func JSONLogMiddleware(stage string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process Request
		c.Next()

		// Stop timer
		duration := time.Since(start)

		bodyInBytes, _ := io.ReadAll(c.Request.Body)

		var unmarshalledBody any
		json.Unmarshal(bodyInBytes, &unmarshalledBody)

		entry := logrus.WithFields(logrus.Fields{
			"client_ip":  getClientIP(c),
			"duration":   duration,
			"method":     c.Request.Method,
			"path":       c.Request.RequestURI,
			"status":     c.Writer.Status(),
			"user_id":    getUserID(c),
			"referrer":   c.Request.Referer(),
			"request_id": c.Writer.Header().Get("Request-Id"),
		})

		// Only log body for beta so debugging is possible
		if strings.ToLower(stage) == "beta" {
			entry = entry.WithField("body", unmarshalledBody)
		}

		if c.Writer.Status() >= 500 {
			entry.Error(c.Errors.String())
		} else {
			entry.Info("")
		}
	}
}
