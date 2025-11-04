package armor

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"network-sec-micro/internal/armor/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Service handles business logic for armors
type Service struct{}

// NewService creates a new service instance
func NewService() *Service {
	return &Service{}
}

// CreateArmor creates a new armor
func (s *Service) CreateArmor(ctx context.Context, cmd dto.CreateArmorCommand) (*Armor, error) {
	armorType := ArmorType(cmd.Type)

	armor := Armor{
		Name:         cmd.Name,
		Description:  cmd.Description,
		Type:         armorType,
		Defense:      cmd.Defense,
		HPBonus:      cmd.HPBonus,
		Price:        cmd.Price,
		CreatedBy:    cmd.CreatedBy,
		OwnedBy:      []string{},
		Durability:   cmd.MaxDurability, // Initialize with max durability
		MaxDurability: cmd.MaxDurability,
		IsBroken:     false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	result, err := ArmorColl.InsertOne(ctx, armor)
	if err != nil {
		return nil, fmt.Errorf("failed to create armor: %w", err)
	}

	armor.ID = result.InsertedID.(primitive.ObjectID)
	return &armor, nil
}

// BuyArmor handles armor purchase
func (s *Service) BuyArmor(ctx context.Context, cmd dto.BuyArmorCommand) error {
	armorID, err := primitive.ObjectIDFromHex(cmd.ArmorID)
	if err != nil {
		return errors.New("invalid armor ID")
	}

	var armor Armor
	if err := ArmorColl.FindOne(ctx, bson.M{"_id": armorID}).Decode(&armor); err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("armor not found")
		}
		return err
	}

	if !armor.CanBeBoughtBy(cmd.BuyerRole) {
		return errors.New("you don't have permission to buy this armor")
	}

	// Check if already owned (legacy check)
	for _, owner := range armor.OwnedBy {
		if owner == cmd.BuyerID {
			return errors.New("you already own this armor")
		}
	}

	// Add to owned_by (legacy)
	armor.OwnedBy = append(armor.OwnedBy, cmd.BuyerID)
	
	// Add to owners array (new generalized ownership)
	ownerType := cmd.OwnerType
	if ownerType == "" {
		ownerType = "warrior" // Default to warrior for backward compatibility
	}
	armor.Owners = append(armor.Owners, OwnerRef{
		OwnerType: ownerType,
		OwnerID:   cmd.BuyerID,
	})

	armor.UpdatedAt = time.Now()

	_, err = ArmorColl.UpdateOne(ctx,
		bson.M{"_id": armorID},
		bson.M{"$set": bson.M{"owned_by": armor.OwnedBy, "owners": armor.Owners, "updated_at": armor.UpdatedAt}},
	)

	if err != nil {
		return fmt.Errorf("failed to update armor: %w", err)
	}

	// Publish armor purchase event to Kafka
	if err := PublishArmorPurchase(ctx, &armor, cmd.BuyerUserID, cmd.BuyerUsername, ownerType); err != nil {
		log.Printf("Failed to publish armor purchase event: %v", err)
		// Don't fail the transaction if event publishing fails
	}

	return nil
}

// GetArmors gets all armors
func (s *Service) GetArmors(ctx context.Context, query dto.GetArmorsQuery) ([]Armor, error) {
	filter := bson.M{}

	if query.Type != "" {
		filter["type"] = query.Type
	}
	if query.CreatedBy != "" {
		filter["created_by"] = query.CreatedBy
	}
	if query.OwnedBy != "" {
		filter["owned_by"] = query.OwnedBy
	}

	cursor, err := ArmorColl.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query armors: %w", err)
	}
	defer cursor.Close(ctx)

	var armors []Armor
	if err := cursor.All(ctx, &armors); err != nil {
		return nil, fmt.Errorf("failed to decode armors: %w", err)
	}

	return armors, nil
}

// GetArmorsByOwner gets armors owned by a specific owner
func GetArmorsByOwner(ctx context.Context, ownerType, ownerID string) ([]Armor, error) {
	filter := bson.M{}
	
	// Prefer new owners array; fallback to legacy for warriors
	if ownerType != "" && ownerID != "" {
		filter = bson.M{"owners": bson.M{"$elemMatch": bson.M{"owner_type": ownerType, "owner_id": ownerID}}}
		if ownerType == "warrior" {
			// Also include legacy field match
			filter = bson.M{"$or": []bson.M{
				{"owners": bson.M{"$elemMatch": bson.M{"owner_type": ownerType, "owner_id": ownerID}}},
				{"owned_by": ownerID},
			}}
		}
	}

	cursor, err := ArmorColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var armors []Armor
	if err := cursor.All(ctx, &armors); err != nil {
		return nil, err
	}

	return armors, nil
}

