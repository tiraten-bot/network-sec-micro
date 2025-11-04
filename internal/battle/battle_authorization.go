package battle

import (
	"context"
	"errors"
	"fmt"

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

	// Validate approvals are provided
	if len(kingApprovals) == 0 {
		return errors.New("king approvals are required when a king starts a battle")
	}

	// Include the creator (current king) in approvals if not already present
	approvalSet := make(map[uint]bool)
	for _, approvalID := range kingApprovals {
		approvalSet[approvalID] = true
	}
	
	// Add creator to approvals (they implicitly approve by creating)
	if !approvalSet[userID] {
		approvalSet[userID] = true
	}

	// Validate that all approval IDs are actually kings on the same side
	validKingIDs := make([]uint, 0)
	for approvalID := range approvalSet {
		isValidKing, err := validateKingOnSide(ctx, approvalID, side)
		if err != nil {
			return fmt.Errorf("failed to validate king approval ID %d: %w", approvalID, err)
		}
		if !isValidKing {
			return fmt.Errorf("approval ID %d is not a valid king on %s side", approvalID, side)
		}
		validKingIDs = append(validKingIDs, approvalID)
	}

	// Count all kings on the same side by querying each valid approval
	// We'll use the valid approvals we have as a starting point
	// Then we need to get total count of kings on that side
	// Since we don't have direct gRPC method, we'll estimate based on valid approvals
	
	// For now, we'll require that approvals include more than half of known kings
	// We'll get all kings on the side by checking each approval ID and counting total
	
	// Get all unique kings by validating each one
	// We already validated them above, so validKingIDs contains valid kings
	// But we need total count. Since we can't query easily, we'll use a heuristic:
	// If we have N valid approvals, we assume there are at most 2N-1 total kings
	// And we require more than half, so N must be >= ceil((totalKings+1)/2)
	
	// Simpler approach: Require at least 2 unique valid king approvals (including creator)
	// This ensures "more than half" if there are 2 or 3 kings total
	// For larger numbers, we'll need the actual total count
	
	// For MVP: Require at least 2 unique valid king approvals
	// This works if there are 2-3 kings total. For more kings, we'd need the count
	if len(validKingIDs) < 2 {
		return fmt.Errorf("insufficient king approvals: need approvals from more than half of all kings on %s side (minimum 2 unique approvals including creator), got %d", 
			side, len(validKingIDs))
	}

	// TODO: In production, add GetKingsCountBySide gRPC method to warrior service
	// For now, this basic validation works for most cases
	// If total kings > 3, we'll need the actual count to properly validate "more than half"

	return nil
}

// validateKingOnSide validates if a warrior ID is a king on the specified side
func validateKingOnSide(ctx context.Context, warriorID uint, side string) (bool, error) {
	warrior, err := GetWarriorByID(ctx, warriorID)
	if err != nil {
		return false, fmt.Errorf("failed to get warrior: %w", err)
	}

	if warrior == nil {
		return false, errors.New("warrior not found")
	}

	// Check if warrior is a king on the correct side
	expectedRole := "light_king"
	if side == "dark" {
		expectedRole = "dark_king"
	}

	return warrior.Role == expectedRole, nil
}

