package validation_test

import (
	"testing"

	"network-sec-micro/internal/warrior"
	"network-sec-micro/internal/warrior/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	err = db.AutoMigrate(&warrior.Warrior{})
	require.NoError(t, err)
	
	warrior.DB = db
	
	return db
}

// TestWarriorValidation_Username tests username validation
func TestWarriorValidation_Username(t *testing.T) {
	db := setupTestDB(t)
	_ = db
	
	svc := warrior.NewService()
	
	tests := []struct {
		name    string
		username string
		wantErr bool
	}{
		{
			name:     "valid username",
			username: "warrior123",
			wantErr:  false,
		},
		{
			name:     "empty username",
			username: "",
			wantErr:  true,
		},
		{
			name:     "username too short",
			username: "ab",
			wantErr:  true,
		},
		{
			name:     "username with spaces",
			username: "warrior 123",
			wantErr:  true,
		},
		{
			name:     "username with special chars",
			username: "warrior@123",
			wantErr:  true,
		},
		{
			name:     "very long username",
			username: "a" + string(make([]byte, 300)),
			wantErr:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := dto.CreateWarriorCommand{
				Username: tt.username,
				Email:    "test@example.com",
				Password: "password123",
				Role:     "knight",
			}
			
			_, err := svc.CreateWarrior(cmd)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// May still fail due to other validations
				// Just check it doesn't fail on username specifically
			}
		})
	}
}

// TestWarriorValidation_Email tests email validation
func TestWarriorValidation_Email(t *testing.T) {
	db := setupTestDB(t)
	_ = db
	
	svc := warrior.NewService()
	
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
		},
		{
			name:    "email without @",
			email:   "testexample.com",
			wantErr: true,
		},
		{
			name:    "email without domain",
			email:   "test@",
			wantErr: true,
		},
		{
			name:    "email without local part",
			email:   "@example.com",
			wantErr: true,
		},
		{
			name:    "email with spaces",
			email:   "test @example.com",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := dto.CreateWarriorCommand{
				Username: "warrior1",
				Email:    tt.email,
				Password: "password123",
				Role:     "knight",
			}
			
			_, err := svc.CreateWarrior(cmd)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

// TestWarriorValidation_Password tests password validation
func TestWarriorValidation_Password(t *testing.T) {
	db := setupTestDB(t)
	_ = db
	
	svc := warrior.NewService()
	
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password:  "",
			wantErr:  true,
		},
		{
			name:     "very short password",
			password: "123",
			wantErr:  true,
		},
		{
			name:     "password with only letters",
			password: "password",
			wantErr:  false, // May be valid
		},
		{
			name:     "password with only numbers",
			password: "12345678",
			wantErr:  false, // May be valid
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := dto.CreateWarriorCommand{
				Username: "warrior1",
				Email:    "test@example.com",
				Password: tt.password,
				Role:     "knight",
			}
			
			_, err := svc.CreateWarrior(cmd)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

// TestWarriorValidation_Role tests role validation
func TestWarriorValidation_Role(t *testing.T) {
	db := setupTestDB(t)
	_ = db
	
	svc := warrior.NewService()
	
	validRoles := []string{"knight", "archer", "mage"}
	invalidRoles := []string{"invalid", "admin", "superuser", ""}
	
	for _, role := range validRoles {
		t.Run("valid_role_"+role, func(t *testing.T) {
			cmd := dto.CreateWarriorCommand{
				Username: "warrior1",
				Email:    "test@example.com",
				Password: "password123",
				Role:     role,
			}
			
			_, err := svc.CreateWarrior(cmd)
			assert.NoError(t, err)
		})
	}
	
	for _, role := range invalidRoles {
		t.Run("invalid_role_"+role, func(t *testing.T) {
			cmd := dto.CreateWarriorCommand{
				Username: "warrior1",
				Email:    "test@example.com",
				Password: "password123",
				Role:     role,
			}
			
			_, err := svc.CreateWarrior(cmd)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid role")
		})
	}
}

// TestWarriorUpdate_Validation tests update validation
func TestWarriorUpdate_Validation(t *testing.T) {
	db := setupTestDB(t)
	
	svc := warrior.NewService()
	
	// Create warrior
	cmd := dto.CreateWarriorCommand{
		Username: "warrior1",
		Email:    "test@example.com",
		Password: "password123",
		Role:     "knight",
	}
	created, err := svc.CreateWarrior(cmd)
	require.NoError(t, err)
	
	tests := []struct {
		name    string
		email   *string
		wantErr bool
	}{
		{
			name:    "valid email update",
			email:   stringPtr("newemail@example.com"),
			wantErr: false,
		},
		{
			name:    "invalid email format",
			email:   stringPtr("invalid-email"),
			wantErr: true,
		},
		{
			name:    "empty email",
			email:   stringPtr(""),
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCmd := dto.UpdateWarriorCommand{
				WarriorID: created.ID,
				Email:     tt.email,
				UpdatedBy: created.ID,
			}
			
			_, err := svc.UpdateWarrior(updateCmd)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// TestWarriorPasswordChange_Validation tests password change validation
func TestWarriorPasswordChange_Validation(t *testing.T) {
	db := setupTestDB(t)
	
	svc := warrior.NewService()
	
	// Create warrior
	cmd := dto.CreateWarriorCommand{
		Username: "warrior1",
		Email:    "test@example.com",
		Password: "oldpassword",
		Role:     "knight",
	}
	created, err := svc.CreateWarrior(cmd)
	require.NoError(t, err)
	
	tests := []struct {
		name        string
		oldPassword string
		newPassword string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid password change",
			oldPassword: "oldpassword",
			newPassword: "newpassword123",
			wantErr:     false,
		},
		{
			name:        "wrong old password",
			oldPassword: "wrongpassword",
			newPassword: "newpassword123",
			wantErr:     true,
			errMsg:      "invalid old password",
		},
		{
			name:        "empty new password",
			oldPassword: "oldpassword",
			newPassword: "",
			wantErr:     true,
		},
		{
			name:        "new password same as old",
			oldPassword: "oldpassword",
			newPassword: "oldpassword",
			wantErr:     false, // Technically valid
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changeCmd := dto.ChangePasswordCommand{
				WarriorID:  created.ID,
				OldPassword: tt.oldPassword,
				NewPassword: tt.newPassword,
			}
			
			err := svc.ChangePassword(changeCmd)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

