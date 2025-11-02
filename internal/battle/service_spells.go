package battle

import (
	"context"
	"errors"
	"fmt"

	"network-sec-micro/internal/battle/dto"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CastSpellCommand represents a command to cast a spell (internal use, matches dto.CastSpellCommand)
type CastSpellCommand = dto.CastSpellCommand

// CastSpell casts a spell in a battle
func (s *Service) CastSpell(ctx context.Context, cmd dto.CastSpellCommand) error {
	// Validate spell type
	spellType := SpellType(cmd.SpellType)
	if spellType != SpellCallOfTheLightKing && spellType != SpellResistance && spellType != SpellRebirth &&
		spellType != SpellDragonEmperor && spellType != SpellDestroyTheLight && spellType != SpellWraithOfDragon {
		return errors.New("invalid spell type")
	}

	// Validate caster can cast this spell
	if !spellType.CanBeCastBy(cmd.CasterRole) {
		return fmt.Errorf("role %s cannot cast spell %s", cmd.CasterRole, cmd.SpellType)
	}

	battleID, err := primitive.ObjectIDFromHex(cmd.BattleID)
	if err != nil {
		return errors.New("invalid battle ID")
	}

	// Route to appropriate spell handler
	switch spellType {
	case SpellCallOfTheLightKing:
		return s.CastCallOfTheLightKing(ctx, battleID, cmd.CasterUsername, cmd.CasterUserID)
	
	case SpellResistance:
		return s.CastResistance(ctx, battleID, cmd.CasterUsername, cmd.CasterUserID)
	
	case SpellRebirth:
		_, err := s.CastRebirth(ctx, battleID, cmd.CasterUsername, cmd.CasterUserID)
		return err
	
	case SpellDragonEmperor:
		if cmd.TargetDragonID == "" || cmd.TargetDarkEmperorID == "" {
			return errors.New("dragon_participant_id and dark_emperor_participant_id are required for Dragon Emperor spell")
		}
		return s.CastDragonEmperor(ctx, battleID, cmd.TargetDragonID, cmd.TargetDarkEmperorID, cmd.CasterUsername, cmd.CasterUserID)
	
	case SpellDestroyTheLight:
		return s.CastDestroyTheLight(ctx, battleID, cmd.CasterUsername, cmd.CasterUserID)
	
	case SpellWraithOfDragon:
		return s.CastWraithOfDragon(ctx, battleID, cmd.CasterUsername, cmd.CasterUserID)
	
	default:
		return errors.New("spell not implemented")
	}
}

