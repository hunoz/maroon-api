package authentication

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

// Auth ...
type Auth struct {
	jwk               *JWK
	jwkURL            string
	cognitoRegion     string
	cognitoUserPoolID string
}

// Config ...
type Config struct {
	CognitoRegion     string
	CognitoUserPoolID string
}

// JWK ...
type JWK struct {
	Keys []struct {
		Alg string `json:"alg"`
		E   string `json:"e"`
		Kid string `json:"kid"`
		Kty string `json:"kty"`
		N   string `json:"n"`
	} `json:"keys"`
}

func NewAuth(config *Config) *Auth {
	a := &Auth{
		cognitoRegion:     config.CognitoRegion,
		cognitoUserPoolID: config.CognitoUserPoolID,
	}

	a.jwkURL = fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", a.cognitoRegion, a.cognitoUserPoolID)
	err := a.CacheJWK()
	if err != nil {
		log.Fatal(err)
	}

	// This will refresh the JWKS every 30 minutes. See https://stackoverflow.com/questions/16466320/is-there-a-way-to-do-repetitive-tasks-at-intervals.
	ticker := time.NewTicker(30 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				logrus.Info("Refreshing JWKS")
				a.CacheJWK()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	return a
}

func (a *Auth) CacheJWK() error {
	req, err := http.NewRequest("GET", a.jwkURL, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	jwk := new(JWK)
	err = json.Unmarshal(body, jwk)
	if err != nil {
		return err
	}

	a.jwk = jwk
	return nil
}

func (a *Auth) ParseJWT(tokenString string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		index := 0
		found := false
		for i, v := range a.jwk.Keys {
			if v.Kid == token.Header["kid"] {
				index = i
				found = true
			}
		}
		if !found {
			return nil, errors.New("key not found")
		}
		if token.Method.Alg() != "RS256" {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		key := convertKey(a.jwk.Keys[index].E, a.jwk.Keys[index].N)

		return key, nil
	})
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func (a *Auth) JWK() *JWK {
	return a.jwk
}

func (a *Auth) JWKURL() string {
	return a.jwkURL
}

// https://gist.github.com/MathieuMailhos/361f24316d2de29e8d41e808e0071b13
func convertKey(rawE, rawN string) *rsa.PublicKey {
	decodedE, err := base64.RawURLEncoding.DecodeString(rawE)
	if err != nil {
		panic(err)
	}
	if len(decodedE) < 4 {
		ndata := make([]byte, 4)
		copy(ndata[4-len(decodedE):], decodedE)
		decodedE = ndata
	}
	pubKey := &rsa.PublicKey{
		N: &big.Int{},
		E: int(binary.BigEndian.Uint32(decodedE[:])),
	}
	decodedN, err := base64.RawURLEncoding.DecodeString(rawN)
	if err != nil {
		panic(err)
	}
	pubKey.N.SetBytes(decodedN)
	return pubKey
}

func JWTMiddleware(auth Auth) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenHeader := ctx.GetHeader("Authorization")
		if tokenHeader == "" {
			ctx.AbortWithStatus(401)
		}

		claims, err := auth.ParseJWT(tokenHeader)
		if err != nil {
			logrus.Errorf("Invalid token")
			ctx.AbortWithStatus(401)
		} else {
			for key, val := range claims {
				if key == "cognito:username" {
					ctx.Set("username", val)
				} else if key == "cognito:groups" {
					ctx.Set("groups", val)
				} else {
					ctx.Set(key, val)
				}
			}
			ctx.Set("token", tokenHeader)
			username, _ := ctx.Get("username")
			logrus.Infof("Validated token for user '%v'", username)
			ctx.Next()
		}
	}
}
