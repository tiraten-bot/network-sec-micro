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

type Service struct{}

func NewService() *Service {
	return &Service{}
}

// CreateEnemy creates a new enemy
func (s *Service) CreateEnemy(ctx context.Context, cmd dto.CreateEnemyCommand) (*Enemy, error) {
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

	result, err := EnemyColl.InsertOne(ctx, enemy)
	if err != nil {
		return nil, fmt.Errorf("failed to create enemy: %w", err)
	}

	enemy.ID = result.InsertedID.(primitive.ObjectID)
	return &enemy, nil
}

// AttackWarrior handles goblin coin steal or pirate weapon steal
func (s *Service) AttackWarrior(ctx context.Context, cmd dto.AttackWarriorCommand) error {
	var enemy Enemy
	enemyID, err := primitive.ObjectIDFromHex(cmd.EnemyID)
	if err != nil {
		return errors.New("invalid enemy ID")
	}

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

func publishEnemyAttackEvent(event *kafka.EnemyAttackEvent) error {
	publisher, err := GetKafkaPublisher()
	if err != nil {
		return err
	}
	return publisher.Publish(kafka.TopicEnemyAttack, event)
}
