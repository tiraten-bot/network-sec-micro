package arena

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	pbWarrior "network-sec-micro/api/proto/warrior"
	"network-sec-micro/internal/arena/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Service handles arena business logic with CQRS pattern
type Service struct{}

// NewService creates a new arena service
func NewService() *Service {
	return &Service{}
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

	// Calculate HP based on total power
	challengerMaxHP := int(challenger.TotalPower) * 10
	if challengerMaxHP < 100 {
		challengerMaxHP = 100
	}

	opponentMaxHP := int(opponent.TotalPower) * 10
	if opponentMaxHP < 100 {
		opponentMaxHP = 100
	}

	// Create arena match directly (no battle service dependency)
	now := time.Now()
	match := &ArenaMatch{
		Player1ID:      invitation.ChallengerID,
		Player1Name:    invitation.ChallengerName,
		Player1HP:      challengerMaxHP,
		Player1MaxHP:   challengerMaxHP,
		Player1Attack:  int(challenger.AttackPower),
		Player1Defense: int(challenger.Defense),
		Player2ID:      invitation.OpponentID,
		Player2Name:    invitation.OpponentName,
		Player2HP:      opponentMaxHP,
		Player2MaxHP:   opponentMaxHP,
		Player2Attack:  int(opponent.AttackPower),
		Player2Defense: int(opponent.Defense),
		CurrentTurn:    0,
		MaxTurns:        50, // Default for arena battles
		CurrentAttacker: 1,  // Player1 starts first
		Status:          MatchStatusInProgress,
		StartedAt:       &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	result, err := MatchColl.InsertOne(ctx, match)
	if err != nil {
		return nil, fmt.Errorf("failed to create match: %w", err)
	}

	match.ID = result.InsertedID.(primitive.ObjectID)

	// Update invitation
	updateData := bson.M{
		"status":       InvitationStatusAccepted,
		"responded_at": now,
		"battle_id":    match.ID.Hex(), // Store match ID as battle_id for reference
		"updated_at":   now,
	}

	_, err = InvitationColl.UpdateOne(ctx, bson.M{"_id": invitationID}, bson.M{"$set": updateData})
	if err != nil {
		return nil, fmt.Errorf("failed to update invitation: %w", err)
	}

	// Publish Kafka events
	go func() {
		matchID := match.ID.Hex()
		if err := PublishInvitationAccepted(invitation.ID.Hex(), invitation.ChallengerID, invitation.ChallengerName, invitation.OpponentID, invitation.OpponentName, matchID); err != nil {
			log.Printf("Failed to publish invitation accepted event: %v", err)
		}
		if err := PublishMatchStarted(matchID, match.Player1ID, match.Player1Name, match.Player2ID, match.Player2Name, matchID); err != nil {
			log.Printf("Failed to publish match started event: %v", err)
		}
	}()

	log.Printf("Arena invitation accepted: %s vs %s, match: %s", invitation.ChallengerName, invitation.OpponentName, match.ID.Hex())
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

// PerformAttack performs an attack in an arena match
func (s *Service) PerformAttack(ctx context.Context, matchID primitive.ObjectID, attackerID uint) (*ArenaMatch, error) {
	var match ArenaMatch
	err := MatchColl.FindOne(ctx, bson.M{"_id": matchID}).Decode(&match)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("match not found")
		}
		return nil, fmt.Errorf("failed to get match: %w", err)
	}

	if match.Status != MatchStatusInProgress {
		return nil, errors.New("match is not in progress")
	}

	// Validate attacker
	var attackerHP, attackerAttack *int
	var defenderHP, defenderDefense *int
	var defenderID uint
	var attackerName, defenderName string

	if attackerID == match.Player1ID {
		if match.CurrentAttacker != 1 && match.CurrentAttacker != 0 {
			return nil, errors.New("not your turn")
		}
		attackerHP = &match.Player1HP
		attackerAttack = &match.Player1Attack
		defenderHP = &match.Player2HP
		defenderDefense = &match.Player2Defense
		defenderID = match.Player2ID
		attackerName = match.Player1Name
		defenderName = match.Player2Name
	} else if attackerID == match.Player2ID {
		if match.CurrentAttacker != 2 {
			return nil, errors.New("not your turn")
		}
		attackerHP = &match.Player2HP
		attackerAttack = &match.Player2Attack
		defenderHP = &match.Player1HP
		defenderDefense = &match.Player1Defense
		defenderID = match.Player1ID
		attackerName = match.Player2Name
		defenderName = match.Player1Name
	} else {
		return nil, errors.New("you are not a participant in this match")
	}

	// Calculate damage
	damage := *attackerAttack - *defenderDefense
	if damage < 10 {
		damage = 10 // Minimum damage
	}

	// Apply damage
	*defenderHP -= damage
	if *defenderHP < 0 {
		*defenderHP = 0
	}

	// Update match
	match.CurrentTurn++
	// Switch attacker for next turn
	if match.CurrentAttacker == 1 {
		match.CurrentAttacker = 2
	} else {
		match.CurrentAttacker = 1
	}

	// Update HP values
	if attackerID == match.Player1ID {
		match.Player2HP = *defenderHP
	} else {
		match.Player1HP = *defenderHP
	}

    // Threshold events (first time HP <= 50%)
    checkAndPublish := func(pHP, pMax int, announced *bool, pid uint, pname string) {
        if pMax > 0 && !*announced && (pHP*100 <= pMax*50) {
            percent := float64(pHP) / float64(pMax) * 100.0
            _ = PublishSpellWindowOpened(match.ID.Hex(), pid, pname, percent)
            *announced = true
        }
    }
    checkAndPublish(match.Player1HP, match.Player1MaxHP, &match.P1Below50Announced, match.Player1ID, match.Player1Name)
    checkAndPublish(match.Player2HP, match.Player2MaxHP, &match.P2Below50Announced, match.Player2ID, match.Player2Name)

    // Crisis threshold (<=10%)
    checkAndPublishCrisis := func(pHP, pMax int, announced *bool, pid uint, pname string) {
        if pMax > 0 && !*announced && (pHP*100 <= pMax*10) {
            percent := float64(pHP) / float64(pMax) * 100.0
            _ = PublishCrisisWindowOpened(match.ID.Hex(), pid, pname, percent)
            *announced = true
        }
    }
    checkAndPublishCrisis(match.Player1HP, match.Player1MaxHP, &match.P1Below10Announced, match.Player1ID, match.Player1Name)
    checkAndPublishCrisis(match.Player2HP, match.Player2MaxHP, &match.P2Below10Announced, match.Player2ID, match.Player2Name)

    // Check if match is over
	if *defenderHP <= 0 {
		// Defender is defeated
		match.Status = MatchStatusCompleted
		now := time.Now()
		match.CompletedAt = &now

		// Set winner
		if attackerID == match.Player1ID {
			winnerID := match.Player1ID
			match.WinnerID = &winnerID
			match.WinnerName = match.Player1Name
		} else {
			winnerID := match.Player2ID
			match.WinnerID = &winnerID
			match.WinnerName = match.Player2Name
		}

		// Publish match completed event
		go func() {
			if err := PublishMatchCompleted(
				match.ID.Hex(),
				match.Player1ID,
				match.Player1Name,
				match.Player2ID,
				match.Player2Name,
				match.WinnerID,
				match.WinnerName,
				match.ID.Hex(),
			); err != nil {
				log.Printf("Failed to publish match completed event: %v", err)
			}
		}()
	} else if match.CurrentTurn >= match.MaxTurns {
		// Match timeout - draw (or determine winner by HP)
		if match.Player1HP > match.Player2HP {
			winnerID := match.Player1ID
			match.WinnerID = &winnerID
			match.WinnerName = match.Player1Name
		} else if match.Player2HP > match.Player1HP {
			winnerID := match.Player2ID
			match.WinnerID = &winnerID
			match.WinnerName = match.Player2Name
		}
		// If equal HP, no winner (draw)

		match.Status = MatchStatusCompleted
		now := time.Now()
		match.CompletedAt = &now

		go func() {
			if err := PublishMatchCompleted(
				match.ID.Hex(),
				match.Player1ID,
				match.Player1Name,
				match.Player2ID,
				match.Player2Name,
				match.WinnerID,
				match.WinnerName,
				match.ID.Hex(),
			); err != nil {
				log.Printf("Failed to publish match completed event: %v", err)
			}
		}()
	}

	match.UpdatedAt = time.Now()

	updateData := bson.M{
		"player1_hp":      match.Player1HP,
		"player2_hp":      match.Player2HP,
		"current_turn":    match.CurrentTurn,
		"current_attacker": match.CurrentAttacker,
		"status":          match.Status,
		"updated_at":      match.UpdatedAt,
	}

	if match.CompletedAt != nil {
		updateData["completed_at"] = match.CompletedAt
	}
	if match.WinnerID != nil {
		updateData["winner_id"] = *match.WinnerID
		updateData["winner_name"] = match.WinnerName
	}

	_, err = MatchColl.UpdateOne(ctx, bson.M{"_id": matchID}, bson.M{"$set": updateData})
	if err != nil {
		return nil, fmt.Errorf("failed to update match: %w", err)
	}

	log.Printf("Arena attack: %s dealt %d damage to %s (HP: %d)", attackerName, damage, defenderName, *defenderHP)
	return &match, nil
}

// ApplySpellEffect applies a 1v1 arenaspell's immediate effect to the match
func (s *Service) ApplySpellEffect(ctx context.Context, matchID primitive.ObjectID, casterID uint, spellType string) (*ArenaMatch, error) {
    var match ArenaMatch
    err := MatchColl.FindOne(ctx, bson.M{"_id": matchID}).Decode(&match)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, errors.New("match not found")
        }
        return nil, fmt.Errorf("failed to get match: %w", err)
    }

    if match.Status != MatchStatusInProgress {
        return nil, errors.New("match is not in progress")
    }

    // Resolve caster/opponent pointers
    var casterAttack, casterDefense, casterHP, casterMaxHP *int
    var opponentAttack, opponentDefense *int
    var casterName, opponentName string

    if casterID == match.Player1ID {
        casterAttack = &match.Player1Attack
        casterDefense = &match.Player1Defense
        casterHP = &match.Player1HP
        casterMaxHP = &match.Player1MaxHP
        opponentAttack = &match.Player2Attack
        opponentDefense = &match.Player2Defense
        casterName = match.Player1Name
        opponentName = match.Player2Name
    } else if casterID == match.Player2ID {
        casterAttack = &match.Player2Attack
        casterDefense = &match.Player2Defense
        casterHP = &match.Player2HP
        casterMaxHP = &match.Player2MaxHP
        opponentAttack = &match.Player1Attack
        opponentDefense = &match.Player1Defense
        casterName = match.Player2Name
        opponentName = match.Player1Name
    } else {
        return nil, errors.New("caster is not a participant in this match")
    }

    // Enforce spell window: allow only if any player's HP <= 50%
    allow := func(hp, max int) bool { return max > 0 && (hp*100 <= max*50) }
    if !(allow(match.Player1HP, match.Player1MaxHP) || allow(match.Player2HP, match.Player2MaxHP)) {
        return nil, errors.New("spell window not open (no player below or equal to 50% HP)")
    }

    // Apply effect
    switch spellType {
    case "call_of_the_light_king":
        *casterAttack = *casterAttack * 2
    case "resistance":
        *casterDefense = *casterDefense * 2
    case "rebirth":
        if *casterHP == 0 {
            half := (*casterMaxHP) / 2
            if half < 1 { half = 1 }
            *casterHP = half
        }
    case "destroy_the_light":
        // reduce opponent stats by 30% (stack enforcement handled by arenaspell)
        reduce := func(v int) int { nv := int(float64(v) * 0.7); if nv < 1 { nv = 1 }; return nv }
        *opponentAttack = reduce(*opponentAttack)
        *opponentDefense = reduce(*opponentDefense)
    default:
        return nil, fmt.Errorf("unsupported spell type: %s", spellType)
    }

    match.UpdatedAt = time.Now()
    update := bson.M{
        "player1_hp":      match.Player1HP,
        "player1_attack":  match.Player1Attack,
        "player1_defense": match.Player1Defense,
        "player2_hp":      match.Player2HP,
        "player2_attack":  match.Player2Attack,
        "player2_defense": match.Player2Defense,
        "updated_at":      match.UpdatedAt,
    }

    _, err = MatchColl.UpdateOne(ctx, bson.M{"_id": matchID}, bson.M{"$set": update})
    if err != nil {
        return nil, fmt.Errorf("failed to update match: %w", err)
    }

    log.Printf("Arena spell applied: %s by %s on %s", spellType, casterName, opponentName)
    return &match, nil
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

