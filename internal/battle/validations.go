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
	// Validate light side
	maxLightLevel := ParticipantLevel(0)
	hasLightKing := false
	hasLightEmperor := false

	for _, p := range cmd.LightParticipants {
		pType := ParticipantType(p.Type)
		
		// Check for invalid types
		if pType == ParticipantTypeEnemy || pType == ParticipantTypeDragon || 
		   pType == ParticipantTypeDarkKing || pType == ParticipantTypeDarkEmperor {
			return fmt.Errorf("invalid participant type for light side: %s", p.Type)
		}

		// Check for dark_king (not allowed)
		if p.Side != "light" {
			return errors.New("light side can only contain light participants")
		}

		level := GetParticipantLevel(pType)
		if level == 0 {
			return fmt.Errorf("invalid participant type: %s", p.Type)
		}

		// Track highest level
		if level > maxLightLevel {
			maxLightLevel = level
		}

		// Track specific types
		if pType == ParticipantTypeWarrior {
			// Warrior level check: warriors cannot exceed king level
			// This is automatic since LevelWarrior < LevelLightKing
		} else if p.Type == "light_king" {
			hasLightKing = true
		} else if p.Type == "light_emperor" {
			hasLightEmperor = true
		}
	}

	// Validate: Warriors cannot be at a level higher than king
	// Since LevelWarrior (1) < LevelLightKing (2), this is always satisfied
	// But we check explicitly: if there's a king, warriors must be level 1
	if hasLightKing || hasLightEmperor {
		for _, p := range cmd.LightParticipants {
			if p.Type == "warrior" && GetParticipantLevel(ParticipantTypeWarrior) >= LevelLightKing {
				return errors.New("warriors cannot be at king level or higher")
			}
		}
	}

	// Validate dark side
	maxDarkLevel := ParticipantLevel(0)
	hasDarkKing := false
	hasDarkEmperor := false
	hasDragon := false
	hasEnemy := false
	maxEnemyLevel := 0
	minDragonLevel := 999

	for _, p := range cmd.DarkParticipants {
		pType := ParticipantType(p.Type)
		
		// Check for invalid types for dark side
		if pType == ParticipantTypeWarrior || p.Type == "light_king" || p.Type == "light_emperor" {
			return fmt.Errorf("invalid participant type for dark side: %s", p.Type)
		}

		// Check for dark_king (not allowed in battles)
		if pType == ParticipantTypeDarkKing {
			return errors.New("dark_king is not allowed in battles, only dark_emperor can participate")
		}

		if p.Side != "dark" {
			return errors.New("dark side can only contain dark participants")
		}

		level := GetParticipantLevel(pType)
		if level == 0 {
			return fmt.Errorf("invalid participant type: %s", p.Type)
		}

		// Track highest level
		if level > maxDarkLevel {
			maxDarkLevel = level
		}

		// Track specific types
		if pType == ParticipantTypeDragon {
			hasDragon = true
			// Get dragon level from participant info (if available)
			// For now, we'll check later
		} else if pType == ParticipantTypeEnemy {
			hasEnemy = true
			// Get enemy level from participant info
			// We'll need level in participant info
		} else if pType == ParticipantTypeDarkEmperor {
			hasDarkEmperor = true
		}
	}

	// Validate: Only dark_emperor allowed (no dark_king)
	if hasDarkKing {
		return errors.New("dark_king is not allowed in battles, only dark_emperor can participate")
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
					// Default dragon level if not specified
					dragonLevels = append(dragonLevels, 50) // Default dragon level
				}
			}
		}

		// Check enemy levels
		for _, p := range cmd.DarkParticipants {
			if p.Type == "enemy" {
				enemyLevel := p.Level
				if enemyLevel == 0 {
					enemyLevel = 10 // Default enemy level
				}

				// Enemy level must be less than or equal to all dragon levels
				for _, dragonLevel := range dragonLevels {
					if enemyLevel > dragonLevel {
						return fmt.Errorf("enemy level (%d) cannot exceed dragon level (%d)", enemyLevel, dragonLevel)
					}
				}
			}
		}
	}

	return nil
}

