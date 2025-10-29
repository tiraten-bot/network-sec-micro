package dto

type CreateEnemyCommand struct {
	Name        string
	Type        string
	Level       int
	Health      int
	AttackPower int
	CreatedBy   string
}

type AttackWarriorCommand struct {
	EnemyID     string
	WarriorID   uint
	WarriorName string
	Amount      int
	WeaponID    string
}

type DestroyEnemyCommand struct {
    EnemyID           string
    KillerWarriorID   uint
    KillerWarriorName string
}
