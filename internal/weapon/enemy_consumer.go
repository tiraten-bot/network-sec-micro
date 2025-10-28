package weapon

import (
	"context"
	"encoding/json"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PirateWeaponStealEvent represents pirate weapon steal event
type PirateWeaponStealEvent struct {
	EventType     string `json:"event_type"`
	Timestamp     string `json:"timestamp"`
	SourceService string `json:"source_service"`
	EnemyID       string `json:"enemy_id"`
	EnemyType     string `json:"enemy_type"`
	EnemyName     string `json:"enemy_name"`
	WarriorID     uint   `json:"warrior_id"`
	WarriorName   string `json:"warrior_name"`
	AttackType    string `json:"attack_type"`
	WeaponID      string `json:"weapon_id"`
}

// ProcessEnemyAttackMessage processes enemy attack events from Kafka
func ProcessEnemyAttackMessage(message []byte) error {
	var event PirateWeaponStealEvent
	if err := json.Unmarshal(message, &event); err != nil {
		log.Printf("Failed to unmarshal enemy attack event: %v", err)
		return err
	}

	// Only handle pirate weapon steal events
	if event.AttackType != "weapon_steal" || event.EnemyType != "pirate" {
		return nil
	}

	log.Printf("Processing pirate weapon steal: %s stole weapon %s from warrior %d", 
		event.EnemyName, event.WeaponID, event.WarriorID)

	// Remove weapon from warrior's owned weapons
	weaponID, err := primitive.ObjectIDFromHex(event.WeaponID)
	if err != nil {
		log.Printf("Invalid weapon ID: %v", err)
		return err
	}

	ctx := context.Background()
	
	// Get weapon
	var weapon Weapon
	if err := WeaponColl.FindOne(ctx, bson.M{"_id": weaponID}).Decode(&weapon); err != nil {
		log.Printf("Weapon not found: %v", err)
		return err
	}

	// Remove warrior from owned_by array
	updatedOwnedBy := make([]string, 0)
	for _, owner := range weapon.OwnedBy {
		if owner != event.WarriorName {
			updatedOwnedBy = append(updatedOwnedBy, owner)
		}
	}

	// Update weapon
	_, err = WeaponColl.UpdateOne(ctx,
		bson.M{"_id": weaponID},
		bson.M{"$set": bson.M{"owned_by": updatedOwnedBy}},
	)
	if err != nil {
		log.Printf("Failed to update weapon: %v", err)
		return err
	}

	log.Printf("Successfully removed weapon %s from warrior %s", event.WeaponID, event.WarriorName)
	return nil
}

