package weapon

import (
	"context"
	"errors"
	"fmt"
	"time"

	"network-sec-micro/internal/weapon/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Service handles business logic for weapons
type Service struct{}

// NewService creates a new service instance
func NewService() *Service {
	return &Service{}
}

// CreateWeapon creates a new weapon
func (s *Service) CreateWeapon(ctx context.Context, cmd dto.CreateWeaponCommand) (*Weapon, error) {
	weaponType := WeaponType(cmd.Type)
	
	weapon := Weapon{
		Name:        cmd.Name,
		Description: cmd.Description,
		Type:        weaponType,
		Damage:      cmd.Damage,
		Price:       cmd.Price,
		CreatedBy:   cmd.CreatedBy,
		OwnedBy:     []string{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := WeaponColl.InsertOne(ctx, weapon)
	if err != nil {
		return nil, fmt.Errorf("failed to create weapon: %w", err)
	}

	weapon.ID = result.InsertedID.(primitive.ObjectID)
	return &weapon, nil
}

// BuyWeapon handles weapon purchase
func (s *Service) BuyWeapon(ctx context.Context, cmd dto.BuyWeaponCommand) error {
	weaponID, err := primitive.ObjectIDFromHex(cmd.WeaponID)
	if err != nil {
		return errors.New("invalid weapon ID")
	}

	var weapon Weapon
	if err := WeaponColl.FindOne(ctx, bson.M{"_id": weaponID}).Decode(&weapon); err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("weapon not found")
		}
		return err
	}

	if !weapon.CanBeBoughtBy(cmd.BuyerRole) {
		return errors.New("you don't have permission to buy this weapon")
	}

	for _, owner := range weapon.OwnedBy {
		if owner == cmd.BuyerID {
			return errors.New("you already own this weapon")
		}
	}

	weapon.OwnedBy = append(weapon.OwnedBy, cmd.BuyerID)
	weapon.UpdatedAt = time.Now()

	_, err = WeaponColl.UpdateOne(ctx,
		bson.M{"_id": weaponID},
		bson.M{"$set": bson.M{"owned_by": weapon.OwnedBy, "updated_at": weapon.UpdatedAt}},
	)

	if err != nil {
		return fmt.Errorf("failed to update weapon: %w", err)
	}

	// Publish weapon purchase event to Kafka
	// TODO: Get warriorID from username - need to add this to command
	warriorID := uint(1) // Temporary - should get from warrior service
	if err := PublishWeaponPurchase(ctx, &weapon, warriorID, cmd.BuyerID); err != nil {
		log.Printf("Failed to publish weapon purchase event: %v", err)
		// Don't fail the transaction if event publishing fails
	}

	return nil
}

// GetWeapons gets all weapons
func (s *Service) GetWeapons(ctx context.Context, query dto.GetWeaponsQuery) ([]Weapon, error) {
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

	cursor, err := WeaponColl.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query weapons: %w", err)
	}
	defer cursor.Close(ctx)

	var weapons []Weapon
	if err := cursor.All(ctx, &weapons); err != nil {
		return nil, fmt.Errorf("failed to decode weapons: %w", err)
	}

	return weapons, nil
}
