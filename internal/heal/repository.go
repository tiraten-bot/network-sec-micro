package heal

import (
	"context"
)

// Repository abstracts persistence for Heal domain
type Repository interface {
	SaveHealingRecord(ctx context.Context, record *HealingRecord) error
	GetHealingHistory(ctx context.Context, warriorID uint) ([]*HealingRecord, error)
}

var defaultRepo Repository

// GetRepository returns a singleton repo based on env
func GetRepository() Repository {
	if defaultRepo != nil {
		return defaultRepo
	}
	if SQLDB.Enabled {
		defaultRepo = &sqlRepo{}
	} else {
		defaultRepo = &memoryRepo{} // Fallback to in-memory
	}
	return defaultRepo
}

// memoryRepo is a simple in-memory repository (fallback)
type memoryRepo struct {
	records []*HealingRecord
}

func (r *memoryRepo) SaveHealingRecord(ctx context.Context, record *HealingRecord) error {
	r.records = append(r.records, record)
	return nil
}

func (r *memoryRepo) GetHealingHistory(ctx context.Context, warriorID uint) ([]*HealingRecord, error) {
	var result []*HealingRecord
	for _, rec := range r.records {
		if rec.WarriorID == warriorID {
			result = append(result, rec)
		}
	}
	return result, nil
}

