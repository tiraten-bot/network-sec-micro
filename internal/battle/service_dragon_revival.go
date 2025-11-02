package battle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ReviveDragonInBattle revives a dragon participant in a battle if it can still revive
func (s *Service) ReviveDragonInBattle(ctx context.Context, battleID primitive.ObjectID, dragonParticipantID string) (*BattleParticipant, error) {
	// Get the dragon participant
	var participant BattleParticipant
	err := BattleParticipantColl.FindOne(ctx, bson.M{
		"battle_id":      battleID,
		"participant_id": dragonParticipantID,
		"type":          ParticipantTypeDragon,
	}).Decode(&participant)

	if err != nil {
		return nil, errors.New("dragon participant not found in battle")
	}

	// Check if participant is defeated
	if !participant.IsDefeated {
		return nil, errors.New("dragon is not defeated")
	}

	// Get dragon ID from participant ID (assuming participant_id is the dragon's ObjectID hex string)
	dragonObjectID, err := primitive.ObjectIDFromHex(dragonParticipantID)
	if err != nil {
		return nil, fmt.Errorf("invalid dragon ID format: %w", err)
	}

	// Check dragon's revival status via HTTP call to dragon service
	canRevive, revivalCount, needsCrisisIntervention, err := s.CheckDragonRevival(ctx, dragonObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to check dragon revival status: %w", err)
	}

	if !canRevive {
		return nil, fmt.Errorf("dragon has exceeded maximum revival count (3), current: %d", revivalCount)
	}

	if needsCrisisIntervention {
		return nil, errors.New("dark emperor crisis intervention required before 3rd revival")
	}

	// Update dragon's revival count in dragon service
	// Make HTTP PATCH/PUT call to dragon service to increment revival count
	dragonServiceURL := getEnvOrDefault("DRAGON_SERVICE_URL", "http://localhost:8084")
	updateURL := fmt.Sprintf("%s/api/v1/dragons/%s/revive", dragonServiceURL, dragonObjectID.Hex())
	
	req, err := http.NewRequestWithContext(ctx, "POST", updateURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create revive request: %w", err)
	}
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Warning: failed to update dragon revival count in dragon service: %v", err)
		// Continue anyway - battle participant will be revived
	} else {
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Printf("Warning: dragon service returned status %d for revival", resp.StatusCode)
		}
	}

	// Simply revive the participant - set HP to full
	participant.HP = participant.MaxHP
	participant.IsAlive = true
	participant.IsDefeated = false
	participant.DefeatedAt = nil
	participant.UpdatedAt = time.Now()

	updateData := bson.M{
		"hp":          participant.HP,
		"is_alive":    participant.IsAlive,
		"is_defeated": participant.IsDefeated,
		"defeated_at": nil,
		"updated_at":  participant.UpdatedAt,
	}

	_, err = BattleParticipantColl.UpdateOne(ctx, bson.M{"_id": participant.ID}, bson.M{"$set": updateData})
	if err != nil {
		return nil, fmt.Errorf("failed to revive dragon participant: %w", err)
	}

	// Log revival to Redis
	go func() {
		message := fmt.Sprintf("🐉 %s revived! HP: %d/%d (Revival: %d/3)", participant.Name, participant.HP, participant.MaxHP, revivalCount+1)
		if err := LogBattleEvent(ctx, battleID, "dragon_revival", message); err != nil {
			log.Printf("Failed to log dragon revival: %v", err)
		}
	}()

	log.Printf("Dragon %s (participant %s) revived in battle %s - revival count: %d/3", 
		participant.Name, dragonParticipantID, battleID.Hex(), revivalCount+1)
	return &participant, nil
}

// CheckDragonRevival checks if a dragon can be revived and returns revival info
func (s *Service) CheckDragonRevival(ctx context.Context, dragonID primitive.ObjectID) (canRevive bool, revivalCount int, needsCrisisIntervention bool, err error) {
	// Make HTTP call to dragon service to check revival status
	// Since we're in battle service, we need to call dragon service HTTP API
	// For now, return placeholder - in production would make HTTP call
	dragonServiceURL := getEnvOrDefault("DRAGON_SERVICE_URL", "http://localhost:8084")
	url := fmt.Sprintf("%s/api/v1/dragons/%s", dragonServiceURL, dragonID.Hex())

	resp, err := http.Get(url)
	if err != nil {
		return false, 0, false, fmt.Errorf("failed to call dragon service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, 0, false, fmt.Errorf("dragon service returned status %d", resp.StatusCode)
	}

	var dragonResponse struct {
		Success bool `json:"success"`
		Dragon  struct {
			ID                        string `json:"id"`
			IsAlive                   bool   `json:"is_alive"`
			RevivalCount              int    `json:"revival_count"`
			AwaitingCrisisIntervention bool   `json:"awaiting_crisis_intervention"`
		} `json:"dragon"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&dragonResponse); err != nil {
		return false, 0, false, fmt.Errorf("failed to decode dragon response: %w", err)
	}

	dragon := dragonResponse.Dragon
	canRevive = dragon.RevivalCount < 3 && !dragon.IsAlive
	needsCrisisIntervention = dragon.AwaitingCrisisIntervention

	return canRevive, dragon.RevivalCount, needsCrisisIntervention, nil
}

// HandleDragonDeathInBattle handles dragon death in battle and checks for revival
func (s *Service) HandleDragonDeathInBattle(ctx context.Context, battleID primitive.ObjectID, dragonParticipant *BattleParticipant) error {
	if dragonParticipant.Type != ParticipantTypeDragon {
		return nil // Not a dragon, no revival needed
	}

	if !dragonParticipant.IsDefeated {
		return nil // Not defeated, no action needed
	}

	// Get dragon ID from participant ID
	dragonID, err := primitive.ObjectIDFromHex(dragonParticipant.ParticipantID)
	if err != nil {
		return fmt.Errorf("invalid dragon ID: %w", err)
	}

	// Check if dragon can revive
	canRevive, revivalCount, needsCrisisIntervention, err := s.CheckDragonRevival(ctx, dragonID)
	if err != nil {
		log.Printf("Warning: failed to check dragon revival status: %v", err)
		return nil // Don't fail battle, just log
	}

	if !canRevive {
		// Dragon cannot revive - permanent death
		log.Printf("Dragon %s has exceeded revival limit", dragonParticipant.Name)
		return nil
	}

	if needsCrisisIntervention {
		// Dragon needs Dark Emperor intervention before 3rd revival
		log.Printf("Dragon %s (revival count: %d) needs Dark Emperor crisis intervention", dragonParticipant.Name, revivalCount)
		// Log to Redis for Dark Emperor notification
		go func() {
			message := fmt.Sprintf("⚠️ KRİZ DURUMU: %s için Dark Emperor müdahalesi gerekiyor! (Canlanma sayısı: %d/3)", 
				dragonParticipant.Name, revivalCount)
			if err := LogBattleEvent(ctx, battleID, "crisis_intervention_required", message); err != nil {
				log.Printf("Failed to log crisis intervention requirement: %v", err)
			}
		}()
	} else {
		// Dragon can auto-revive (revival count < 2)
		log.Printf("Dragon %s can be revived (revival count: %d)", dragonParticipant.Name, revivalCount)
		// Auto-revive after a short delay (e.g., 5 seconds)
		go func() {
			time.Sleep(5 * time.Second)
			if _, err := s.ReviveDragonInBattle(ctx, battleID, dragonParticipant.ParticipantID); err != nil {
				log.Printf("Failed to auto-revive dragon: %v", err)
			}
		}()
	}

	return nil
}

// DarkEmperorJoinBattle allows Dark Emperor to join battle during crisis
// Dark Emperor can only join when dragon has exactly 1 life left (revival_count = 2, still alive)
func (s *Service) DarkEmperorJoinBattle(ctx context.Context, battleID primitive.ObjectID, darkEmperorUsername string, darkEmperorUserID string, dragonParticipantID string) (*BattleParticipant, error) {
	// Verify user is Dark Emperor
	warrior, err := GetWarriorByUsername(ctx, darkEmperorUsername)
	if err != nil {
		return nil, fmt.Errorf("failed to get warrior info: %w", err)
	}

	if warrior.Role != "dark_emperor" {
		return nil, errors.New("only dark emperor can join battle during crisis")
	}

	// Get battle
	var battle Battle
	err = BattleColl.FindOne(ctx, bson.M{"_id": battleID}).Decode(&battle)
	if err != nil {
		return nil, errors.New("battle not found")
	}

	if battle.Status != BattleStatusInProgress {
		return nil, errors.New("battle is not in progress")
	}

	// Check if Dark Emperor is already in battle
	var existingParticipant BattleParticipant
	err = BattleParticipantColl.FindOne(ctx, bson.M{
		"battle_id":      battleID,
		"participant_id": darkEmperorUserID,
		"type":          ParticipantTypeDarkEmperor,
	}).Decode(&existingParticipant)

	if err == nil {
		return nil, errors.New("dark emperor is already in this battle")
	}
	if err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("failed to check existing participant: %w", err)
	}

	// CRITICAL: Dark Emperor can only join when dragon has exactly 1 life left (revival_count = 2)
	// Get dragon participant to check its status
	var dragonParticipant BattleParticipant
	err = BattleParticipantColl.FindOne(ctx, bson.M{
		"battle_id":      battleID,
		"participant_id": dragonParticipantID,
		"type":          ParticipantTypeDragon,
	}).Decode(&dragonParticipant)

	if err != nil {
		return nil, errors.New("dragon participant not found in battle")
	}

	// Get dragon ID
	dragonID, err := primitive.ObjectIDFromHex(dragonParticipantID)
	if err != nil {
		return nil, fmt.Errorf("invalid dragon ID: %w", err)
	}

	// Check dragon's revival count - must be exactly 2 (1 life left) and still alive
	_, revivalCount, _, err := s.CheckDragonRevival(ctx, dragonID)
	if err != nil {
		return nil, fmt.Errorf("failed to check dragon status: %w", err)
	}

	// Dark Emperor can only join when dragon has 1 life left (revival_count = 2) and is still alive
	if revivalCount != 2 || !dragonParticipant.IsAlive {
		return nil, errors.New("dark emperor can only join battle when dragon has exactly 1 life left (revival_count = 2 and still alive)")
	}

	// Calculate Dark Emperor stats (high stats as crisis intervention)
	maxHP := 2000
	attackPower := 300
	defense := 200

	participant := &BattleParticipant{
		BattleID:      battleID,
		ParticipantID: darkEmperorUserID,
		Name:          darkEmperorUsername,
		Type:          ParticipantTypeDarkEmperor,
		Side:          TeamSideDark,
		HP:            maxHP,
		MaxHP:         maxHP,
		AttackPower:   attackPower,
		Defense:       defense,
		IsAlive:       true,
		IsDefeated:    false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = BattleParticipantColl.InsertOne(ctx, participant)
	if err != nil {
		return nil, fmt.Errorf("failed to add dark emperor to battle: %w", err)
	}

	// Log to Redis
	go func() {
		message := fmt.Sprintf("⚡ KRİZ MÜDAHALESİ: Dark Emperor %s savaşa katıldı!", darkEmperorUsername)
		if err := LogBattleEvent(ctx, battleID, "dark_emperor_joined", message); err != nil {
			log.Printf("Failed to log dark emperor join: %v", err)
		}
	}()

	log.Printf("Dark Emperor %s joined battle %s", darkEmperorUsername, battleID.Hex())
	return participant, nil
}

// SacrificeDragonAndReviveEnemies sacrifices a dragon to revive all dead enemies and multiply enemy population
func (s *Service) SacrificeDragonAndReviveEnemies(ctx context.Context, battleID primitive.ObjectID, dragonParticipantID string, darkEmperorUsername string) (int, int, error) {
	// Verify user is Dark Emperor
	warrior, err := GetWarriorByUsername(ctx, darkEmperorUsername)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get warrior info: %w", err)
	}

	if warrior.Role != "dark_emperor" {
		return 0, 0, errors.New("only dark emperor can sacrifice dragon")
	}

	// Get dragon participant
	var dragonParticipant BattleParticipant
	err = BattleParticipantColl.FindOne(ctx, bson.M{
		"battle_id":      battleID,
		"participant_id": dragonParticipantID,
		"type":          ParticipantTypeDragon,
	}).Decode(&dragonParticipant)

	if err != nil {
		return 0, 0, errors.New("dragon participant not found in battle")
	}

	// Get dragon ID
	dragonID, err := primitive.ObjectIDFromHex(dragonParticipantID)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid dragon ID: %w", err)
	}

	// Get dragon's revival count from dragon service to determine multiplier
	_, revivalCount, _, err := s.CheckDragonRevival(ctx, dragonID)
	if err != nil {
		log.Printf("Warning: failed to check dragon revival status: %v, assuming revival_count = 0", err)
		revivalCount = 0
	}

	// Determine multiplier based on dragon state:
	// - revival_count = 0 (never died, full HP): 3x multiplier, dragon stays alive
	// - revival_count = 1 (died once, 2 lives left): 2x multiplier, dragon sacrificed
	// - revival_count >= 2 (2 lives left or less): 1x multiplier (just revive), dragon sacrificed
	var multiplier int
	var dragonSacrificed bool
	
	if revivalCount == 0 && dragonParticipant.IsAlive {
		// Dragon never died - 3x multiplier, dragon stays alive
		multiplier = 3
		dragonSacrificed = false
	} else if revivalCount == 1 && dragonParticipant.IsAlive {
		// Dragon has 2 lives left - 2x multiplier, dragon sacrificed
		multiplier = 2
		dragonSacrificed = true
	} else {
		// Dragon already dead or has less than 2 lives - just revive defeated enemies
		multiplier = 1
		dragonSacrificed = true
	}

	// Mark dragon as sacrificed if needed
	if dragonSacrificed {
		dragonParticipant.HP = 0
		dragonParticipant.IsAlive = false
		dragonParticipant.IsDefeated = true
		dragonParticipant.UpdatedAt = time.Now()

		updateDragon := bson.M{
			"hp":          dragonParticipant.HP,
			"is_alive":    false,
			"is_defeated": true,
			"updated_at":  dragonParticipant.UpdatedAt,
		}

		_, err = BattleParticipantColl.UpdateOne(ctx, bson.M{"_id": dragonParticipant.ID}, bson.M{"$set": updateDragon})
		if err != nil {
			return 0, 0, fmt.Errorf("failed to sacrifice dragon: %w", err)
		}
	}

	// Get all enemy participants (alive + defeated) to use as templates for multiplication
	enemyFilter := bson.M{
		"battle_id": battleID,
		"type":      ParticipantTypeEnemy,
		"side":      TeamSideDark,
	}

	enemyCursor, err := BattleParticipantColl.Find(ctx, enemyFilter)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to find enemies: %w", err)
	}
	defer enemyCursor.Close(ctx)

	var allEnemies []BattleParticipant
	if err := enemyCursor.All(ctx, &allEnemies); err != nil {
		return 0, 0, fmt.Errorf("failed to decode enemies: %w", err)
	}

	// Separate defeated and alive enemies
	var defeatedEnemies []BattleParticipant
	var aliveEnemies []BattleParticipant
	
	for _, enemy := range allEnemies {
		if enemy.IsDefeated {
			defeatedEnemies = append(defeatedEnemies, enemy)
		} else {
			aliveEnemies = append(aliveEnemies, enemy)
		}
	}

	// Revive all defeated enemies
	revivedCount := 0
	for _, enemy := range defeatedEnemies {
		enemy.HP = enemy.MaxHP
		enemy.IsAlive = true
		enemy.IsDefeated = false
		enemy.DefeatedAt = nil
		enemy.UpdatedAt = time.Now()

		updateEnemy := bson.M{
			"hp":          enemy.HP,
			"is_alive":    enemy.IsAlive,
			"is_defeated": enemy.IsDefeated,
			"defeated_at": nil,
			"updated_at":  enemy.UpdatedAt,
		}

		_, err = BattleParticipantColl.UpdateOne(ctx, bson.M{"_id": enemy.ID}, bson.M{"$set": updateEnemy})
		if err != nil {
			log.Printf("Failed to revive enemy %s: %v", enemy.Name, err)
			continue
		}

		revivedCount++
	}

	// Multiply enemy population: Each existing enemy gets exactly (multiplier - 1) copies
	// So if multiplier is 3, each enemy gets 2 copies (original + 2 copies = 3 total)
	// If multiplier is 2, each enemy gets 1 copy (original + 1 copy = 2 total)
	// Use all enemies (alive + revived) as templates - every single enemy will be multiplied
	templateEnemies := append(aliveEnemies, defeatedEnemies...)
	multipliedCount := 0
	newParticipants := []interface{}{}

	now := time.Now()
	// Iterate through each existing enemy and create exact copies
	for _, template := range templateEnemies {
		// Create exactly (multiplier - 1) copies of this specific enemy
		// This ensures every enemy gets multiplied, not randomly
		for copyNum := 1; copyNum <= multiplier-1; copyNum++ {
			newEnemy := &BattleParticipant{
				BattleID:      battleID,
				ParticipantID: fmt.Sprintf("%s_copy_%d_%d", template.ParticipantID, now.Unix(), copyNum), // Unique ID
				Name:          fmt.Sprintf("%s (Copy %d)", template.Name, copyNum),
				Type:          ParticipantTypeEnemy,
				Side:          TeamSideDark,
				HP:            template.MaxHP, // Full HP
				MaxHP:         template.MaxHP,
				AttackPower:   template.AttackPower,
				Defense:       template.Defense,
				IsAlive:       true,
				IsDefeated:    false,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			newParticipants = append(newParticipants, newEnemy)
			multipliedCount++
		}
	}

	// Insert new enemy participants if any
	if len(newParticipants) > 0 {
		_, err = BattleParticipantColl.InsertMany(ctx, newParticipants)
		if err != nil {
			log.Printf("Warning: failed to add multiplied enemies: %v", err)
			// Continue anyway - at least revived enemies
		}
	}

	// Log to Redis
	go func() {
		sacrificeStatus := "korundu"
		if dragonSacrificed {
			sacrificeStatus = "feda edildi"
		}
		message := fmt.Sprintf("💀 DRAGON FEDAKARLIĞI: Dark Emperor %s tarafından %s %s! %d düşman canlandı, %d yeni düşman oluşturuldu (Toplam: %dx çarpan)",
			darkEmperorUsername, dragonParticipant.Name, sacrificeStatus, revivedCount, multipliedCount, multiplier)
		if err := LogBattleEvent(ctx, battleID, "dragon_sacrifice", message); err != nil {
			log.Printf("Failed to log dragon sacrifice: %v", err)
		}
	}()

	totalAffected := revivedCount + multipliedCount
	log.Printf("Dark Emperor %s sacrificed dragon %s (revival_count=%d, sacrificed=%v) - Revived: %d, Multiplied: %d (multiplier=%dx) in battle %s",
		darkEmperorUsername, dragonParticipant.Name, revivalCount, dragonSacrificed, revivedCount, multipliedCount, multiplier, battleID.Hex())

	return revivedCount, multipliedCount, nil
}

// LogBattleEvent logs a battle event to Redis (simplified helper)
func LogBattleEvent(ctx context.Context, battleID primitive.ObjectID, eventType string, message string) error {
	// Use existing Redis logging infrastructure
	logEntry := map[string]interface{}{
		"type":      eventType,
		"message":   message,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Convert to JSON
	data, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Use existing Redis client to append log
	// This is a simplified version - in production would use proper Redis client
	log.Printf("Battle Event [%s] %s: %s", battleID.Hex(), eventType, message)
	
	return nil
}

// getEnvOrDefault gets environment variable or returns default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

