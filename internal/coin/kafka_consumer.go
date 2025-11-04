package coin

import (
	"context"
	"encoding/json"
	"log"
    "strconv"

	pb "network-sec-micro/api/proto/coin"
    "network-sec-micro/pkg/kafka"
)

// WeaponPurchaseEvent represents the event structure
type WeaponPurchaseEvent struct {
	EventType     string `json:"event_type"`
	Timestamp     string `json:"timestamp"`
	SourceService string `json:"source_service"`
	WeaponID      string `json:"weapon_id"`
	WarriorID     uint   `json:"warrior_id"`
	WarriorName   string `json:"warrior_name"`
	WeaponName    string `json:"weapon_name"`
	WeaponPrice   int    `json:"weapon_price"`
}

// WeaponRepairEvent event to deduct coins for repair
type WeaponRepairEvent struct {
    Type      string `json:"type"`
    OwnerType string `json:"owner_type"`
    OwnerID   string `json:"owner_id"`
    Cost      int    `json:"cost"`
    WeaponID  string `json:"weapon_id"`
    OrderID   string `json:"order_id"`
}

// ArmorRepairEvent event to deduct coins for armor repair
type ArmorRepairEvent struct {
    Type      string `json:"type"`
    OwnerType string `json:"owner_type"`
    OwnerID   string `json:"owner_id"`
    Cost      int    `json:"cost"`
    ArmorID   string `json:"armor_id"`
    OrderID   string `json:"order_id"`
}

// ArmorPurchaseEvent represents the event structure for armor purchase
type ArmorPurchaseEvent struct {
    EventType     string `json:"event_type"`
    Timestamp     string `json:"timestamp"`
    SourceService string `json:"source_service"`
    ArmorID       string `json:"armor_id"`
    BuyerID       uint   `json:"buyer_id"`
    BuyerName     string `json:"buyer_name"`
    ArmorName     string `json:"armor_name"`
    ArmorPrice    int    `json:"armor_price"`
    OwnerType     string `json:"owner_type"`
}

// HandleWeaponPurchase handles weapon purchase events from Kafka
func (s *CoinServiceServer) HandleWeaponPurchase(event WeaponPurchaseEvent) error {
	log.Printf("Received weapon purchase event: %+v", event)

	// Deduct coins from warrior's balance
	ctx := context.Background()
	_, err := s.DeductCoins(ctx, &pb.DeductCoinsRequest{
		WarriorId: uint32(event.WarriorID),
		Amount:    int64(event.WeaponPrice),
		Reason:    "weapon_purchase: " + event.WeaponName,
	})

	if err != nil {
		log.Printf("Failed to deduct coins for warrior %d: %v", event.WarriorID, err)
		return err
	}

	log.Printf("Successfully deducted %d coins from warrior %d", event.WeaponPrice, event.WarriorID)
	return nil
}

// ProcessKafkaMessage processes incoming Kafka messages
func ProcessKafkaMessage(message []byte) error {
	// Try to unmarshal as weapon purchase event
	var weaponEvent WeaponPurchaseEvent
	if err := json.Unmarshal(message, &weaponEvent); err == nil {
		if weaponEvent.EventType == "weapon_purchased" {
			// Handle weapon purchase
			service := NewService()
			server := NewCoinServiceServer(service)
			return server.HandleWeaponPurchase(weaponEvent)
		}
	}

    // Try to unmarshal as weapon.repair event
    var repair WeaponRepairEvent
    if err := json.Unmarshal(message, &repair); err == nil {
        if repair.Type == "weapon.repair" && repair.Cost > 0 {
            // Only warriors have coin accounts directly; for dragons/enemies, you may map to creators/owners as needed.
            if repair.OwnerType == "warrior" && repair.OwnerID != "" {
                if id64, err := strconv.ParseUint(repair.OwnerID, 10, 32); err == nil {
                    ctx := context.Background()
                    service := NewService()
                    server := NewCoinServiceServer(service)
                    _, err := server.DeductCoins(ctx, &pb.DeductCoinsRequest{WarriorId: uint32(id64), Amount: int64(repair.Cost), Reason: "weapon_repair"})
                    if err != nil { log.Printf("Failed to deduct coins for repair: %v", err) }
                }
            }
            return nil
        }
    }

    // Try to unmarshal as armor purchase event
    var armorEvent ArmorPurchaseEvent
    if err := json.Unmarshal(message, &armorEvent); err == nil {
        if armorEvent.EventType == "armor_purchased" {
            // Deduct coins (only for warrior ownerType in this simple flow)
            if armorEvent.OwnerType == "warrior" {
                ctx := context.Background()
                service := NewService()
                server := NewCoinServiceServer(service)
                _, err := server.DeductCoins(ctx, &pb.DeductCoinsRequest{WarriorId: uint32(armorEvent.BuyerID), Amount: int64(armorEvent.ArmorPrice), Reason: "armor_purchase: " + armorEvent.ArmorName})
                if err != nil { log.Printf("Failed to deduct coins for armor purchase: %v", err) }
            }
            return nil
        }
    }

	// Try to unmarshal as arena match completed
	var arenaCompleted kafka.ArenaMatchCompletedEvent
	if err := json.Unmarshal(message, &arenaCompleted); err == nil {
		if arenaCompleted.Event.EventType == "arena_match_completed" && arenaCompleted.WinnerID != nil {
			winnerID := *arenaCompleted.WinnerID
			var loserID uint
			if winnerID == arenaCompleted.Player1ID { loserID = arenaCompleted.Player2ID } else { loserID = arenaCompleted.Player1ID }
			// Fetch loser warrior to derive coin award amount (use total_power)
			if w, err := GetWarriorByID(loserID); err == nil {
				amount := int64(w.TotalPower)
				ctx := context.Background()
				service := NewService()
				server := NewCoinServiceServer(service)
				_, err := server.AddCoins(ctx, &pb.AddCoinsRequest{WarriorId: uint32(winnerID), Amount: amount, Reason: "arena_victory"})
				if err != nil { log.Printf("Failed to add coins for arena victory: %v", err) }
				return nil
			}
		}
	}

	// Try to unmarshal as enemy attack event
	if err := ProcessEnemyAttackMessage(message); err == nil {
		return nil // Successfully processed
	}

	// Try to unmarshal as battle wager resolved
	var wager kafka.BattleWagerResolvedEvent
	if err := json.Unmarshal(message, &wager); err == nil {
		if wager.Event.EventType == "battle_wager_resolved" && wager.WagerAmount > 0 {
			winnerIDStr := wager.LightEmperorID
			if wager.WinnerSide == "dark" { winnerIDStr = wager.DarkEmperorID }
			if winnerIDStr != "" {
				if id64, err := strconv.ParseUint(winnerIDStr, 10, 32); err == nil {
					ctx := context.Background()
					service := NewService()
					server := NewCoinServiceServer(service)
					_, err := server.AddCoins(ctx, &pb.AddCoinsRequest{WarriorId: uint32(id64), Amount: int64(wager.WagerAmount), Reason: "battle_wager"})
					if err != nil { log.Printf("Failed to add wager coins: %v", err) }
				}
			}
			return nil
		}
	}

	log.Printf("Unknown event type or failed to process message")
	return nil
}

