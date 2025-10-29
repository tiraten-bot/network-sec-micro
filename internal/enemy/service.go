package enemy

import (
	"context"
	"errors"
	"fmt"
	"time"

	"network-sec-micro/internal/enemy/dto"
	kafka "network-sec-micro/pkg/kafka"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Service handles enemy business logic with CQRS pattern
type Service struct{}

func NewService() *Service {
	return &Service{}
}

// ==================== COMMANDS (WRITE OPERATIONS) ====================

// CreateEnemy creates a new enemy
func (s *Service) CreateEnemy(cmd dto.CreateEnemyCommand) (*Enemy, error) {
	enemy := Enemy{
		Name:        cmd.Name,
		Type:        EnemyType(cmd.Type),
		Level:       cmd.Level,
		Health:      cmd.Health,
		AttackPower: cmd.AttackPower,
		CreatedBy:   cmd.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	ctx := context.Background()
	result, err := EnemyColl.InsertOne(ctx, enemy)
	if err != nil {
		return nil, fmt.Errorf("failed to create enemy: %w", err)
	}

	enemy.ID = result.InsertedID.(primitive.ObjectID)
	return &enemy, nil
}

// AttackWarrior handles enemy attacks (goblin coin steal or pirate weapon steal)
func (s *Service) AttackWarrior(cmd dto.AttackWarriorCommand) error {
	ctx := context.Background()

	enemyID, err := primitive.ObjectIDFromHex(cmd.EnemyID)
	if err != nil {
		return errors.New("invalid enemy ID")
	}

	var enemy Enemy
	if err := EnemyColl.FindOne(ctx, bson.M{"_id": enemyID}).Decode(&enemy); err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("enemy not found")
		}
		return err
	}

	// Goblin steals coins
	if enemy.Type == EnemyTypeGoblin {
		stolenCoins := cmd.Amount
		event := kafka.NewGoblinCoinStealEvent(
			enemy.ID.Hex(),
			enemy.Name,
			cmd.WarriorName,
			int(cmd.WarriorID),
			stolenCoins,
		)
		if err := publishEnemyAttackEvent(event); err != nil {
			return fmt.Errorf("failed to publish coin steal event: %w", err)
		}
	}

	// Pirate steals weapon
	if enemy.Type == EnemyTypePirate {
		event := kafka.NewPirateWeaponStealEvent(
			enemy.ID.Hex(),
			enemy.Name,
			cmd.WeaponID,
			cmd.WarriorName,
			int(cmd.WarriorID),
		)
		if err := publishEnemyAttackEvent(event); err != nil {
			return fmt.Errorf("failed to publish weapon steal event: %w", err)
		}
	}

	return nil
}

// DestroyEnemy marks an enemy as destroyed and publishes an event
func (s *Service) DestroyEnemy(cmd dto.DestroyEnemyCommand) error {
    ctx := context.Background()

    enemyID, err := primitive.ObjectIDFromHex(cmd.EnemyID)
    if err != nil {
        return errors.New("invalid enemy ID")
    }

    var enemy Enemy
    if err := EnemyColl.FindOne(ctx, bson.M{"_id": enemyID}).Decode(&enemy); err != nil {
        if err == mongo.ErrNoDocuments {
            return errors.New("enemy not found")
        }
        return err
    }

    // Remove the enemy document to represent destruction
    if _, err := EnemyColl.DeleteOne(ctx, bson.M{"_id": enemyID}); err != nil {
        return fmt.Errorf("failed to destroy enemy: %w", err)
    }

    // Publish enemy destroyed event
    evt := kafka.NewEnemyDestroyedEvent(
        enemy.ID.Hex(),
        string(enemy.Type),
        enemy.Name,
        cmd.KillerWarriorName,
        cmd.KillerWarriorID,
    )
    publisher, err := GetKafkaPublisher()
    if err != nil {
        return err
    }
    if err := publisher.Publish(kafka.TopicEnemyDestroyed, evt); err != nil {
        return fmt.Errorf("failed to publish enemy destroyed event: %w", err)
    }

    return nil
}

// ==================== QUERIES (READ OPERATIONS) ====================

// GetEnemy gets an enemy by ID
func (s *Service) GetEnemy(query dto.GetEnemyQuery) (*Enemy, error) {
	enemyID, err := primitive.ObjectIDFromHex(query.EnemyID)
	if err != nil {
		return nil, errors.New("invalid enemy ID")
	}

	var enemy Enemy
	ctx := context.Background()
	if err := EnemyColl.FindOne(ctx, bson.M{"_id": enemyID}).Decode(&enemy); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("enemy not found")
		}
		return nil, err
	}

	return &enemy, nil
}

// GetEnemiesByType gets enemies by type
func (s *Service) GetEnemiesByType(query dto.GetEnemiesByTypeQuery) ([]Enemy, int64, error) {
	ctx := context.Background()
	filter := bson.M{"type": query.Type}

	count, err := EnemyColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := EnemyColl.Find(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var enemies []Enemy
	if err := cursor.All(ctx, &enemies); err != nil {
		return nil, 0, err
	}

	return enemies, count, nil
}

// GetEnemiesByCreator gets enemies created by a specific user
func (s *Service) GetEnemiesByCreator(query dto.GetEnemiesByCreatorQuery) ([]Enemy, int64, error) {
	ctx := context.Background()
	filter := bson.M{"created_by": query.CreatedBy}

	count, err := EnemyColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := EnemyColl.Find(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var enemies []Enemy
	if err := cursor.All(ctx, &enemies); err != nil {
		return nil, 0, err
	}

	return enemies, count, nil
}

// Helper function
func publishEnemyAttackEvent(event *kafka.EnemyAttackEvent) error {
	publisher, err := GetKafkaPublisher()
	if err != nil {
		return err
	}
	return publisher.Publish(kafka.TopicEnemyAttack, event)
}
