package v1

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type User struct {
	Username string   `json:"username" type:"string"`
	Groups   []string `json:"groups" type:"slice"`
}

func GetUserInfo(context *gin.Context) {
	username := context.GetString("username")
	userGroups, _ := context.Get("groups")

	groups := make([]string, len(userGroups.([]interface{})))

	for i, v := range userGroups.([]interface{}) {
		groups[i] = fmt.Sprint(v)
	}

	context.JSON(200, User{
		Username: username,
		Groups:   groups,
	})
}
