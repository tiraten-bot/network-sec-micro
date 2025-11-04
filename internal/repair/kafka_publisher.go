package repair

import (
    "context"
    "encoding/json"
    "log"
    "os"

    "github.com/IBM/sarama"
)

func PublishRepairEvent(ctx context.Context, ownerType, ownerID string, cost int, weaponID, orderID string) error {
    brokers := os.Getenv("KAFKA_BROKERS")
    if brokers == "" { brokers = "localhost:9092" }
    cfg := sarama.NewConfig()
    cfg.Producer.Return.Successes = true
    prod, err := sarama.NewSyncProducer([]string{brokers}, cfg)
    if err != nil { return err }
    defer prod.Close()

    evt := map[string]interface{}{
        "type": "weapon.repair",
        "owner_type": ownerType,
        "owner_id": ownerID,
        "cost": cost,
        "weapon_id": weaponID,
        "order_id": orderID,
    }
    payload, _ := json.Marshal(evt)
    msg := &sarama.ProducerMessage{Topic: "weapon.repair", Value: sarama.ByteEncoder(payload)}
    _, _, err = prod.SendMessage(msg)
    if err != nil { return err }
    log.Printf("Published weapon.repair event for order=%s weapon=%s cost=%d", orderID, weaponID, cost)
    return nil
}

func PublishArmorRepairEvent(ctx context.Context, ownerType, ownerID string, cost int, armorID, orderID string) error {
    brokers := os.Getenv("KAFKA_BROKERS")
    if brokers == "" { brokers = "localhost:9092" }
    cfg := sarama.NewConfig()
    cfg.Producer.Return.Successes = true
    prod, err := sarama.NewSyncProducer([]string{brokers}, cfg)
    if err != nil { return err }
    defer prod.Close()

    evt := map[string]interface{}{
        "type": "armor.repair",
        "owner_type": ownerType,
        "owner_id": ownerID,
        "cost": cost,
        "armor_id": armorID,
        "order_id": orderID,
    }
    payload, _ := json.Marshal(evt)
    msg := &sarama.ProducerMessage{Topic: "armor.repair", Value: sarama.ByteEncoder(payload)}
    _, _, err = prod.SendMessage(msg)
    if err != nil { return err }
    log.Printf("Published armor.repair event for order=%s armor=%s cost=%d", orderID, armorID, cost)
    return nil
}


