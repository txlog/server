package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/txlog/server/database"
)

func main() {
	r := gin.Default()
	database.ConnectDatabase()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/add", addUser)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

type User struct {
	Username string
	Password string
}

func addUser(ctx *gin.Context) {
	body := User{}
	data, err := ctx.GetRawData()
	if err != nil {
		ctx.AbortWithStatusJSON(400, "User is not defined")
		return
	}
	err = json.Unmarshal(data, &body)
	if err != nil {
		ctx.AbortWithStatusJSON(400, "Bad Input")
		return
	}
	//use Exec whenever we want to insert update or delete
	//Doing Exec(query) will not use a prepared statement, so lesser TCP calls to the SQL server
	_, err = database.Db.Exec("insert into users(username,password) values ($1,$2)", body.Username, body.Password)
	if err != nil {
		fmt.Println(err)
		ctx.AbortWithStatusJSON(400, "Couldn't create the new user.")
	} else {
		ctx.JSON(http.StatusOK, "User is successfully created.")
	}

}
