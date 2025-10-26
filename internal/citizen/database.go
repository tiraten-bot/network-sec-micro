package citizen

import (
	"errors"
	"log"

	"network-sec-micro/internal/citizen/dto"
	"network-sec-micro/pkg/auth"
	"network-sec-micro/pkg/database"

	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDatabase initializes the database for citizen service
func InitDatabase() error {
	config := database.GetConfigFromEnv()
	
	var err error
	DB, err = database.Connect(config)
	if err != nil {
		return err
	}

	// Auto migrate
	if err := database.AutoMigrate(DB, &Citizen{}); err != nil {
		return err
	}

	// No seed data for citizens - they register themselves
	log.Println("Citizen database initialized")
	return nil
}

// Register creates a new citizen
func Register(req dto.RegisterRequest) (*Citizen, error) {
	// Check if username already exists
	var existing Citizen
	if err := DB.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		return nil, errors.New("username already exists")
	}

	// Check if email already exists
	if err := DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create citizen
	citizen := Citizen{
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if err := DB.Create(&citizen).Error; err != nil {
		return nil, err
	}

	// Remove password
	citizen.Password = ""
	return &citizen, nil
}
