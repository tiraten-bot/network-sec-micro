package battle

import (
	"context"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ParticipantKillTracker tracks which participants have damaged a target before it dies
type ParticipantKillTracker struct {
	mu           sync.RWMutex
	battleKills  map[primitive.ObjectID]map[string][]string // battleID -> targetID -> []attackerIDs
}

var killTracker = &ParticipantKillTracker{
	battleKills: make(map[primitive.ObjectID]map[string][]string),
}

// AddDamage records damage dealt to a target by an attacker
func (pkt *ParticipantKillTracker) AddDamage(battleID primitive.ObjectID, targetID, attackerID string) {
	pkt.mu.Lock()
	defer pkt.mu.Unlock()

	if pkt.battleKills[battleID] == nil {
		pkt.battleKills[battleID] = make(map[string][]string)
	}

	// Add attacker if not already in list
	attackers := pkt.battleKills[battleID][targetID]
	found := false
	for _, id := range attackers {
		if id == attackerID {
			found = true
			break
		}
	}
	if !found {
		pkt.battleKills[battleID][targetID] = append(attackers, attackerID)
	}
}

// GetKillers returns all participants who damaged a target before it died
func (pkt *ParticipantKillTracker) GetKillers(battleID primitive.ObjectID, targetID string) []string {
	pkt.mu.RLock()
	defer pkt.mu.RUnlock()

	if pkt.battleKills[battleID] == nil {
		return []string{}
	}

	return pkt.battleKills[battleID][targetID]
}

// ClearKills clears kill tracking for a battle (after it ends)
func (pkt *ParticipantKillTracker) ClearKills(battleID primitive.ObjectID) {
	pkt.mu.Lock()
	defer pkt.mu.Unlock()

	delete(pkt.battleKills, battleID)
}

// GetParticipantObjects returns BattleParticipant objects from IDs
func GetParticipantObjects(ctx context.Context, battleID primitive.ObjectID, participantIDs []string) ([]*BattleParticipant, error) {
	participants := make([]*BattleParticipant, 0, len(participantIDs))

	for _, pid := range participantIDs {
		var p BattleParticipant
		err := BattleParticipantColl.FindOne(ctx, bson.M{
			"battle_id":      battleID,
			"participant_id": pid,
		}).Decode(&p)
		if err == nil {
			participants = append(participants, &p)
		}
	}

	return participants, nil
}

