package config

import (
	"fmt"
	"log"
	"os"

	"account-microservice/models"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	DB = database

	DB.AutoMigrate(&models.Account{}, &models.Role{}, &models.Specialization{})

	initializeAccounts()
}

func initializeAccounts() {
	defaultAccounts := []struct {
		Username string
		Password string
		Role     string
	}{
		{"admin", "admin", "Admin"},
		{"manager", "manager", "Manager"},
		{"doctor", "doctor", "Doctor"},
		{"user", "user", "User"},
	}

	for _, acc := range defaultAccounts {
		var role models.Role
		if err := DB.FirstOrCreate(&role, models.Role{Name: acc.Role}).Error; err != nil {
			log.Fatalf("Failed to create or find role %s: %v", acc.Role, err)
		}

		var account models.Account
		if err := DB.Where("username = ?", acc.Username).Preload("Roles").First(&account).Error; err == gorm.ErrRecordNotFound {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(acc.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Fatalf("Failed to hash password for account %s: %v", acc.Username, err)
			}

			newAccount := models.Account{
				Username: acc.Username,
				Password: string(hashedPassword),
				Roles:    []*models.Role{&role},
			}

			if err := DB.Create(&newAccount).Error; err != nil {
				log.Fatalf("Failed to create account %s: %v", acc.Username, err)
			}

			log.Printf("Account %s created with role %s", acc.Username, acc.Role)
		} else if err != nil {
			log.Fatalf("Error checking account %s: %v", acc.Username, err)
		} else {
			log.Printf("Account %s already exists", acc.Username)
		}
	}
}
