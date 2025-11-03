package arena

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	pbWarrior "network-sec-micro/api/proto/warrior"
	"network-sec-micro/internal/arena/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Service handles arena business logic with CQRS pattern
type Service struct {
	battleServiceURL string
}

// NewService creates a new arena service
func NewService() *Service {
	battleURL := getEnv("BATTLE_SERVICE_URL", "http://localhost:8085")
	return &Service{
		battleServiceURL: battleURL,
	}
}

const (
	InvitationExpirationMinutes = 10 // Invitations expire after 10 minutes
)

// ==================== COMMANDS (WRITE OPERATIONS) ====================

// SendInvitation sends an arena invitation from challenger to opponent
func (s *Service) SendInvitation(ctx context.Context, cmd dto.SendInvitationCommand) (*ArenaInvitation, error) {
	// Validate: challenger cannot challenge themselves
	if cmd.ChallengerName == cmd.OpponentName {
		return nil, errors.New("cannot challenge yourself")
	}

	// Get opponent info via gRPC
	opponent, err := GetWarriorByUsername(ctx, cmd.OpponentName)
	if err != nil {
		return nil, fmt.Errorf("opponent not found: %w", err)
	}

	opponentID := uint(opponent.Id)

	// Check if there's already a pending invitation between these two
	var existingInvitation ArenaInvitation
	err = InvitationColl.FindOne(ctx, bson.M{
		"challenger_id": cmd.ChallengerID,
		"opponent_id":   opponentID,
		"status":        InvitationStatusPending,
	}).Decode(&existingInvitation)

	if err == nil {
		// Check if expired
		if existingInvitation.IsExpired() {
			// Mark as expired and continue
			s.markInvitationAsExpired(ctx, existingInvitation.ID)
		} else {
			return nil, fmt.Errorf("invitation already sent to %s", cmd.OpponentName)
		}
	}

	// Check if opponent has a pending invitation to challenger
	err = InvitationColl.FindOne(ctx, bson.M{
		"challenger_id": opponentID,
		"opponent_id":   cmd.ChallengerID,
		"status":        InvitationStatusPending,
	}).Decode(&existingInvitation)

	if err == nil && !existingInvitation.IsExpired() {
		return nil, fmt.Errorf("%s has already sent you an invitation. Please accept or reject it first", cmd.OpponentName)
	}

	// Create invitation
	now := time.Now()
	expiresAt := now.Add(InvitationExpirationMinutes * time.Minute)

	invitation := &ArenaInvitation{
		ChallengerID:   cmd.ChallengerID,
		ChallengerName: cmd.ChallengerName,
		OpponentID:     opponentID,
		OpponentName:   cmd.OpponentName,
		Status:          InvitationStatusPending,
		ExpiresAt:       expiresAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	result, err := InvitationColl.InsertOne(ctx, invitation)
	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	invitation.ID = result.InsertedID.(primitive.ObjectID)

	// Publish Kafka event
	go func() {
		expiresAtStr := expiresAt.Format(time.RFC3339)
		if err := PublishInvitationSent(invitation.ID.Hex(), cmd.ChallengerID, cmd.ChallengerName, opponentID, cmd.OpponentName, expiresAtStr); err != nil {
			log.Printf("Failed to publish invitation sent event: %v", err)
		}
	}()

	log.Printf("Arena invitation sent: %s -> %s", cmd.ChallengerName, cmd.OpponentName)
	return invitation, nil
}

// AcceptInvitation accepts an arena invitation and starts the match
func (s *Service) AcceptInvitation(ctx context.Context, cmd dto.AcceptInvitationCommand) (*ArenaMatch, error) {
	invitationID, err := primitive.ObjectIDFromHex(cmd.InvitationID)
	if err != nil {
		return nil, errors.New("invalid invitation ID")
	}

	// Get invitation
	var invitation ArenaInvitation
	err = InvitationColl.FindOne(ctx, bson.M{"_id": invitationID}).Decode(&invitation)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("invitation not found")
		}
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	// Validate: only opponent can accept
	if invitation.OpponentID != cmd.OpponentID {
		return nil, errors.New("only the invited player can accept the invitation")
	}

	// Check if can be accepted
	if !invitation.CanBeAccepted() {
		if invitation.IsExpired() {
			s.markInvitationAsExpired(ctx, invitationID)
			return nil, errors.New("invitation has expired")
		}
		return nil, fmt.Errorf("invitation cannot be accepted (status: %s)", invitation.Status)
	}

	// Get both warriors' info via gRPC
	challenger, err := GetWarriorByID(ctx, invitation.ChallengerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get challenger info: %w", err)
	}

	opponent, err := GetWarriorByID(ctx, invitation.OpponentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get opponent info: %w", err)
	}

	// Start arena battle via Battle Service HTTP API
	battleID, err := s.startArenaBattle(ctx, invitation.ChallengerID, invitation.ChallengerName, invitation.OpponentID, invitation.OpponentName)
	if err != nil {
		return nil, fmt.Errorf("failed to start arena battle: %w", err)
	}

	// Update invitation
	now := time.Now()
	updateData := bson.M{
		"status":       InvitationStatusAccepted,
		"responded_at": now,
		"battle_id":    battleID,
		"updated_at":   now,
	}

	_, err = InvitationColl.UpdateOne(ctx, bson.M{"_id": invitationID}, bson.M{"$set": updateData})
	if err != nil {
		return nil, fmt.Errorf("failed to update invitation: %w", err)
	}

	// Create arena match
	match := &ArenaMatch{
		Player1ID:   invitation.ChallengerID,
		Player1Name: invitation.ChallengerName,
		Player2ID:   invitation.OpponentID,
		Player2Name: invitation.OpponentName,
		BattleID:    battleID,
		Status:      MatchStatusInProgress,
		StartedAt:   &now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	result, err := MatchColl.InsertOne(ctx, match)
	if err != nil {
		return nil, fmt.Errorf("failed to create match: %w", err)
	}

	match.ID = result.InsertedID.(primitive.ObjectID)

	// Publish Kafka events
	go func() {
		if err := PublishInvitationAccepted(invitation.ID.Hex(), invitation.ChallengerID, invitation.ChallengerName, invitation.OpponentID, invitation.OpponentName, battleID); err != nil {
			log.Printf("Failed to publish invitation accepted event: %v", err)
		}
		if err := PublishMatchStarted(match.ID.Hex(), match.Player1ID, match.Player1Name, match.Player2ID, match.Player2Name, battleID); err != nil {
			log.Printf("Failed to publish match started event: %v", err)
		}
	}()

	log.Printf("Arena invitation accepted: %s vs %s, battle: %s", invitation.ChallengerName, invitation.OpponentName, battleID)
	return match, nil
}

// RejectInvitation rejects an arena invitation
func (s *Service) RejectInvitation(ctx context.Context, cmd dto.RejectInvitationCommand) error {
	invitationID, err := primitive.ObjectIDFromHex(cmd.InvitationID)
	if err != nil {
		return errors.New("invalid invitation ID")
	}

	// Get invitation
	var invitation ArenaInvitation
	err = InvitationColl.FindOne(ctx, bson.M{"_id": invitationID}).Decode(&invitation)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("invitation not found")
		}
		return fmt.Errorf("failed to get invitation: %w", err)
	}

	// Validate: only opponent can reject
	if invitation.OpponentID != cmd.OpponentID {
		return errors.New("only the invited player can reject the invitation")
	}

	if invitation.Status != InvitationStatusPending {
		return fmt.Errorf("invitation cannot be rejected (status: %s)", invitation.Status)
	}

	// Update invitation
	now := time.Now()
	updateData := bson.M{
		"status":       InvitationStatusRejected,
		"responded_at": now,
		"updated_at":   now,
	}

	_, err = InvitationColl.UpdateOne(ctx, bson.M{"_id": invitationID}, bson.M{"$set": updateData})
	if err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	// Publish Kafka event
	go func() {
		if err := PublishInvitationRejected(invitation.ID.Hex(), invitation.ChallengerID, invitation.ChallengerName, invitation.OpponentID, invitation.OpponentName); err != nil {
			log.Printf("Failed to publish invitation rejected event: %v", err)
		}
	}()

	log.Printf("Arena invitation rejected: %s -> %s", invitation.ChallengerName, invitation.OpponentName)
	return nil
}

// CancelInvitation cancels an invitation (only by challenger)
func (s *Service) CancelInvitation(ctx context.Context, cmd dto.CancelInvitationCommand) error {
	invitationID, err := primitive.ObjectIDFromHex(cmd.InvitationID)
	if err != nil {
		return errors.New("invalid invitation ID")
	}

	// Get invitation
	var invitation ArenaInvitation
	err = InvitationColl.FindOne(ctx, bson.M{"_id": invitationID}).Decode(&invitation)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("invitation not found")
		}
		return fmt.Errorf("failed to get invitation: %w", err)
	}

	// Validate: only challenger can cancel
	if invitation.ChallengerID != cmd.ChallengerID {
		return errors.New("only the challenger can cancel the invitation")
	}

	if invitation.Status != InvitationStatusPending {
		return fmt.Errorf("invitation cannot be cancelled (status: %s)", invitation.Status)
	}

	// Update invitation
	now := time.Now()
	updateData := bson.M{
		"status":     InvitationStatusCancelled,
		"updated_at": now,
	}

	_, err = InvitationColl.UpdateOne(ctx, bson.M{"_id": invitationID}, bson.M{"$set": updateData})
	if err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	log.Printf("Arena invitation cancelled: %s -> %s", invitation.ChallengerName, invitation.OpponentName)
	return nil
}

// startArenaBattle starts a 1v1 arena battle via Battle Service HTTP API
func (s *Service) startArenaBattle(ctx context.Context, player1ID uint, player1Name string, player2ID uint, player2Name string) (string, error) {
	// Create arena battle request to battle service
	type ArenaBattleRequest struct {
		Player1ID   uint   `json:"player1_id"`
		Player1Name string `json:"player1_name"`
		Player2ID   uint   `json:"player2_id"`
		Player2Name string `json:"player2_name"`
		BattleType  string `json:"battle_type"` // "arena"
	}

	reqBody := ArenaBattleRequest{
		Player1ID:   player1ID,
		Player1Name: player1Name,
		Player2ID:   player2ID,
		Player2Name: player2Name,
		BattleType:  "arena",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Call battle service
	url := fmt.Sprintf("%s/api/arena/start", s.battleServiceURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call battle service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("battle service returned error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	type ArenaBattleResponse struct {
		BattleID string `json:"battle_id"`
	}

	var battleResp ArenaBattleResponse
	if err := json.NewDecoder(resp.Body).Decode(&battleResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return battleResp.BattleID, nil
}

// markInvitationAsExpired marks an invitation as expired
func (s *Service) markInvitationAsExpired(ctx context.Context, invitationID primitive.ObjectID) {
	var invitation ArenaInvitation
	err := InvitationColl.FindOne(ctx, bson.M{"_id": invitationID}).Decode(&invitation)
	if err != nil {
		return
	}

	if invitation.Status == InvitationStatusPending {
		updateData := bson.M{
			"status":     InvitationStatusExpired,
			"updated_at": time.Now(),
		}
		InvitationColl.UpdateOne(ctx, bson.M{"_id": invitationID}, bson.M{"$set": updateData})

		// Publish event
		go func() {
			if err := PublishInvitationExpired(invitation.ID.Hex(), invitation.ChallengerID, invitation.ChallengerName, invitation.OpponentID, invitation.OpponentName); err != nil {
				log.Printf("Failed to publish invitation expired event: %v", err)
			}
		}()
	}
}

// ==================== QUERIES (READ OPERATIONS) ====================

// GetInvitation gets an invitation by ID
func (s *Service) GetInvitation(ctx context.Context, query dto.GetInvitationQuery) (*ArenaInvitation, error) {
	invitationID, err := primitive.ObjectIDFromHex(query.InvitationID)
	if err != nil {
		return nil, errors.New("invalid invitation ID")
	}

	var invitation ArenaInvitation
	err = InvitationColl.FindOne(ctx, bson.M{"_id": invitationID}).Decode(&invitation)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("invitation not found")
		}
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	// Check if expired
	if invitation.IsExpired() && invitation.Status == InvitationStatusPending {
		s.markInvitationAsExpired(ctx, invitationID)
		invitation.Status = InvitationStatusExpired
	}

	return &invitation, nil
}

// GetMyInvitations gets user's invitations (sent or received)
func (s *Service) GetMyInvitations(ctx context.Context, query dto.GetMyInvitationsQuery) ([]ArenaInvitation, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"challenger_id": query.UserID},
			{"opponent_id": query.UserID},
		},
	}

	if query.Status != "" {
		filter["status"] = ArenaInvitationStatus(query.Status)
	}

	cursor, err := InvitationColl.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find invitations: %w", err)
	}
	defer cursor.Close(ctx)

	var invitations []ArenaInvitation
	if err := cursor.All(ctx, &invitations); err != nil {
		return nil, fmt.Errorf("failed to decode invitations: %w", err)
	}

	// Check for expired invitations
	for i := range invitations {
		if invitations[i].IsExpired() && invitations[i].Status == InvitationStatusPending {
			s.markInvitationAsExpired(ctx, invitations[i].ID)
			invitations[i].Status = InvitationStatusExpired
		}
	}

	return invitations, nil
}

// GetMyMatches gets user's arena matches
func (s *Service) GetMyMatches(ctx context.Context, query dto.GetMyMatchesQuery) ([]ArenaMatch, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"player1_id": query.UserID},
			{"player2_id": query.UserID},
		},
	}

	if query.Status != "" {
		filter["status"] = ArenaMatchStatus(query.Status)
	}

	cursor, err := MatchColl.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find matches: %w", err)
	}
	defer cursor.Close(ctx)

	var matches []ArenaMatch
	if err := cursor.All(ctx, &matches); err != nil {
		return nil, fmt.Errorf("failed to decode matches: %w", err)
	}

	return matches, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

