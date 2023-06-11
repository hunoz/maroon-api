package v1

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func GetUserInfo(ctx *gin.Context) {
	username := ctx.GetString("username")
	userGroups, _ := ctx.Get("groups")

	groups := make([]string, len(userGroups.([]interface{})))

	for i, v := range userGroups.([]interface{}) {
		groups[i] = fmt.Sprint(v)
	}

	renderResponse(ctx, 200, GetUserInfoOutput{
		Username: username,
		Groups:   groups,
	})
}
