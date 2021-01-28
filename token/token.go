package token

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	db "testuhpostgres/db/sqlc"
	"testuhpostgres/hash"
	"testuhpostgres/rdstore"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/twinj/uuid"
)

type AccessDetails struct {
	AccessUuid string
	userID string

}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

func PrepareToken(user db.User) *TokenDetails {

	var err error 
	td := &TokenDetails{}

	td.AtExpires = time.Now().Add(time.Minute * 15).Unix()
	td.AccessUuid = uuid.NewV4().String()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUuid = td.AccessUuid + "++" + user.Email
	
	atContent := jwt.MapClaims{
		"user_id": user.Email,
		"expiry": td.AtExpires,
		"access_uuid": td.AccessUuid,
		"authorized": true,

	}
	aToken := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), atContent)
	td.AccessToken, err = aToken.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	hash.HandleErr(err)

		
	rtContent := jwt.MapClaims{
		"user_id": user.Email,
		"expiry": td.RtExpires,
		"refresh_uuid": td.RefreshUuid,
		"authorized": true,

	}
	rToken := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), rtContent)
	td.RefreshToken, err = rToken.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	hash.HandleErr(err)

	return td
}


func CreateAuth(user db.User, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	errAccess := rdstore.Client.Set(td.AccessUuid, user.Email, at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}
	errRefresh := rdstore.Client.Set(td.RefreshUuid,  user.Email, rt.Sub(now)).Err()
	if errRefresh != nil {
		return errRefresh
	}
	return nil
}

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
	   err := TokenValid(c.Request)
	   if err != nil {
		  c.JSON(http.StatusUnauthorized, err.Error())
		  c.Abort()
		  return
	   }
	   c.Next()
	}
  }

func TokenValid(r *http.Request) error {
	token, err := VerifyToken(r)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok || !token.Valid {
		return err
	}
	return nil
}

func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func ExtractTokenMetadata(r *http.Request) (*AccessDetails, error) {
	token, err := VerifyToken(r)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}
		// userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)

		userID := fmt.Sprintf("%v", claims["user_id"])
		if err != nil {
			return nil, err
		}
		return &AccessDetails{
			AccessUuid: accessUuid,
			userID:   userID,
		}, nil
	}
	return nil, err
}


func FetchAuth(authD *AccessDetails) (string, error) {
	userid, err := rdstore.Client.Get(authD.AccessUuid).Result()
	if err != nil {
		return "", err
	}

	userID := fmt.Sprintf("%v", userid)
	// userID, _ := strconv.ParseInt(userid, 10, 64)
	if authD.userID != userID {
		return "", errors.New("unauthorized")
	}
	return userID, nil
}

func  DeleteTokens(authD *AccessDetails) error {
	//get the refresh uuid
	refreshUuid := fmt.Sprintf("%s++%s", authD.AccessUuid, authD.userID)
	//delete access token
	deletedAt, err := rdstore.Client.Del(authD.AccessUuid).Result()
	if err != nil {
		return err
	}
	//delete refresh token
	deletedRt, err := rdstore.Client.Del(refreshUuid).Result()
	if err != nil {
		return err
	}
	//When the record is deleted, the return value is 1
	if deletedAt != 1 || deletedRt != 1 {
		return errors.New("something went wrong")
	}
	return nil
}

func DeleteAuth(givenUuid string) (int64,error) {
	deleted, err := rdstore.Client.Del(givenUuid).Result()
	if err != nil {
		return 0, err
	}
	return deleted, nil
}