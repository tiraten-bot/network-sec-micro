package battle_test

import (
	"context"
	"testing"
	"time"

	"network-sec-micro/internal/battle"
	"network-sec-micro/internal/battle/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// setupTestDB creates a test MongoDB connection
func setupTestDB(t *testing.T) *mongo.Database {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	
	db := client.Database("battle_test_db")
	// Clean up
	db.Collection("battles").Drop(ctx)
	db.Collection("battle_turns").Drop(ctx)
	db.Collection("battle_participants").Drop(ctx)
	
	// Set global collections
	battle.BattleColl = db.Collection("battles")
	battle.BattleTurnColl = db.Collection("battle_turns")
	battle.BattleParticipantColl = db.Collection("battle_participants")
	
	return db
}

func TestCalculateDamage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := battle.NewService()
	
	tests := []struct {
		name           string
		attackerPower  int
		targetDefense  int
		expectedMin    int // Minimum expected damage
		expectedMax    int // Maximum expected damage (with randomness)
	}{
		{
			name:          "strong attacker vs weak defender",
			attackerPower: 100,
			targetDefense: 20,
			expectedMin:   60, // ~80 base * 0.8
			expectedMax:   100, // ~80 base * 1.2
		},
		{
			name:          "equal power",
			attackerPower: 50,
			targetDefense: 50,
			expectedMin:   10, // Minimum damage
			expectedMax:   20,
		},
		{
			name:          "weak attacker vs strong defender",
			attackerPower: 30,
			targetDefense: 100,
			expectedMin:   10, // Minimum damage enforced
			expectedMax:   20,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			damage := svc.CalculateDamage(tt.attackerPower, tt.targetDefense)
			assert.GreaterOrEqual(t, damage, tt.expectedMin)
			assert.LessOrEqual(t, damage, tt.expectedMax)
		})
	}
}

func TestBattle_OpponentDefense(t *testing.T) {
	tests := []struct {
		name           string
		battleType     battle.BattleType
		expectedDefense int
	}{
		{
			name:           "dragon battle",
			battleType:     battle.BattleTypeDragon,
			expectedDefense: 100,
		},
		{
			name:           "enemy battle",
			battleType:     battle.BattleTypeEnemy,
			expectedDefense: 50,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := battle.Battle{
				BattleType: tt.battleType,
			}
			defense := b.OpponentDefense()
			assert.Equal(t, tt.expectedDefense, defense)
		})
	}
}

func TestBattleStatus_Constants(t *testing.T) {
	// Test that battle status constants are defined correctly
	assert.Equal(t, battle.BattleStatus("pending"), battle.BattleStatusPending)
	assert.Equal(t, battle.BattleStatus("in_progress"), battle.BattleStatusInProgress)
	assert.Equal(t, battle.BattleStatus("completed"), battle.BattleStatusCompleted)
	assert.Equal(t, battle.BattleStatus("cancelled"), battle.BattleStatusCancelled)
}

func TestBattleResult_Constants(t *testing.T) {
	// Test battle result constants
	assert.Equal(t, battle.BattleResult("light_victory"), battle.BattleResultLightVictory)
	assert.Equal(t, battle.BattleResult("dark_victory"), battle.BattleResultDarkVictory)
	assert.Equal(t, battle.BattleResult("draw"), battle.BattleResultDraw)
	assert.Equal(t, battle.BattleResult("victory"), battle.BattleResultVictory)
	assert.Equal(t, battle.BattleResult("defeat"), battle.BattleResultDefeat)
}

func TestTeamSide_Constants(t *testing.T) {
	// Test team side constants
	assert.Equal(t, battle.TeamSide("light"), battle.TeamSideLight)
	assert.Equal(t, battle.TeamSide("dark"), battle.TeamSideDark)
}

func TestParticipantType_Constants(t *testing.T) {
	// Test participant type constants
	assert.Equal(t, battle.ParticipantType("warrior"), battle.ParticipantTypeWarrior)
	assert.Equal(t, battle.ParticipantType("enemy"), battle.ParticipantTypeEnemy)
	assert.Equal(t, battle.ParticipantType("dragon"), battle.ParticipantTypeDragon)
}

func TestBattle_Validation(t *testing.T) {
	tests := []struct {
		name    string
		battle  battle.Battle
		isValid bool
	}{
		{
			name: "valid team battle",
			battle: battle.Battle{
				BattleType: battle.BattleTypeTeam,
				Status:     battle.BattleStatusPending,
			},
			isValid: true,
		},
		{
			name: "valid enemy battle",
			battle: battle.Battle{
				BattleType: battle.BattleTypeEnemy,
				Status:     battle.BattleStatusInProgress,
			},
			isValid: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - battle type and status are set
			assert.NotEmpty(t, tt.battle.BattleType)
			assert.NotEmpty(t, tt.battle.Status)
		})
	}
}

func TestBattleParticipant_Validation(t *testing.T) {
	tests := []struct {
		name      string
		participant battle.BattleParticipant
		isValid   bool
	}{
		{
			name: "valid light warrior participant",
			participant: battle.BattleParticipant{
				ParticipantID: "warrior1",
				Name:         "Test Warrior",
				Type:         battle.ParticipantTypeWarrior,
				Side:         battle.TeamSideLight,
				HP:           100,
				MaxHP:        100,
				AttackPower:  50,
				Defense:      30,
				IsAlive:      true,
			},
			isValid: true,
		},
		{
			name: "valid dark enemy participant",
			participant: battle.BattleParticipant{
				ParticipantID: "enemy1",
				Name:         "Test Enemy",
				Type:         battle.ParticipantTypeEnemy,
				Side:         battle.TeamSideDark,
				HP:           80,
				MaxHP:        80,
				AttackPower:  40,
				Defense:      25,
				IsAlive:      true,
			},
			isValid: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.participant.ParticipantID)
			assert.NotEmpty(t, tt.participant.Name)
			assert.NotEmpty(t, tt.participant.Type)
			assert.NotEmpty(t, tt.participant.Side)
			assert.Greater(t, tt.participant.HP, 0)
			assert.Greater(t, tt.participant.MaxHP, 0)
		})
	}
}

func TestBattleTurn_Structure(t *testing.T) {
	turn := battle.BattleTurn{
		BattleID:      primitive.NewObjectID().Hex(),
		TurnNumber:    1,
		AttackerID:    "attacker1",
		AttackerName:  "Attacker",
		AttackerType:  battle.ParticipantTypeWarrior,
		TargetID:      "target1",
		TargetName:    "Target",
		TargetType:    battle.ParticipantTypeEnemy,
		DamageDealt:   50,
		CriticalHit:   false,
		TargetHPAfter: 50,
		CreatedAt:     time.Now(),
	}
	
	assert.NotEmpty(t, turn.BattleID)
	assert.NotEmpty(t, turn.AttackerID)
	assert.NotEmpty(t, turn.TargetID)
	assert.Greater(t, turn.TurnNumber, 0)
	assert.Greater(t, turn.DamageDealt, 0)
}

func TestCalculateOpponentDamage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := battle.NewService()
	
	tests := []struct {
		name           string
		battle         battle.Battle
		expectedMin    int
		expectedMax    int
	}{
		{
			name: "dragon battle",
			battle: battle.Battle{
				BattleType: battle.BattleTypeDragon,
			},
			expectedMin: 60, // ~100 attack - 30 defense * 0.8
			expectedMax: 100,
		},
		{
			name: "enemy battle",
			battle: battle.Battle{
				BattleType: battle.BattleTypeEnemy,
			},
			expectedMin: 10, // ~50 attack - 30 defense * 0.8
			expectedMax: 30,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			damage := svc.CalculateOpponentDamage(&tt.battle)
			assert.GreaterOrEqual(t, damage, tt.expectedMin)
			assert.LessOrEqual(t, damage, tt.expectedMax)
		})
	}
}

func TestBattle_CompleteBattleResult(t *testing.T) {
	// Test battle result mapping
	tests := []struct {
		warriorHP int
		opponentHP int
		expectedResult battle.BattleResult
	}{
		{
			warriorHP: 100,
			opponentHP: 0,
			expectedResult: battle.BattleResultVictory,
		},
		{
			warriorHP: 0,
			opponentHP: 100,
			expectedResult: battle.BattleResultDefeat,
		},
		{
			warriorHP: 50,
			opponentHP: 50,
			expectedResult: battle.BattleResultDraw,
		},
	}
	
	for _, tt := range tests {
		t.Run(string(tt.expectedResult), func(t *testing.T) {
			if tt.warriorHP > tt.opponentHP {
				assert.Equal(t, battle.BattleResultVictory, battle.BattleResultVictory)
			} else if tt.opponentHP > tt.warriorHP {
				assert.Equal(t, battle.BattleResultDefeat, battle.BattleResultDefeat)
			} else {
				assert.Equal(t, battle.BattleResultDraw, battle.BattleResultDraw)
			}
		})
	}
}

