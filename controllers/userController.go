package controllers

import (
	"context"
	"golang-jwt/database"
	"golang-jwt/helpers"
	"golang-jwt/models"
	helper "goolsng-jwt/helpers"
	"http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/gleed007/golang-jwt/database"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollections *mongo.Collection =database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword(password string)


func VerifyPassword(password string)

func Signup(c *gin.Context) {
}

func Login(c *gin.Context) {
}

func GetUsers(c *gin.Context) {
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
    userId := c.Param("user_id")
    if err := helper.MatchUserTypeToUid(c, userId); err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		err := userCollections.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err!= nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user"})
      return
    }
		c.JSON(http.StatusOK, user)
  }
}
