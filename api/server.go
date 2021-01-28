package api

import (

	db "testuhpostgres/db/sqlc"
	"github.com/gin-gonic/gin"
)

// Server struct to hold our store and the gin router
type Server struct {
	store *db.Store
	router *gin.Engine
}

//NewServer contains our server routes 
func NewServer(store *db.Store) *Server {
	server := &Server{store : store}
	r := gin.Default()



	// r.Run(":29090")

	
	r.POST("/register", server.createUser)
	r.POST("/login", server.login)
	r.GET("/user/:email", server.getUser)

	
	server.router = r
	return server
}

// Start helps start a new server
func (server *Server) Start (address string) error {

	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}