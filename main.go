package main

import (
	"github.com/appleboy/gin-jwt"
	jwt2 "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"net/http"
	"os"
	"time"
)

var db *gorm.DB
var identityKey = "id"
var key = "my_secret_key_8F6E2P"

func init() {
	var err error
	//db, err = gorm.Open("postgres", "host=localhost port=5432 user=admin dbname=tododb password=123  sslmode=disable")
	db, err = gorm.Open("postgres", "host=postgres.render.com port=5432 user=admin dbname=tododb password=tQGGa3UsRV")

	if err != nil {
		panic(err.Error())
	}

	db.AutoMigrate(&Todo{})
	db.AutoMigrate(&User{})
}

func main() {
	router := gin.Default()

	/*
	 * The JWT middleware authentication
	 */
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "apitodogo", // https://tools.ietf.org/html/rfc7235#section-2.2
		Key:         []byte(key),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*User); ok {
				return jwt.MapClaims{
					identityKey: v.ID,
				}
			}
			return jwt.MapClaims{} // Retorna tu token al hacer login
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			var user User
			db.Where("id = ?", claims[identityKey]).First(&user)

			return user
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals User
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			var result User
			db.Where("username = ? AND password = ?", loginVals.Username, loginVals.Password).First(&result)

			if result.ID == 0 {
				return nil, jwt.ErrFailedAuthentication // Checa en el repositorio que el usuario exista
			}

			return &result, nil
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			//if v, ok := data.(*User); ok    where v is a User
			if v, ok := data.(User); ok && v.ID != 0 {
				return true
			}

			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			}) // En caso de que no sea un usuario valido
		},
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			c.JSON(code, gin.H{
				"expire": expire,
				"token":  token,
			})
		},
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	})

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	/*
	 * Routes
	 */
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to my Todo App")
	})

	v1 := router.Group("/v1")
	{
		//v1.POST("/login", loginEndPoint)
		v1.POST("/login", authMiddleware.LoginHandler)

		v1.POST("/register", registerEndPoint)

		todo := v1.Group("todo")
		{
			todo.POST("/create", authMiddleware.MiddlewareFunc(), createTodo)
			todo.GET("/all", authMiddleware.MiddlewareFunc(), fetchAllTodo)
			todo.GET("/get/:id", authMiddleware.MiddlewareFunc(), fetchSingleTodo)
			todo.PUT("/update/:id", authMiddleware.MiddlewareFunc(), updateTodo)
			todo.DELETE("/delete/:id", authMiddleware.MiddlewareFunc(), deleteTodo)
		}
	}

	router.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Printf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	auth := router.Group("/auth")
	// Refresh time can be longer than token timeout
	auth.GET("/refresh_token", authMiddleware.RefreshHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		log.Panicf("error: %s", err)
	}
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

func registerEndPoint(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.Save(&user)

	token := jwt2.NewWithClaims(jwt2.SigningMethodHS256, jwt2.MapClaims{
		identityKey: user.ID,
	})

	tokenString, err := token.SignedString([]byte(key))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully!", "token": tokenString})
}

func createTodo(c *gin.Context) {
	claims := jwt.ExtractClaims(c)

	var user User
	db.Where("id = ?", claims[identityKey]).First(&user)

	if user.ID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}

	var todo Todo
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	todo.UserID = user.ID
	db.Save(&todo)
	c.JSON(http.StatusCreated, gin.H{"message": "Todo created successfully!", "todoId": todo.ID})
}

func fetchAllTodo(c *gin.Context) {
	claims := jwt.ExtractClaims(c)

	var user User
	db.Where("id = ?", claims[identityKey]).First(&user)

	if user.ID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}

	var todos []Todo
	db.Where("user_id = ?", user.ID).Find(&todos)

	if len(todos) <= 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No todos found!", "data": todos})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": todos})
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
		c.JSON(http.StatusNotFound, gin.H{"message": "No todo found!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": todo})
}

func updateTodo(c *gin.Context) {
	todoID := c.Param("id")

	// Check id from parameter is correct
	if len(todoID) <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}

	// Create the todo with values updated
	var newTodo Todo
	if err := c.ShouldBindJSON(&newTodo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Search the original todo
	var todo Todo
	db.First(&todo, todoID)

	// Check if the Todo exist
	if todo.ID <= 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No todo found!"})
		return
	}

	// Update the original todo with the values from newTodo
	db.Model(&todo).Update("title", newTodo.Title)
	db.Model(&todo).Update("description", newTodo.Description)
	c.JSON(http.StatusOK, gin.H{"message": "Todo updated successfully!"})
}

func deleteTodo(c *gin.Context) {
	var todo Todo
	todoID := c.Param("id")

	if len(todoID) <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}

	db.First(&todo, todoID)

	if todo.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No todo found!"})
		return
	}

	db.Delete(&todo)
	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully!"})
}
