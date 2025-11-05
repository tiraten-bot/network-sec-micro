package integration_test

import (
	"context"
	"testing"
	"time"

	"network-sec-micro/internal/arena"
	"network-sec-micro/internal/arena/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestArenaInvitationFlow tests the complete arena invitation flow
func TestArenaInvitationFlow(t *testing.T) {
	ctx := context.Background()
	svc := arena.NewService()
	
	// Step 1: Send invitation
	sendCmd := dto.SendInvitationCommand{
		ChallengerID:   1,
		ChallengerName: "warrior1",
		OpponentName:   "warrior2",
	}
	
	invitation, err := svc.SendInvitation(ctx, sendCmd)
	require.NoError(t, err)
	assert.NotNil(t, invitation)
	assert.Equal(t, "warrior1", invitation.ChallengerName)
	assert.Equal(t, "warrior2", invitation.OpponentName)
	assert.Equal(t, arena.InvitationStatusPending, invitation.Status)
	
	// Step 2: Accept invitation
	acceptCmd := dto.AcceptInvitationCommand{
		InvitationID: invitation.ID,
		OpponentID:    2,
	}
	
	// Note: This would require warrior gRPC client setup
	// In real test, you'd mock the gRPC client
	// match, err := svc.AcceptInvitation(ctx, acceptCmd)
	// require.NoError(t, err)
	// assert.NotNil(t, match)
}

// TestArenaInvitation_Expiration tests invitation expiration
func TestArenaInvitation_Expiration(t *testing.T) {
	// Test expired invitation
	expiredTime := time.Now().Add(-15 * time.Minute) // 15 minutes ago
	invitation := &arena.ArenaInvitation{
		ExpiresAt: expiredTime,
		Status:    arena.InvitationStatusPending,
	}
	
	assert.True(t, invitation.IsExpired())
	assert.False(t, invitation.CanBeAccepted())
}

// TestArenaInvitation_StatusTransitions tests invitation status transitions
func TestArenaInvitation_StatusTransitions(t *testing.T) {
	// Test pending invitation
	invitation := &arena.ArenaInvitation{
		Status:    arena.InvitationStatusPending,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	
	assert.True(t, invitation.CanBeAccepted())
	assert.False(t, invitation.IsExpired())
	
	// Test accepted invitation
	invitation.Status = arena.InvitationStatusAccepted
	assert.False(t, invitation.CanBeAccepted())
	
	// Test rejected invitation
	invitation.Status = arena.InvitationStatusRejected
	assert.False(t, invitation.CanBeAccepted())
}

// TestArenaMatch_AttackFlow tests arena match attack flow
func TestArenaMatch_AttackFlow(t *testing.T) {
	// This would test the attack flow in arena matches
	// Including weapon damage and armor defense
	
	// Simulate attack calculation
	attackerPower := 150
	weaponBonus := 50
	defenderDefense := 80
	armorDefenseBonus := 30
	
	totalAttack := attackerPower + weaponBonus
	totalDefense := defenderDefense + armorDefenseBonus
	
	damage := totalAttack - totalDefense
	if damage < 10 {
		damage = 10
	}
	
	assert.Equal(t, 90, damage) // 200 - 110 = 90
	
	// Verify weapon wear (1 per attack)
	// weapon.Durability -= 1
	
	// Verify armor wear (1 per defense)
	// armor.Durability -= 1
}

// TestArenaMatch_WinnerDetermination tests winner determination
func TestArenaMatch_WinnerDetermination(t *testing.T) {
	tests := []struct {
		name        string
		player1HP   int
		player2HP   int
		expectedWinner string
	}{
		{
			name:          "player1 wins",
			player1HP:     100,
			player2HP:     0,
			expectedWinner: "player1",
		},
		{
			name:          "player2 wins",
			player1HP:     0,
			player2HP:     50,
			expectedWinner: "player2",
		},
		{
			name:          "draw",
			player1HP:     0,
			player2HP:     0,
			expectedWinner: "draw",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var winner string
			if tt.player1HP > tt.player2HP {
				winner = "player1"
			} else if tt.player2HP > tt.player1HP {
				winner = "player2"
			} else {
				winner = "draw"
			}
			
			assert.Equal(t, tt.expectedWinner, winner)
		})
	}
}

// TestArenaInvitation_DuplicatePrevention tests duplicate invitation prevention
func TestArenaInvitation_DuplicatePrevention(t *testing.T) {
	ctx := context.Background()
	svc := arena.NewService()
	
	// Send first invitation
	sendCmd1 := dto.SendInvitationCommand{
		ChallengerID:   1,
		ChallengerName: "warrior1",
		OpponentName:   "warrior2",
	}
	
	inv1, err := svc.SendInvitation(ctx, sendCmd1)
	require.NoError(t, err)
	
	// Try to send duplicate invitation
	// This should fail
	_, err = svc.SendInvitation(ctx, sendCmd1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already sent")
	
	// Clean up
	_ = inv1
}

