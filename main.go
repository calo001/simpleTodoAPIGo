package main

import "C"
import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"net/http"
	"strconv"
)

var db *gorm.DB

func init() {
	var err error
	db, err = gorm.Open("postgres", "host=localhost port=5432 user=admin dbname=tododb password=123  sslmode=disable")

	if err != nil {
		panic(err.Error())
	}

	db.AutoMigrate(&Todo{})
	db.AutoMigrate(&User{})
}

func main() {
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to my Todo App")
	})

	v1 := router.Group("/v1")
	{
		v1.POST("/login", loginEndPoint)

		v1.POST("/register", registerEndPoint)

		todo := v1.Group("todo")
		{
			todo.POST("/create/:id", createTodo)
			todo.GET("/all/:id", fetchAllTodo)
			todo.GET("/get/:id", fetchSingleTodo)
			todo.PUT("/update/:id", updateTodo)
			todo.DELETE("/delete/:id", deleteTodo)
		}
	}

	router.Run(":5000")
}

type User struct {
	gorm.Model
	Username string `json:"username"`
	Password string `json:"password"`
	Todos    []Todo `json:"todos"`
}

type Todo struct {
	gorm.Model
	Title       string `json:"title"`
	Description string `json:"description"`
	UserID      uint   `json:"userid"`
}

func loginEndPoint(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var result User
	db.Where("username = ? AND password = ?", user.Username, user.Password).First(&result)

	if result.ID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "message": "User or password incorrect"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Authentication correct!"})
}

func registerEndPoint(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.Save(&user)
	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "User created successfully!", "userId": user.ID})
}

func createTodo(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))

	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}

	var todo Todo
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	todo.UserID = uint(userID)
	db.Save(&todo)
	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "Todo created successfully!", "todoId": todo.ID})
}

func fetchAllTodo(c *gin.Context) {
	userID := c.Param("id")

	if len(userID) <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}

	var todos []Todo
	db.Where("user_id = ?", userID).Find(&todos)

	if len(todos) <= 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No todos found!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusCreated, "data": todos})
}

func fetchSingleTodo(c *gin.Context) {
	todoID := c.Param("id")

	if len(todoID) <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}

	var todo Todo
	db.First(&todo, todoID)

	if todo.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No todo found!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": todo})
}

func updateTodo(c *gin.Context) {
	var todo Todo
	todoID := c.Param("id")

	db.First(&todo, todoID)

	if todo.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No todo found!"})
		return
	}

	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.Model(&todo).Update("title", todo.Title)
	db.Model(&todo).Update("description", todo.Description)
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Todo updated successfully!"})
}

func deleteTodo(c *gin.Context) {
	var todo Todo
	todoID := c.Param("id")

	db.First(&todo, todoID)

	if todo.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No todo found!"})
		return
	}

	db.Delete(&todo)
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Todo deleted successfully!"})
}
