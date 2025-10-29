package warrior

import (
    "encoding/json"
    "log"
    "strings"
)

// ProcessKafkaMessage handles incoming Kafka messages for warrior achievements
func ProcessKafkaMessage(message []byte) error {
    // Try to detect event type without strict schema dependency
    var base map[string]interface{}
    if err := json.Unmarshal(message, &base); err != nil {
        log.Printf("warrior: failed to unmarshal kafka message: %v", err)
        return err
    }

    eventType := ""
    if v, ok := base["event_type"].(string); ok {
        eventType = v
    }
    source := ""
    if v, ok := base["source_service"].(string); ok {
        source = v
    }

    // Handle dragon death events
    if strings.EqualFold(eventType, "dragon_death") || strings.EqualFold(source, "dragon") {
        return handleDragonDeath(base)
    }

    // Handle enemy destroyed events
    if strings.EqualFold(eventType, "enemy_destroyed") || strings.EqualFold(source, "enemy") {
        // Distinguish from enemy attack by presence of fields
        if _, has := base["killer_warrior_id"]; has {
            return handleEnemyDestroyed(base)
        }
    }

    return nil
}

func handleDragonDeath(base map[string]interface{}) error {
    killer, _ := base["killer_username"].(string)
    if killer == "" {
        return nil
    }

    // Increment dragon kill count for killer
    var w Warrior
    if err := DB.Where("username = ?", killer).First(&w).Error; err != nil {
        return nil
    }

    w.DragonKillCount += 1
    // Title precedence: EmperorOfDrags (>=10) > DragonSlayer (>=3)
    if w.DragonKillCount >= 10 {
        w.Title = "EmperorOfDrags"
    } else if w.DragonKillCount >= 3 && w.Title != "EmperorOfDrags" {
        w.Title = "DragonSlayer"
    }

    if err := DB.Save(&w).Error; err != nil {
        log.Printf("warrior: failed to update dragon kill count: %v", err)
        return err
    }
    return nil
}

func handleEnemyDestroyed(base map[string]interface{}) error {
    // Prefer warrior ID when available
    var warriorIDFloat float64
    if v, ok := base["killer_warrior_id"].(float64); ok {
        warriorIDFloat = v
    }
    killerName, _ := base["killer_warrior_name"].(string)

    var w Warrior
    if warriorIDFloat > 0 {
        if err := DB.First(&w, uint(warriorIDFloat)).Error; err != nil {
            return nil
        }
    } else if killerName != "" {
        if err := DB.Where("username = ?", killerName).First(&w).Error; err != nil {
            return nil
        }
    } else {
        return nil
    }

    w.EnemyKillCount += 1
    // Set EnemyDestroyer if at least 100, unless a higher title already assigned
    if w.EnemyKillCount >= 100 && w.Title != "EmperorOfDrags" {
        // Keep DragonSlayer/EmperorOfDrags precedence
        if w.DragonKillCount < 10 { // do not override EmperorOfDrags
            w.Title = "EnemyDestroyer"
        }
    }

    if err := DB.Save(&w).Error; err != nil {
        log.Printf("warrior: failed to update enemy kill count: %v", err)
        return err
    }
    return nil
}


