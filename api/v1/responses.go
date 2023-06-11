package v1

import (
	"encoding/xml"
	"time"

	"github.com/gin-gonic/gin"
)

type XMLResponse struct {
	XMLName xml.Name `xml:"Data" json:"-"`
}

type JSONResponse[T any] struct {
	Data T `json:"data" xml:"-"`
}

type AssumeRoleOutput struct {
	XMLResponse
	AccessKeyId     string `json:"accessKeyid"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
	Expiration      time.Time
}

type GetConsoleUrlOutput struct {
	XMLResponse
	ConsoleUrl string `json:"consoleUrl"`
}

type GetUserInfoOutput struct {
	XMLResponse
	Username string   `json:"username" type:"string"`
	Groups   []string `json:"groups" type:"slice"`
}

func renderResponse(ctx *gin.Context, statusCode int, body interface{}) {
	switch ctx.Request.Header.Get("Accept") {
	case "application/xml":
		ctx.XML(statusCode, body)
	default:
		switch b := body.(type) {
		case *RestError:
			ctx.JSON(b.Status, JSONError{Error: b})
		default:
			ctx.JSON(statusCode, JSONResponse[interface{}]{Data: b})
		}
	}
}
