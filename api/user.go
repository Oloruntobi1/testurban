package api

import (
	"database/sql"
	"net/http"
	db "testuhpostgres/db/sqlc"
	"testuhpostgres/hash"
	"testuhpostgres/token"

	"github.com/gin-gonic/gin"
)

type createUserRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required"`
}

// type Response struct {
	
// }

type createRegistrationResponse struct {
	Token *token.TokenDetails
}

func (server *Server) createUser(ctx *gin.Context) {

	var req createUserRequest
	var res createRegistrationResponse
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, errorResponse(err))
		return
	}

	password := hash.HashAndSalt([]byte(req.Password))
	arg := db.CreateUserParams{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  password,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, "duplicate email")
		return
	}

	res.Token = token.PrepareToken(user)

	saveErr := token.CreateAuth(user, res.Token)
	if saveErr != nil {
		ctx.JSON(http.StatusUnprocessableEntity, saveErr.Error())
	}
	tokens := map[string]string{
		"access_token":  res.Token.AccessToken,
		"refresh_token": res.Token.RefreshToken,
	}
	ctx.JSON(http.StatusOK, tokens)
}

type deleteUserRequest struct {
	Email string `uri:"email" binding:"required,email"`
}

func (server *Server) deleteUser(ctx *gin.Context) {

	var req deleteUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	gottenUser, err := server.store.GetUser(ctx, req.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	err = server.store.DeleteUser(ctx, gottenUser.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, "Successfully deleted")

}

type getUserRequest struct {
	Email string `uri:"email" binding:"required,email"`
}

func (server *Server) getUser(ctx *gin.Context) {

	var req getUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	gottenUser, err := server.store.GetUser(ctx, req.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gottenUser)

}
