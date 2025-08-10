package services

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
)

func VerifyTurnstile(token, remoteip string) bool {
	secret := "0x4AAAAAABqNPZGix-kOFOKHjLeF0GkwPo"
	endpoint := "https://challenges.cloudflare.com/turnstile/v0/siteverify"

	data := url.Values{}
	data.Set("secret", secret)
	data.Set("response", token)
	data.Set("remoteip", remoteip)

	resp, err := http.PostForm(endpoint, data)
	if err != nil {
		log.Println("Error verifying Turnstile CAPTCHA:", err)
		return false
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Println("Error decoding Turnstile response:", err)
		return false
	}

	return result.Success
}
