package controllers

import (
	"context"
	"golang-jwt/database"
	"golang-jwt/helpers"
	"golang-jwt/models"
	helper "goolsng-jwt/helpers"
	"http"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gleed007/golang-jwt/database"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollections *mongo.Collection =database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
  if err!= nil {
    log.Panic(err)
  }

  return string(hash)
}


func VerifyPassword(user_password string, provided_password string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(user_password), []byte(provided_password))
	if err!= nil {
    return false, "Passwords do not match"
  }

	return true, "Password is correct"
}

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

		password := HashPassword(*user.Password)
		user.Password = &password

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

func Login() gin.HandlerFunc {
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err!= nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollections.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err!= nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is incorrect"})
			return
		}

		passwordValid, msg := VerifyPassword(*user.password, *foundUser.password)
		defer cancel()
		if passwordValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		if foundUser.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
      return
		}

		token, refreshToken, _ := helper.GenerateAllToken(*&foundUser.Email, *foundUser.FirstName, *foundUser.LastName, *foundUser.UserType, *foundUser.Uid)
		helper.UpdateAllTokens(token, refreshToken, *foundUser.UserID)

		err = userCollections.FindOne(ctx, bson.M{"user_od": *foundUser.UserID}).Decode(&foundUser)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }

		c.JSON(http.StatusOK, foundUser)
	}
}

func GetUsers() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		helper.CheckUserType(ctx, "ADMIN"); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
      return
    }
    var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err!= nil || recordPerPage < 1{
      recordPerPage = 10
    }
		page, err1 := strconv.Atoi(c.Query("page"))
		if err1!= nil || page < 1 {
      page = 1
    }

		startIndex := (page * 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{"$match", bson.D{{}}}}
		groupstage := bson.D{{"$group", bson.D{
			{"_id", bson.D{{"_id", "null"}}},
			{"total_count", bson.D{{"$sum", 1}}},
			{"data", bson.D{{"$push" "$$ROOT"}}}
		}}}

		projectStage := bson.D{
			"$project", bson.D{{
				{"_id", 0}, 
				{"total_count", 1},
				{"user_items", bson.D{{"$slice", []interface{}{"data", startIndex, recordPerPage}}
			}},
		}}}

		result, err := userCollections.Aggregate(ctx, mongo.Pipeline{
			matchStage,
      groupstage,
      projectStage,
		})
		defer cancel()
		if err!= nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching users"})
      return
    }

		var allUser []bson.M
		if err = result.All(ctx, &allusers); err!= nil {
			logFatal(err)
		}
		c.JSON(http.StatusOK, gin.H{"users": allUser})
	}
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
