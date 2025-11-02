package battle

import (
	"errors"
	"fmt"

	"network-sec-micro/internal/battle/dto"
)

// ParticipantLevel represents hierarchy level for validation
type ParticipantLevel int

const (
	// Light side levels
	LevelWarrior ParticipantLevel = 1 // knight, archer, mage
	LevelLightKing ParticipantLevel = 2
	LevelLightEmperor ParticipantLevel = 3

	// Dark side levels
	LevelEnemy ParticipantLevel = 1 // goblin, pirate, skeleton
	LevelDragon ParticipantLevel = 2 // fire, ice, lightning, shadow
	LevelDarkEmperor ParticipantLevel = 3
)

// GetParticipantLevel returns the hierarchy level for a participant type
func GetParticipantLevel(pType ParticipantType) ParticipantLevel {
	switch pType {
	case ParticipantTypeWarrior:
		return LevelWarrior
	case ParticipantTypeDarkKing:
		// Dark king not allowed in battle
		return 0
	case ParticipantTypeDarkEmperor:
		return LevelDarkEmperor
	case ParticipantTypeDragon:
		return LevelDragon
	case ParticipantTypeEnemy:
		return LevelEnemy
	default:
		return 0
	}
}

// ValidateBattleParticipants validates team composition for battle
func ValidateBattleParticipants(cmd dto.StartBattleCommand) error {
	// Basic validation: Only light vs dark
	// Light side can only have: warrior, light_king, light_emperor
	// Dark side can only have: enemy, dragon, dark_emperor (NOT dark_king)
	
	// Validate light side
	for _, p := range cmd.LightParticipants {
		// Side validation
		if p.Side != "light" {
			return errors.New("light side can only contain light participants")
		}

		// Type validation
		validLightTypes := map[string]bool{
			"warrior":       true,
			"light_king":    true,
			"light_emperor": true,
		}
		if !validLightTypes[p.Type] {
			return fmt.Errorf("invalid participant type for light side: %s (allowed: warrior, light_king, light_emperor)", p.Type)
		}

		// Level validation for warriors
		if p.Type == "warrior" {
			if p.Level > 1 {
				return fmt.Errorf("warrior %s has level %d, but warriors cannot exceed level 1", p.Name, p.Level)
			}
		}
	}

	// Validate dark side
	hasDragon := false
	hasEnemy := false

	for _, p := range cmd.DarkParticipants {
		// Side validation
		if p.Side != "dark" {
			return errors.New("dark side can only contain dark participants")
		}

		// Type validation
		validDarkTypes := map[string]bool{
			"enemy":         true,
			"dragon":        true,
			"dark_emperor": true,
		}
		if !validDarkTypes[p.Type] {
			return fmt.Errorf("invalid participant type for dark side: %s (allowed: enemy, dragon, dark_emperor). Note: dark_king is NOT allowed", p.Type)
		}

		// Check for dark_king (explicitly not allowed)
		if p.Type == "dark_king" {
			return errors.New("dark_king is not allowed in battles, only dark_emperor can participate")
		}

		// Track types
		if p.Type == "dragon" {
			hasDragon = true
		} else if p.Type == "enemy" {
			hasEnemy = true
		}
	}

	// Validate: Enemy level cannot exceed dragon level
	// Check if any enemy has a higher level than any dragon
	if hasEnemy && hasDragon {
		// Get all dragon levels
		dragonLevels := make([]int, 0)
		for _, p := range cmd.DarkParticipants {
			if p.Type == "dragon" {
				if p.Level > 0 {
					dragonLevels = append(dragonLevels, p.Level)
				} else {
					// Default dragon level if not specified (dragons are typically higher level)
					dragonLevels = append(dragonLevels, 50)
				}
			}
		}

		// Check enemy levels against dragon levels
		for _, p := range cmd.DarkParticipants {
			if p.Type == "enemy" {
				enemyLevel := p.Level
				if enemyLevel == 0 {
					enemyLevel = 10 // Default enemy level
				}

				// Enemy level must be less than or equal to all dragon levels
				for _, dragonLevel := range dragonLevels {
					if enemyLevel > dragonLevel {
						return fmt.Errorf("enemy '%s' has level %d which exceeds dragon level %d - enemy level cannot exceed dragon level", 
							p.Name, enemyLevel, dragonLevel)
					}
				}
			}
		}
	}

	return nil
}

