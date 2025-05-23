package controllers

import (
	"auth-service/db"
	"auth-service/models"
	"auth-service/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var jwtKey = []byte("15£$£%a4A124d><hab?ad34w18a@92Asd13£$")

func Login(c *gin.Context) {

	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User

	db.DB.Where("email = ?", user.Email).First(&existingUser)

	if existingUser.ID == 0 {
		c.JSON(400, gin.H{"error": "user does not exist"})
		return
	}

	errHash := utils.CompareHashPassword(user.Password, existingUser.Password)

	if !errHash {
		c.JSON(400, gin.H{"error": "invalid password"})
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)

	claims := &models.Claims{
		Role: existingUser.Role,
		StandardClaims: jwt.StandardClaims{
			Subject:   existingUser.Email,
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		c.JSON(500, gin.H{"error": "could not generate token"})
		return
	}

	c.SetCookie("token", tokenString, int(expirationTime.Unix()), "/", "localhost", false, true)
	c.JSON(200, gin.H{"success": "user logged in"})
}

func Signup(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User

	db.DB.Where("email = ?", user.Email).First(&existingUser)

	if existingUser.ID != 0 {
		c.JSON(400, gin.H{"error": "user already exists"})
		return
	}

	var errHash error
	user.Password, errHash = utils.GenerateHashPassword(user.Password)

	if errHash != nil {
		c.JSON(500, gin.H{"error": "could not generate password hash"})
		return
	}

	db.DB.Create(&user)

	c.JSON(200, gin.H{"success": "user created"})
}

func Home(c *gin.Context) {
	cookie, err := c.Cookie("token")

	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	claims, err := utils.ParseToken(cookie)

	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	if claims.Role != "user" && claims.Role != "admin" {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(200, gin.H{"success": "home page", "role": claims.Role})
}

func Premium(c *gin.Context) {
	cookie, err := c.Cookie("token")

	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	claims, err := utils.ParseToken(cookie)

	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	if claims.Role != "admin" {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(200, gin.H{"success": "premium page", "role": claims.Role})
}

func Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "localhost", false, true)
	c.JSON(200, gin.H{"success": "user logged out"})
}

func ResetPassword(c *gin.Context) {
	var payload struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var pr models.PasswordReset
	if err := db.DB.Where("token = ?", payload.Token).First(&pr).Error; err != nil {
		c.JSON(400, gin.H{"error": "invalid or expired token"})
		return
	}

	if time.Now().After(pr.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token expired"})
		return
	}

	db.DB.Delete(&pr)

	hashed, err := utils.GenerateHashPassword(payload.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash error"})
		return
	}

	db.DB.Model(&models.User{}).Where("email = ?", pr.Email).Update("password", hashed)

	c.JSON(http.StatusOK, gin.H{"message": "password updated"})
}

func RequestPasswordReset(c *gin.Context) {
	var payload struct{ Email string }
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := db.DB.Where("email = ?", payload.Email).First(&user).Error; err != nil {
		c.JSON(400, gin.H{"message": "If that email exists, you’ll get a reset link"})
		return
	}

	token, err := utils.GenerateSecureToken(16)
	if err != nil {
		c.JSON(400, gin.H{"error": "could not generate reset token"})
		return
	}

	expiresAt := time.Now().Add(5 * time.Minute)

	db.DB.Where("email = ?", payload.Email).Delete(&models.PasswordReset{})
	db.DB.Create(&models.PasswordReset{
		Email:     payload.Email,
		Token:     token,
		ExpiresAt: expiresAt,
	})

	resetURL := fmt.Sprintf("http://localhost:8080/reset-password?token=%s", token)

	if err := utils.SendResetEmail(payload.Email, resetURL); err != nil {
		c.JSON(500, gin.H{"error": "could not send email"})
		return
	}

	c.JSON(200, gin.H{"message": "If that email exists, you’ll get a reset link"})
}
