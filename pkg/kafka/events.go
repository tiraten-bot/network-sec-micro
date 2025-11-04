package kafka

import "time"

// Event represents a base event
type Event struct {
	EventType    string    `json:"event_type"`
	Timestamp    time.Time `json:"timestamp"`
	SourceService string   `json:"source_service"`
}

// WeaponPurchaseEvent represents a weapon purchase event
type WeaponPurchaseEvent struct {
	Event
	WeaponID     string `json:"weapon_id"`
	WarriorID    uint   `json:"warrior_id"`
	WarriorName  string `json:"warrior_name"`
	WeaponName   string `json:"weapon_name"`
	WeaponPrice  int    `json:"weapon_price"`
}

// NewWeaponPurchaseEvent creates a new weapon purchase event
func NewWeaponPurchaseEvent(weaponID, warriorName, weaponName string, warriorID, weaponPrice int) *WeaponPurchaseEvent {
	return &WeaponPurchaseEvent{
		Event: Event{
			EventType:     "weapon_purchased",
			Timestamp:     time.Now(),
			SourceService: "weapon",
		},
		WeaponID:    weaponID,
		WarriorID:   uint(warriorID),
		WarriorName: warriorName,
		WeaponName:  weaponName,
		WeaponPrice: weaponPrice,
	}
}

// DragonDeathEvent represents dragon death event for weapon loot
type DragonDeathEvent struct {
	EventType       string `json:"event_type"`
	Timestamp       string `json:"timestamp"`
	SourceService   string `json:"source_service"`
	DragonID        string `json:"dragon_id"`
	DragonName      string `json:"dragon_name"`
	DragonType      string `json:"dragon_type"`
	DragonLevel     int    `json:"dragon_level"`
    DragonMaxHealth int    `json:"dragon_max_health"`
    DragonAttack    int    `json:"dragon_attack_power"`
    DragonDefense   int    `json:"dragon_defense"`
	KillerUsername  string `json:"killer_username"`
	LootWeaponType  string `json:"loot_weapon_type"`
	LootWeaponName  string `json:"loot_weapon_name"`
}

// DragonRevivalEvent represents dragon revival event
type DragonRevivalEvent struct {
	Event
	DragonID            string `json:"dragon_id"`
	DragonName          string `json:"dragon_name"`
	DragonType          string `json:"dragon_type"`
	DragonLevel         int    `json:"dragon_level"`
	DragonMaxHealth     int    `json:"dragon_max_health"`
	RevivalCount        int    `json:"revival_count"`
	BattleID            string `json:"battle_id,omitempty"` // Optional: if revived in battle context
	AwaitingCrisisIntervention bool `json:"awaiting_crisis_intervention"` // True if next death needs crisis intervention
}

// NewDragonRevivalEvent creates a new dragon revival event
func NewDragonRevivalEvent(dragonID, dragonName, dragonType string, dragonLevel, dragonMaxHealth, revivalCount int, battleID string, awaitingCrisisIntervention bool) *DragonRevivalEvent {
	return &DragonRevivalEvent{
		Event: Event{
			EventType:     "dragon_revival",
			Timestamp:     time.Now(),
			SourceService: "dragon",
		},
		DragonID:                  dragonID,
		DragonName:                dragonName,
		DragonType:                dragonType,
		DragonLevel:               dragonLevel,
		DragonMaxHealth:           dragonMaxHealth,
		RevivalCount:              revivalCount,
		BattleID:                  battleID,
		AwaitingCrisisIntervention: awaitingCrisisIntervention,
	}
}

// EnemyDestroyedEvent represents when a warrior destroys an enemy
type EnemyDestroyedEvent struct {
    Event
    EnemyID        string `json:"enemy_id"`
    EnemyType      string `json:"enemy_type"`
    EnemyName      string `json:"enemy_name"`
    EnemyLevel     int    `json:"enemy_level"`
    EnemyHealth    int    `json:"enemy_health"`
    EnemyAttack    int    `json:"enemy_attack_power"`
    KillerWarriorID   uint   `json:"killer_warrior_id"`
    KillerWarriorName string `json:"killer_warrior_name"`
}

// NewEnemyDestroyedEvent creates a new enemy destroyed event
func NewEnemyDestroyedEvent(enemyID, enemyType, enemyName string, enemyLevel, enemyHealth, enemyAttack int, killerWarriorName string, killerWarriorID uint) *EnemyDestroyedEvent {
    return &EnemyDestroyedEvent{
        Event: Event{
            EventType:     "enemy_destroyed",
            Timestamp:     time.Now(),
            SourceService: "enemy",
        },
        EnemyID:            enemyID,
        EnemyType:          enemyType,
        EnemyName:          enemyName,
        EnemyLevel:         enemyLevel,
        EnemyHealth:        enemyHealth,
        EnemyAttack:        enemyAttack,
        KillerWarriorID:    killerWarriorID,
        KillerWarriorName:  killerWarriorName,
    }
}

// ArmorPurchaseEvent represents an armor purchase event
type ArmorPurchaseEvent struct {
	Event
	ArmorID      string `json:"armor_id"`
	BuyerID      uint   `json:"buyer_id"`
	BuyerName    string `json:"buyer_name"`
	ArmorName    string `json:"armor_name"`
	ArmorPrice   int    `json:"armor_price"`
	OwnerType    string `json:"owner_type"` // "warrior" | "enemy" | "dragon"
}

// NewArmorPurchaseEvent creates a new armor purchase event
func NewArmorPurchaseEvent(armorID, buyerName, armorName string, buyerID, armorPrice int, ownerType string) *ArmorPurchaseEvent {
	return &ArmorPurchaseEvent{
		Event: Event{
			EventType:     "armor_purchased",
			Timestamp:     time.Now(),
			SourceService: "armor",
		},
		ArmorID:    armorID,
		BuyerID:    uint(buyerID),
		BuyerName:  buyerName,
		ArmorName:  armorName,
		ArmorPrice: armorPrice,
		OwnerType:  ownerType,
	}
}

// Topic names
const (
	TopicWeaponPurchase = "weapon.purchase"
	TopicArmorPurchase  = "armor.purchase"
	TopicCoinDeduct     = "coin.deduct"
	TopicDragonDeath    = "dragon.death"
	TopicDragonRevival  = "dragon.revival"
	TopicEnemyDestroyed = "enemy.destroyed"
	TopicBattleStarted  = "battle.started"
	TopicBattleCompleted = "battle.completed"
    TopicBattleWagerResolved = "battle.wager.resolved"
)

// BattleWagerResolvedEvent represents wager payout after a team battle
type BattleWagerResolvedEvent struct {
    Event
    BattleID    string `json:"battle_id"`
    WinnerSide  string `json:"winner_side"`
    WagerAmount int    `json:"wager_amount"`
    LightEmperorID string `json:"light_emperor_id"`
    DarkEmperorID  string `json:"dark_emperor_id"`
}

func NewBattleWagerResolvedEvent(battleID, winnerSide string, wagerAmount int, lightEmperorID, darkEmperorID string) *BattleWagerResolvedEvent {
    return &BattleWagerResolvedEvent{
        Event: Event{ EventType: "battle_wager_resolved", Timestamp: time.Now(), SourceService: "battle" },
        BattleID: battleID,
        WinnerSide: winnerSide,
        WagerAmount: wagerAmount,
        LightEmperorID: lightEmperorID,
        DarkEmperorID: darkEmperorID,
    }
}

