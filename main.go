package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

// ================== MODELS ==================

type User struct {
	Username string `json:"username"`
	Password string `json:"password"` // bcrypt hash
}

type Todo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Task     string `json:"task"`
	Done     bool   `json:"done"`
}

type turnstileVerifyResp struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

// ================== GLOBALS ==================

var (
	usersFile        = "users.json"
	todosFile        = "todos.json"
	TURNSTILE_SECRET = "init() from env"
	logFilePath      = "project.log"
)

// ================== INIT ==================

func init() {
	// logging to file
	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Could not open log file, logging to stdout")
	} else {
		log.SetOutput(f)
	}
	// load secret from env
	TURNSTILE_SECRET = os.Getenv("TURNSTILE_SECRET")
	log.Println("init init TURNSTILE_SECRET TURNSTILE_SECRET:", TURNSTILE_SECRET)

	if TURNSTILE_SECRET == "" {
		log.Println("WARNING: TURNSTILE_SECRET not set; Turnstile will fail verification.")
	}

	err = godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	TURNSTILE_SECRET = os.Getenv("TURNSTILE_SECRET")
	log.Println("init init2 TURNSTILE_SECRET TURNSTILE_SECRET:", TURNSTILE_SECRET)

}

// ================== STORAGE HELPERS ==================

func loadUsers() ([]User, error) {
	data, err := os.ReadFile(usersFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []User{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return []User{}, nil
	}
	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func saveUsers(users []User) error {
	b, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(usersFile, b, 0644)
}

func loadTodos() ([]Todo, error) {
	data, err := os.ReadFile(todosFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Todo{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return []Todo{}, nil
	}
	var todos []Todo
	if err := json.Unmarshal(data, &todos); err != nil {
		return nil, err
	}
	return todos, nil
}

func saveTodos(todos []Todo) error {
	b, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(todosFile, b, 0644)
}

// ================== CAPTCHA VERIFY ==================

func verifyTurnstile(token string) bool {
	if token == "" || TURNSTILE_SECRET == "" {
		return false
	}
	log.Println("verifyTurnstileverifyTurnstileverifyTurnstile:", TURNSTILE_SECRET)

	resp, err := http.PostForm(
		"https://challenges.cloudflare.com/turnstile/v0/siteverify",
		url.Values{
			"secret":   {TURNSTILE_SECRET},
			"response": {token},
		},
	)
	if err != nil {
		log.Println("Turnstile verify request error:", err)
		return false
	}
	defer resp.Body.Close()

	var ts turnstileVerifyResp
	if err := json.NewDecoder(resp.Body).Decode(&ts); err != nil {
		log.Println("Turnstile decode error:", err)
		return false
	}
	if !ts.Success {
		log.Printf("Turnstile failed: %+v\n", ts.ErrorCodes)
	}
	return ts.Success
}

// ================== MAIN ==================

func main() {
	log.Println("Server starting...")
	r := gin.Default()

	// static files
	r.Static("/static", "./static")

	// load all templates from /templates folder
	r.LoadHTMLGlob("templates/*")

	// --- LOGIN PAGE ---
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	// --- REGISTER PAGE ---
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})

	// --- HOME PAGE ---
	r.GET("/home", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", nil)
	})

	// ============= CRUD APIs =============

	// List todos (for a user)
	r.GET("/todos/:username", func(c *gin.Context) {
		username := c.Param("username")
		todos, err := loadTodos()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
			return
		}
		userTodos := []Todo{}
		for _, t := range todos {
			if t.Username == username {
				userTodos = append(userTodos, t)
			}
		}
		c.JSON(http.StatusOK, userTodos)
	})

	// Create todo
	r.POST("/todos/:username", func(c *gin.Context) {
		username := c.Param("username")
		task := c.PostForm("task")

		todos, _ := loadTodos()
		newID := 1
		if len(todos) > 0 {
			newID = todos[len(todos)-1].ID + 1
		}
		newTodo := Todo{ID: newID, Username: username, Task: task, Done: false}
		todos = append(todos, newTodo)
		saveTodos(todos)
		c.JSON(http.StatusOK, newTodo)
	})

	// Update todo
	r.PUT("/todos/:username/:id", func(c *gin.Context) {
		username := c.Param("username")
		idStr := c.Param("id")
		id, _ := strconv.Atoi(idStr)

		todos, _ := loadTodos()
		updated := false
		for i, t := range todos {
			if t.ID == id && t.Username == username {
				todos[i].Done = !t.Done // toggle done
				updated = true
				break
			}
		}
		if updated {
			saveTodos(todos)
			c.JSON(http.StatusOK, gin.H{"status": "updated"})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		}
	})

	// Delete todo
	r.DELETE("/todos/:username/:id", func(c *gin.Context) {
		username := c.Param("username")
		idStr := c.Param("id")
		id, _ := strconv.Atoi(idStr)

		todos, _ := loadTodos()
		newTodos := []Todo{}
		found := false
		for _, t := range todos {
			if !(t.ID == id && t.Username == username) {
				newTodos = append(newTodos, t)
			} else {
				found = true
			}
		}
		if found {
			saveTodos(newTodos)
			c.JSON(http.StatusOK, gin.H{"status": "deleted"})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		}
	})

	// ============= Registration =============
	r.POST("/register", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		token := c.PostForm("cf-turnstile-response")

		if !verifyTurnstile(token) {
			c.String(http.StatusBadRequest, "Captcha verification failed")
			return
		}

		users, err := loadUsers()
		if err != nil {
			c.String(http.StatusInternalServerError, "Server error")
			return
		}
		for _, u := range users {
			if u.Username == username {
				c.String(http.StatusBadRequest, "Username already exists")
				return
			}
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		users = append(users, User{Username: username, Password: string(hash)})
		saveUsers(users)
		c.Redirect(http.StatusSeeOther, "/")
	})

	// ============= Login =============
	r.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		token := c.PostForm("cf-turnstile-response")
		log.Println("username , password, toke", username, password, token)

		if !verifyTurnstile(token) {
			c.String(http.StatusUnauthorized, "Captcha verification failed")
			return

		}
		users, _ := loadUsers()
		for _, u := range users {
			if u.Username == username {
				if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil {
					c.Redirect(http.StatusSeeOther, "/home")
					return
				}
			}
		}
		c.String(http.StatusUnauthorized, "Invalid username or password")
	})

	r.Run(":8080")
}
