package controllers

import (
	"log"
	"net/http"

	"todoproject/models"
	"todoproject/services"
	"todoproject/utils"

	"github.com/gin-gonic/gin"
)

var users = map[string]models.User{} // Temporary in-memory user storage

type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Captcha  string `json:"captcha" binding:"required"`
}

func Register(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing fields"})
		return
	}

	if !services.VerifyTurnstile(req.Captcha, c.ClientIP()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CAPTCHA"})
		log.Println("Registration CAPTCHA failed for user:", req.Username)
		return
	}

	if _, exists := users[req.Username]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	user := models.User{
		Username: req.Username,
		Password: hashed,
	}

	users[req.Username] = user

	log.Println("User registered:", req.Username)
	c.JSON(http.StatusOK, gin.H{"message": "Registration successful"})
}

func Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing fields"})
		return
	}

	if !services.VerifyTurnstile(req.Captcha, c.ClientIP()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CAPTCHA"})
		log.Println("Login CAPTCHA failed for user:", req.Username)
		return
	}

	user, exists := users[req.Username]
	if !exists || !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		log.Println("Login failed for user:", req.Username)
		return
	}

	token, err := services.GenerateJWT(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	log.Println("User logged in:", req.Username)
	c.JSON(http.StatusOK, gin.H{"token": token})
}
