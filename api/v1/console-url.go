package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AccessType string

const (
	AccessTypeReadOnly AccessType = "ReadOnly"
	AccessTypeAdmin    AccessType = "Administrator"
)

var AccessTypes = [2]string{string(AccessTypeAdmin), string(AccessTypeReadOnly)}

type UrlCredentials struct {
	SessionId    string `json:"sessionId" url:"sessionId"`
	SessionKey   string `json:"sessionKey" url:"sessionKey"`
	SessionToken string `json:"sessionToken" url:"sessionToken"`
}

type SignInTokenResponse struct {
	SignInToken string `json:"SigninToken" binding:"required"`
}

func isValidAccessType(accessType AccessType) bool {
	for _, aType := range AccessTypes {
		if string(accessType) == aType {
			return true
		}
	}
	return false
}

// https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_enable-console-custom-url.html#STSConsoleLink_programPython
// This required that you be using IAM user credentials. Perhaps fetching from Secrets Manager then assuming role?
func GetConsoleUrl(ctx *gin.Context) {
	input := GetConsoleUrlInput{}
	username, _ := ctx.Get("username")

	if err := ctx.ShouldBindQuery(&input); err != nil {
		err := parseBindingError(err)
		renderResponse(ctx, err.Status, err)
		return
	}

	if !isValidAccessType(input.AccessType) {
		renderResponse(ctx, 400, BadRequestError())
		return
	}

	var iamRoleName string
	if input.AccessType == AccessTypeAdmin {
		iamRoleName = "MaroonApiAdminAccessRole-DO-NOT-DELETE"
	} else {
		iamRoleName = "MaroonApiReadOnlyAccessRole-DO-NOT-DELETE"
	}

	credentials, err := assumeRole(fmt.Sprintf("arn:aws:iam::%s:role/%s", input.AccountId, iamRoleName), username.(string), int32(input.Duration), nil)
	if err != nil {
		logrus.Errorf("Error assuming role '%s': %s", iamRoleName, err.Error())
		var e *RestError
		if strings.Contains(err.Error(), "is not authorized to perform") {
			e = ForbiddenError()
		} else {
			e = BadRequestError()
		}
		renderResponse(ctx, e.Status, e)
		return
	}

	urlCredentials := UrlCredentials{
		SessionId:    *credentials.AccessKeyId,
		SessionKey:   *credentials.SecretAccessKey,
		SessionToken: *credentials.SessionToken,
	}

	var jsonCredentials []byte
	if jsonCredentials, err = json.Marshal(urlCredentials); err != nil {
		logrus.Errorf("Error parsing credentials: %s", err.Error())
		err := InternalServerError()
		renderResponse(ctx, err.Status, err)
		return
	}

	federationUrlParameters := fmt.Sprintf("?Action=getSigninToken&SessionDuration=%v&Session=%s", input.Duration, url.QueryEscape(string(jsonCredentials)))

	federationUrl := fmt.Sprintf("https://signin.aws.amazon.com/federation%s", federationUrlParameters)

	response, err := http.Get(federationUrl)
	if err != nil {
		logrus.Errorf("Error getting sign in token: %s", err.Error())
		err := InternalServerError()
		renderResponse(ctx, err.Status, err)
		return
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		logrus.Errorf("Error reading sign in token: %s", err.Error())
		err := InternalServerError()
		renderResponse(ctx, err.Status, err)
		return
	}

	signInToken := SignInTokenResponse{}

	if err = json.Unmarshal(body, &signInToken); err != nil {
		logrus.Errorf("Error unmarshaling sign in token: %s", err.Error())
		err := InternalServerError()
		renderResponse(ctx, err.Status, err)
		return
	}

	federationUrlParameters = fmt.Sprintf(
		"?Action=login&Issuer=MaroonApi&Destination=%s&SigninToken=%s",
		url.QueryEscape("https://console.aws.amazon.com/"),
		url.QueryEscape(signInToken.SignInToken),
	)

	federationUrl = fmt.Sprintf(
		"https://signin.aws.amazon.com/federation%s",
		federationUrlParameters,
	)

	renderResponse(ctx, 200, GetConsoleUrlOutput{
		ConsoleUrl: federationUrl,
	})
}
