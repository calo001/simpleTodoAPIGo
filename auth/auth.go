package auth

import (
	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"time"
	"todoAPI/config"
	"todoAPI/model"
)

func SetupAuth() (*jwt.GinJWTMiddleware, error){
	/*
	 * The JWT middleware authentication
	 */
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       	"	apitodogo", // https://tools.ietf.org/html/rfc7235#section-2.2
		Key:         		[]byte(config.Key),
		Timeout:     		time.Hour*24,
		MaxRefresh:  		time.Hour,
		IdentityKey: 		config.IdentityKey,
		PayloadFunc: 		payload,
		IdentityHandler: 	identityHandler,
		Authenticator: 		authenticator,
		Authorizator: 		authorizator,
		Unauthorized: 		unauthorized,
		LoginResponse: 		loginResponse,
		TokenLookup:   		"header: Authorization, query: token, cookie: jwt",
		TokenHeadName: 		"Bearer",
		TimeFunc:      		time.Now,
	})

	return authMiddleware, err
}

func payload (data interface{}) jwt.MapClaims {
	if v, ok := data.(*model.User); ok {
		return jwt.MapClaims{
			config.IdentityKey: v.ID,
		}
	}
	return jwt.MapClaims{}
}

func identityHandler (c *gin.Context) interface{} {
	claims := jwt.ExtractClaims(c)
	var user model.User
	config.GetDB().Where("id = ?", claims[config.IdentityKey]).First(&user)

	return user
}

func authenticator (c *gin.Context) (interface{}, error) {
	var loginVals model.User
	if err := c.ShouldBind(&loginVals); err != nil {
		return "", jwt.ErrMissingLoginValues
	}

	var result model.User
	config.GetDB().Where("username = ? AND password = ?", loginVals.Username, loginVals.Password).First(&result)

	if result.ID == 0 {
		return nil, jwt.ErrFailedAuthentication
	}

	return &result, nil
}

func authorizator (data interface{}, c *gin.Context) bool {
	if v, ok := data.(model.User); ok && v.ID != 0 {
		return true
	}

	return false
}

func unauthorized (c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"message": message,
	})
}

func loginResponse (c *gin.Context, code int, token string, expire time.Time) {
	c.JSON(code, gin.H{
		"expire": expire,
		"token":  token,
	})
}