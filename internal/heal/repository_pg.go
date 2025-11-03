package heal

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type sqlRepo struct{}

func getGorm() (*gorm.DB, error) {
	if !SQLDB.Enabled {
		return nil, errors.New("sql not enabled")
	}
	db, ok := SQLDB.DB.(*gorm.DB)
	if !ok || db == nil {
		return nil, errors.New("invalid sql handle")
	}
	return db, nil
}

func (r *sqlRepo) SaveHealingRecord(ctx context.Context, record *HealingRecord) error {
	db, err := getGorm()
	if err != nil {
		return err
	}

	row := &HealingRecordSQL{
		WarriorID:    record.WarriorID,
		WarriorName:  record.WarriorName,
		HealType:     string(record.HealType),
		HealedAmount:  record.HealedAmount,
		HPBefore:     record.HPBefore,
		HPAfter:      record.HPAfter,
		CoinsSpent:   record.CoinsSpent,
		Duration:     record.Duration,
		CompletedAt:  record.CompletedAt,
		CreatedAt:    record.CreatedAt,
	}

	if err := db.WithContext(ctx).Create(row).Error; err != nil {
		return fmt.Errorf("failed to save healing record: %w", err)
	}

	// Update record ID
	record.ID = fmt.Sprintf("%d", row.ID)
	return nil
}

func (r *sqlRepo) GetHealingHistory(ctx context.Context, warriorID uint) ([]*HealingRecord, error) {
	db, err := getGorm()
	if err != nil {
		return nil, err
	}

	var rows []HealingRecordSQL
	if err := db.WithContext(ctx).Where("warrior_id = ?", warriorID).Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to get healing history: %w", err)
	}

	records := make([]*HealingRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, &HealingRecord{
			ID:           fmt.Sprintf("%d", row.ID),
			WarriorID:    row.WarriorID,
			WarriorName:  row.WarriorName,
			HealType:     HealType(row.HealType),
			HealedAmount: row.HealedAmount,
			HPBefore:     row.HPBefore,
			HPAfter:      row.HPAfter,
			CoinsSpent:   row.CoinsSpent,
			Duration:     row.Duration,
			CompletedAt:  row.CompletedAt,
			CreatedAt:    row.CreatedAt,
		})
	}

	return records, nil
}

