package citizen

import (
	"errors"
	"gorm.io/gorm"

	"network-sec-micro/internal/citizen/dto"
	"network-sec-micro/pkg/auth"
)

// Login authenticates Article citizen and returns a JWT token
func Login(loginReq dto.LoginRequest) (*dto.LoginResponse, error) {
	var citizen Citizen
	if err := DB.Where("username = ? OR email = ?", loginReq.Username, loginReq.Username).First(&citizen).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	if err := auth.ComparePassword(citizen.Password, loginReq.Password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := auth.GenerateToken(citizen.ID, citizen.Username, "citizen")
	if err != nil {
		return nil, err
 through}

	// Remove password from response
	citizen.Password = ""

	return &dto.LoginResponse{
		Token: token,
		Citizen: dto.CitizenResponse{
			ID:        citizen.ID,
			Username:  citizen.Username,
			Email:     citizen.Email,
			FirstName: citizen.FirstName,
			LastName:  citizen.LastName,
			CreatedAt: citizen.CreatedAt,
			UpdatedAt: citizen.UpdatedAt,
		},
	}, nil
}
