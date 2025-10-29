package dragon

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"network-sec-micro/internal/dragon/dto"
	"network-sec-micro/pkg/kafka"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Service handles dragon business logic with CQRS pattern
type Service struct {
	grpcClient *WarriorClient
}

// NewService creates a new dragon service
func NewService() *Service {
	return &Service{
		grpcClient: GetWarriorClient(),
	}
}

// ==================== COMMANDS (WRITE OPERATIONS) ====================

// CreateDragon creates a new dragon
func (s *Service) CreateDragon(cmd dto.CreateDragonCommand) (*Dragon, error) {
	// Validate dragon type
	dragonType := DragonType(cmd.Type)
	if dragonType != DragonTypeFire && dragonType != DragonTypeIce &&
		dragonType != DragonTypeLightning && dragonType != DragonTypeShadow {
		return nil, errors.New("invalid dragon type")
	}

	// Check if creator can create dragons
	if !dragonType.CanBeCreatedBy(cmd.CreatedByRole) {
		return nil, errors.New("only dark emperor can create dragons")
	}

	// Generate dragon stats based on type and level
	health, attackPower, defense := s.generateDragonStats(dragonType, cmd.Level)

	dragon := &Dragon{
		Name:        cmd.Name,
		Type:        dragonType,
		Level:       cmd.Level,
		Health:      health,
		MaxHealth:   health,
		AttackPower: attackPower,
		Defense:     defense,
		CreatedBy:   cmd.CreatedBy,
		IsAlive:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	ctx := context.Background()
	result, err := DragonColl.InsertOne(ctx, dragon)
	if err != nil {
		return nil, fmt.Errorf("failed to create dragon: %w", err)
	}

	dragon.ID = result.InsertedID.(primitive.ObjectID)
	return dragon, nil
}

// AttackDragon handles dragon attack by warrior
func (s *Service) AttackDragon(cmd dto.AttackDragonCommand) (*Dragon, error) {
	ctx := context.Background()

	// Get dragon
	var dragon Dragon
	err := DragonColl.FindOne(ctx, bson.M{"_id": cmd.DragonID}).Decode(&dragon)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("dragon not found")
		}
		return nil, fmt.Errorf("failed to get dragon: %w", err)
	}

	// Check if dragon is alive
	if !dragon.IsAlive {
		return nil, errors.New("dragon is already dead")
	}

	// Get warrior info via gRPC
	warrior, err := s.grpcClient.GetWarriorByUsername(ctx, cmd.AttackerUsername)
	if err != nil {
		return nil, fmt.Errorf("failed to get warrior info: %w", err)
	}

	// Check if warrior can kill dragons
	if !dragon.CanBeKilledBy(warrior.Role) {
		return nil, errors.New("only light king or light emperor can kill dragons")
	}

	// Calculate damage (warrior power vs dragon defense)
	damage := s.calculateDamage(warrior.TotalPower, dragon.Defense)
	dragon.TakeDamage(damage)

	// Update dragon
	updateData := bson.M{
		"health":     dragon.Health,
		"is_alive":   dragon.IsAlive,
		"updated_at": time.Now(),
	}

	if !dragon.IsAlive {
		now := time.Now()
		updateData["killed_by"] = cmd.AttackerUsername
		updateData["killed_at"] = now
		dragon.KilledBy = cmd.AttackerUsername
		dragon.KilledAt = &now

		// Publish dragon death event for weapon loot
		go s.publishDragonDeathEvent(dragon, cmd.AttackerUsername)
	}

	_, err = DragonColl.UpdateOne(ctx, bson.M{"_id": cmd.DragonID}, bson.M{"$set": updateData})
	if err != nil {
		return nil, fmt.Errorf("failed to update dragon: %w", err)
	}

	return &dragon, nil
}

// ==================== QUERIES (READ OPERATIONS) ====================

// GetDragon gets a dragon by ID
func (s *Service) GetDragon(query dto.GetDragonQuery) (*Dragon, error) {
	ctx := context.Background()

	var dragon Dragon
	err := DragonColl.FindOne(ctx, bson.M{"_id": query.DragonID}).Decode(&dragon)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("dragon not found")
		}
		return nil, fmt.Errorf("failed to get dragon: %w", err)
	}

	return &dragon, nil
}

// GetDragonsByType gets dragons by type
func (s *Service) GetDragonsByType(query dto.GetDragonsByTypeQuery) ([]Dragon, error) {
	ctx := context.Background()

	filter := bson.M{"type": query.Type}
	if query.AliveOnly {
		filter["is_alive"] = true
	}

	cursor, err := DragonColl.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find dragons: %w", err)
	}
	defer cursor.Close(ctx)

	var dragons []Dragon
	if err = cursor.All(ctx, &dragons); err != nil {
		return nil, fmt.Errorf("failed to decode dragons: %w", err)
	}

	return dragons, nil
}

// GetDragonsByCreator gets dragons created by a specific user
func (s *Service) GetDragonsByCreator(query dto.GetDragonsByCreatorQuery) ([]Dragon, error) {
	ctx := context.Background()

	filter := bson.M{"created_by": query.CreatorUsername}
	if query.AliveOnly {
		filter["is_alive"] = true
	}

	cursor, err := DragonColl.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find dragons: %w", err)
	}
	defer cursor.Close(ctx)

	var dragons []Dragon
	if err = cursor.All(ctx, &dragons); err != nil {
		return nil, fmt.Errorf("failed to decode dragons: %w", err)
	}

	return dragons, nil
}

// ==================== HELPER METHODS ====================

// generateDragonStats generates dragon stats based on type and level
func (s *Service) generateDragonStats(dragonType DragonType, level int) (health, attackPower, defense int) {
	baseHealth := 1000
	baseAttack := 200
	baseDefense := 150

	// Type modifiers
	switch dragonType {
	case DragonTypeFire:
		baseHealth += 200
		baseAttack += 100
		baseDefense += 50
	case DragonTypeIce:
		baseHealth += 150
		baseAttack += 80
		baseDefense += 100
	case DragonTypeLightning:
		baseHealth += 100
		baseAttack += 150
		baseDefense += 30
	case DragonTypeShadow:
		baseHealth += 300
		baseAttack += 120
		baseDefense += 80
	}

	// Level scaling
	health = baseHealth + (level * 100)
	attackPower = baseAttack + (level * 20)
	defense = baseDefense + (level * 15)

	return health, attackPower, defense
}

// calculateDamage calculates damage dealt to dragon
func (s *Service) calculateDamage(warriorPower, dragonDefense int) int {
	// Base damage calculation
	baseDamage := warriorPower - dragonDefense
	if baseDamage < 10 {
		baseDamage = 10 // Minimum damage
	}

	// Add some randomness (±20%)
	randomFactor := 0.8 + (rand.Float64() * 0.4)
	return int(float64(baseDamage) * randomFactor)
}

// publishDragonDeathEvent publishes dragon death event for weapon loot
func (s *Service) publishDragonDeathEvent(dragon Dragon, killerUsername string) {
	// Generate random weapon loot
	weaponTypes := []string{"sword", "bow", "staff", "dagger", "axe", "mace", "spear", "wand"}
	randomWeaponType := weaponTypes[rand.Intn(len(weaponTypes))]

	event := kafka.DragonDeathEvent{
		EventType:       "dragon_death",
		Timestamp:       time.Now().Format(time.RFC3339),
		SourceService:   "dragon",
		DragonID:        dragon.ID.Hex(),
		DragonName:      dragon.Name,
		DragonType:      string(dragon.Type),
		DragonLevel:     dragon.Level,
		KillerUsername:  killerUsername,
		LootWeaponType: randomWeaponType,
		LootWeaponName: fmt.Sprintf("%s of %s", randomWeaponType, dragon.Name),
	}

	if err := PublishDragonDeathEvent(event); err != nil {
		log.Printf("Failed to publish dragon death event: %v", err)
	}
}
