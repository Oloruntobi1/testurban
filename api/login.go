package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (server *Server) login(ctx *gin.Context) {

	var req loginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.Email)

	if err != nil {
		ctx.JSON(http.StatusNotFound, "email not found")
		return
	}


	passErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))

		if passErr == bcrypt.ErrMismatchedHashAndPassword && passErr != nil {
			if err == sql.ErrNoRows {
				ctx.JSON(http.StatusNotFound, "incorrect password")
				return
			}
			ctx.JSON(http.StatusNotFound, "password incorrect")
			return
		}

		ctx.JSON(http.StatusOK, user)

}