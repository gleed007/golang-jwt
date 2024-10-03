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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollections *mongo.Collection =database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword(password string)


func VerifyPassword(password string)

func Signup() gin.HandlerFunc {
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

		validationErr := validate.Struct(user)
		if validationErr!= nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
      return
    }

		count, err := userCollections.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()
		if err!= nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting users"})
      return
    }

		count, err = userCollections.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err!= nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting users"})
      return
    }

		if count > 0 {
      c.JSON(http.StatusBadRequest, gin.H{"error": "Email or phone already exists"})
      return
    }

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.UserID = user.ID.Hex()
		Token, RefreshToken, _ := helpers.GenerateAllToken(*user.Email, *user.FirstName, *user.LastName, *user.UserType, *&user.UserID)
		user.Token = &Token
		user.RefreshToken = &RefreshToken

		resultInsertionNumber, inserterr := userCollections.InsertOne(ctx, user)
		if inserterr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting user"})
      return
		}
		defer cancel()
		c.JSON(http.StatusCreated, resultInsertionNumber)
	}
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
