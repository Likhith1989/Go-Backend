package AuthController

import (
	"context"
	databases "example/go-backend/database"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var jwtKey = []byte("secret69")

type User struct {
	Username string `json:"usern"`
	Password string `json:"passw"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func UserExists(client *mongo.Client, dbName string, collectionName string, username string) (bool, error) {
	filter := bson.D{{Key: "username", Value: username}}
	coll := client.Database(dbName).Collection(collectionName)

	var results bson.M

	err := coll.FindOne(context.TODO(), filter).Decode(&results)
	if err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil

}

func InsertUser(client *mongo.Client, dbName string, collectionName string, user User) (bool, error) {
	coll := client.Database(dbName).Collection(collectionName)

	_, err := coll.InsertOne(context.TODO(), user)

	if err != nil {
		return false, err
	}
	return true, err
}

func CheckPassword(client *mongo.Client, dbName, collectionName string, user_ User) (bool, error) {
	collection := client.Database(dbName).Collection(collectionName)
	var user bson.M

	// Find the user by username
	err := collection.FindOne(context.TODO(), bson.M{"username": user_.Username}).Decode(&user)
	if err != nil {
		return false, err
	}

	// Retrieve the stored plaintext password
	storedPassword := user["password"].(string)

	// Compare the stored password with the one provided by the user
	if storedPassword != user_.Password {
		return false, nil
	}

	return true, nil
}

func Signup(c *gin.Context) {
	Client := databases.ConnectDB()
	defer func() {
		if err := Client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	var reqData User
	if err := c.BindJSON(&reqData); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	fmt.Printf("\n username is %v \n", reqData.Username)

	exists, err1 := UserExists(Client, "movieapp", "users", reqData.Username)

	var response bson.M = bson.M{}

	if err1 != nil {
		panic(err1)
	}
	if exists {
		response = bson.M{"result": "TryAgain"}
		c.JSON(200, response)
		return
	}

	inserted, err := InsertUser(Client, "movieapp", "users", reqData)

	if err != nil {
		panic(err)
	}
	if !inserted {
		response = bson.M{"result": "TryAgain"}
		c.JSON(200, response)
		return
	}

	response = bson.M{
		"result": "Success",
	}
	c.JSON(200, response)

}

func SignIn(c *gin.Context) {
	// Connect to MongoDB
	Client := databases.ConnectDB()
	defer func() {
		if err := Client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	// Bind the incoming JSON to credentials struct
	var creds User
	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Check if user exists and verify the password (plain text check)
	userExists, err := UserExists(Client, "movieapp", "users", creds.Username)

	var response bson.M = bson.M{}

	if err != nil {
		panic(err)
	}
	if !userExists {
		response = bson.M{"result": "NoUser"}
		c.JSON(200, response)
		return
	}

	// Verify the plaintext password
	correctPassword, err := CheckPassword(Client, "movieapp", "users", creds)
	if err != nil || !correctPassword {
		response = bson.M{"result": "NoUser"}
		c.JSON(200, response)
		return
	}

	// Generate JWT token if login is successful
	expirationTime := time.Now().Add(24 * time.Hour) // Token expiry set to 24 hours
	claims := &Claims{
		Username: creds.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		response = bson.M{"result": "NoUser"}
		c.JSON(200, response)
		return
	}
	response = bson.M{"token": tokenString}
	// Return the JWT token in the response
	c.JSON(200, response)

}
