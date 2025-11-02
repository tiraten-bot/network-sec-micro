package battle

import (
	"context"
	"errors"
	"fmt"

	"network-sec-micro/internal/battle/dto"
	pbWarrior "network-sec-micro/api/proto/warrior"
)

// ValidateBattleAuthorization validates if user can start a battle
// Rules:
// 1. Emperors (light_emperor, dark_emperor) can start battles directly
// 2. Kings (light_king, dark_king) need approval from more than half of all kings on their side
func ValidateBattleAuthorization(ctx context.Context, userRole string, userID uint, kingApprovals []uint) error {
	// Check if user is an emperor
	if userRole == "light_emperor" || userRole == "dark_emperor" {
		return nil // Emperors can start battles directly
	}

	// Check if user is a king
	if userRole != "light_king" && userRole != "dark_king" {
		return errors.New("only emperors or kings can start battles")
	}

	// If user is a king, need approvals
	// Determine which side the king is on
	var side string
	if userRole == "light_king" {
		side = "light"
	} else {
		side = "dark"
	}

	// Get all kings on the same side via gRPC
	// We need to get all warriors and filter by role
	// For now, we'll use a simplified approach - get all kings
	// In production, we'd have a dedicated gRPC method for this
	
	// Get current user's warrior info to confirm they're a king
	warriorClient := GetWarriorClient()
	if warriorClient == nil {
		return errors.New("failed to connect to warrior service")
	}

	// Get all warriors (we'll filter for kings)
	// Note: This requires a new gRPC method or we filter client-side
	// For now, let's create a helper that gets kings via gRPC
	
	// Count total kings on the same side
	totalKings, err := countKingsOnSide(ctx, side)
	if err != nil {
		return fmt.Errorf("failed to count kings: %w", err)
	}

	// Include the creator (current king) in approvals if not already present
	approvalSet := make(map[uint]bool)
	for _, approvalID := range kingApprovals {
		approvalSet[approvalID] = true
	}
	
	// Add creator to approvals (they implicitly approve by creating)
	if !approvalSet[userID] {
		kingApprovals = append(kingApprovals, userID)
		approvalSet[userID] = true
	}

	// Count unique approvals
	uniqueApprovals := len(approvalSet)

	// Calculate required approvals: more than half
	requiredApprovals := (totalKings / 2) + 1 // More than half

	if uniqueApprovals < requiredApprovals {
		return fmt.Errorf("insufficient king approvals: need %d approvals (more than half of %d kings), got %d", 
			requiredApprovals, totalKings, uniqueApprovals)
	}

	// Validate that all approval IDs are actually kings on the same side
	for _, approvalID := range kingApprovals {
		isValidKing, err := validateKingOnSide(ctx, approvalID, side)
		if err != nil {
			return fmt.Errorf("failed to validate king approval: %w", err)
		}
		if !isValidKing {
			return fmt.Errorf("approval ID %d is not a valid king on %s side", approvalID, side)
		}
	}

	return nil
}

// countKingsOnSide counts total kings on a specific side
func countKingsOnSide(ctx context.Context, side string) (int, error) {
	warriorClient := GetWarriorClient()
	if warriorClient == nil {
		return 0, errors.New("warrior gRPC client not available")
	}

	// Determine king role based on side
	var kingRole string
	if side == "light" {
		kingRole = "light_king"
	} else {
		kingRole = "dark_king"
	}

	// We need to get all warriors and filter - this is inefficient
	// In production, we'd have a dedicated gRPC method: GetWarriorsByRole
	// For now, we'll need to implement a workaround or add the method
	
	// Since we don't have a direct method, we'll need to get warriors differently
	// Let's use the fact that we can query via database or add a new gRPC method
	
	// For now, let's assume we can get all kings and filter
	// We'll need to add this functionality to warrior service or use a workaround
	
	// Temporary solution: We'll need to pass total kings count or fetch via service
	// Let's return a placeholder for now and implement properly
	
	// TODO: Implement proper counting via gRPC or direct database access
	// For now, we'll return an error if we can't determine
	return 0, fmt.Errorf("king counting not yet implemented - requires warrior service gRPC method")
}

// validateKingOnSide validates if a warrior ID is a king on the specified side
func validateKingOnSide(ctx context.Context, warriorID uint, side string) (bool, error) {
	warriorClient := GetWarriorClient()
	if warriorClient == nil {
		return false, errors.New("warrior gRPC client not available")
	}

	// Get warrior by ID
	resp, err := warriorClient.GetWarriorByID(ctx, &pbWarrior.GetWarriorByIDRequest{
		WarriorId: uint32(warriorID),
	})
	if err != nil {
		return false, fmt.Errorf("failed to get warrior: %w", err)
	}

	if resp.Warrior == nil {
		return false, errors.New("warrior not found")
	}

	// Check if warrior is a king on the correct side
	expectedRole := "light_king"
	if side == "dark" {
		expectedRole = "dark_king"
	}

	return resp.Warrior.Role == expectedRole, nil
}

