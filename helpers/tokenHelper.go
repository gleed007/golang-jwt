package helper

import (
	"context"
	"fmt"
	"goolang-jwt/database"
	"log"
	"os"
	"time"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gleed007/golang-jwt/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email string
	FirstName string
	LastName string
	Uid string
	UserType string
	jwt.StandardClaims
}

var userCollections *mongo.Collection = database.OpenCollection(database.Client, "user")

var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllToken(email string, firstName string, lastName string, userType string, uId string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
    Email:       email,
    FirstName:  firstName,
    LastName:   lastName,
    Uid:        uId,
    UserType:   userType,
    StandardClaims: jwt.StandardClaims{
      ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
    },
  }

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
      ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
    },
  }

  token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
  refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err!= nil {
		log.Panic(err)
    log.Fatalf("Error while signing the token: %v", err)
  }

	return token, refreshToken, err
}
