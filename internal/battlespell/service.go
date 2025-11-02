package battlespell

import (
	"context"
	"errors"
	"fmt"

	"network-sec-micro/internal/battlespell/dto"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service handles battlespell business logic with CQRS pattern
type Service struct{}

// NewService creates a new battlespell service
func NewService() *Service {
	return &Service{}
}

// CastSpell casts a spell in a battle
func (s *Service) CastSpell(ctx context.Context, cmd dto.CastSpellCommand) (int, error) {
	// Validate spell type
	spellType := SpellType(cmd.SpellType)
	if spellType != SpellCallOfTheLightKing && spellType != SpellResistance && spellType != SpellRebirth &&
		spellType != SpellDragonEmperor && spellType != SpellDestroyTheLight && spellType != SpellWraithOfDragon {
		return 0, errors.New("invalid spell type")
	}

	// Validate caster can cast this spell
	if !spellType.CanBeCastBy(cmd.CasterRole) {
		return 0, fmt.Errorf("role %s cannot cast spell %s", cmd.CasterRole, cmd.SpellType)
	}

	battleID, err := primitive.ObjectIDFromHex(cmd.BattleID)
	if err != nil {
		return 0, errors.New("invalid battle ID")
	}

	// Route to appropriate spell handler
	switch spellType {
	case SpellCallOfTheLightKing:
		count, err := s.CastCallOfTheLightKing(ctx, battleID, cmd.CasterUsername, cmd.CasterUserID)
		return count, err

	case SpellResistance:
		count, err := s.CastResistance(ctx, battleID, cmd.CasterUsername, cmd.CasterUserID)
		return count, err

	case SpellRebirth:
		count, err := s.CastRebirth(ctx, battleID, cmd.CasterUsername, cmd.CasterUserID)
		return count, err

	case SpellDragonEmperor:
		if cmd.TargetDragonID == "" || cmd.TargetDarkEmperorID == "" {
			return 0, errors.New("dragon_participant_id and dark_emperor_participant_id are required for Dragon Emperor spell")
		}
		err := s.CastDragonEmperor(ctx, battleID, cmd.TargetDragonID, cmd.TargetDarkEmperorID, cmd.CasterUsername, cmd.CasterUserID)
		if err != nil {
			return 0, err
		}
		return 1, nil

	case SpellDestroyTheLight:
		count, err := s.CastDestroyTheLight(ctx, battleID, cmd.CasterUsername, cmd.CasterUserID)
		return count, err

	case SpellWraithOfDragon:
		err := s.CastWraithOfDragon(ctx, battleID, cmd.CasterUsername, cmd.CasterUserID)
		if err != nil {
			return 0, err
		}
		return 1, nil

	default:
		return 0, errors.New("spell not implemented")
	}
}

