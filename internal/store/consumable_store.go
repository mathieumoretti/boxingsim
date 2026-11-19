package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/mormm/boxing/internal/domain"
	"github.com/mormm/boxing/internal/model"
)

var (
	ErrConsumableNotFound   = errors.New("consumable not found")
	ErrInsufficientQuantity = errors.New("insufficient consumable quantity")
	ErrConsumableOnCooldown = errors.New("consumable is on cooldown")
)

// ConsumableStore implements consumable data access
type ConsumableStore struct {
	db *sql.DB
}

func NewConsumableStore(db *sql.DB) *ConsumableStore {
	return &ConsumableStore{db: db}
}

// GetAll returns all consumables
func (s *ConsumableStore) GetAll(ctx context.Context) ([]*model.Consumable, error) {
	query := `
		SELECT id, name, description, consumable_type, energy_restore, health_restore,
		       fatigue_reduction, cooldown_hours, is_unlimited, created_at, updated_at
		FROM consumables ORDER BY name`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var consumables []*model.Consumable
	for rows.Next() {
		c := &model.Consumable{}
		err := rows.Scan(
			&c.ID, &c.Name, &c.Description, &c.ConsumableType, &c.EnergyRestore,
			&c.HealthRestore, &c.FatigueReduction, &c.CooldownHours, &c.IsUnlimited,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		consumables = append(consumables, c)
	}
	return consumables, rows.Err()
}

// GetByID retrieves a consumable by ID
func (s *ConsumableStore) GetByID(ctx context.Context, id int) (*model.Consumable, error) {
	query := `
		SELECT id, name, description, consumable_type, energy_restore, health_restore,
		       fatigue_reduction, cooldown_hours, is_unlimited, created_at, updated_at
		FROM consumables WHERE id = $1`

	c := &model.Consumable{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.Description, &c.ConsumableType, &c.EnergyRestore,
		&c.HealthRestore, &c.FatigueReduction, &c.CooldownHours, &c.IsUnlimited,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrConsumableNotFound
		}
		return nil, err
	}
	return c, nil
}

// GetBoxerInventory returns all consumables for a boxer with quantities
func (s *ConsumableStore) GetBoxerInventory(ctx context.Context, boxerID domain.BoxerID) ([]*model.BoxerConsumable, error) {
	query := `
		SELECT bc.id, bc.boxer_id, bc.consumable_id, bc.quantity, bc.last_used_at,
		       bc.created_at, bc.updated_at
		FROM boxer_consumables bc
		WHERE bc.boxer_id = $1`

	rows, err := s.db.QueryContext(ctx, query, int(boxerID))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var inventory []*model.BoxerConsumable
	for rows.Next() {
		bc := &model.BoxerConsumable{}
		err := rows.Scan(
			&bc.ID, &bc.BoxerID, &bc.ConsumableID, &bc.Quantity,
			&bc.LastUsedAt, &bc.CreatedAt, &bc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		inventory = append(inventory, bc)
	}
	return inventory, rows.Err()
}

// GetBoxerConsumableQuantity returns quantity of a specific consumable for a boxer
func (s *ConsumableStore) GetBoxerConsumableQuantity(ctx context.Context, boxerID domain.BoxerID, consumableID int) (int, error) {
	query := `SELECT quantity FROM boxer_consumables WHERE boxer_id = $1 AND consumable_id = $2`
	var qty int
	err := s.db.QueryRowContext(ctx, query, int(boxerID), consumableID).Scan(&qty)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return qty, nil
}

// UseConsumable atomically decrements quantity and updates last_used_at
func (s *ConsumableStore) UseConsumable(ctx context.Context, boxerID domain.BoxerID, consumableID int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Check and update quantity
	query := `
		UPDATE boxer_consumables
		SET quantity = quantity - 1, last_used_at = $2, updated_at = $2
		WHERE boxer_id = $1 AND consumable_id = $3 AND quantity > 0
		RETURNING quantity`

	var newQty int
	err = tx.QueryRowContext(ctx, query, int(boxerID), time.Now(), consumableID).Scan(&newQty)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInsufficientQuantity
		}
		return err
	}

	// Delete row if quantity reaches 0
	if newQty == 0 {
		_, err = tx.ExecContext(ctx, `DELETE FROM boxer_consumables WHERE boxer_id = $1 AND consumable_id = $2`, int(boxerID), consumableID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// AddToInventory adds quantity to boxer's consumable inventory
func (s *ConsumableStore) AddToInventory(ctx context.Context, boxerID domain.BoxerID, consumableID int, quantity int) error {
	query := `
		INSERT INTO boxer_consumables (boxer_id, consumable_id, quantity, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (boxer_id, consumable_id)
		DO UPDATE SET quantity = boxer_consumables.quantity + $3, updated_at = NOW()`

	_, err := s.db.ExecContext(ctx, query, int(boxerID), consumableID, quantity)
	return err
}

// CheckCooldown returns true if consumable is on cooldown for this boxer
func (s *ConsumableStore) CheckCooldown(ctx context.Context, boxerID domain.BoxerID, consumable *model.Consumable) (bool, *time.Time, error) {
	if consumable.CooldownHours == 0 {
		return false, nil, nil // No cooldown
	}

	query := `SELECT last_used_at FROM boxer_consumables WHERE boxer_id = $1 AND consumable_id = $2`
	var lastUsedAt *time.Time
	err := s.db.QueryRowContext(ctx, query, int(boxerID), consumable.ID).Scan(&lastUsedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil, nil // Never used, no cooldown
		}
		return false, nil, err
	}

	if lastUsedAt == nil {
		return false, nil, nil
	}

	cooldownEnd := lastUsedAt.Add(time.Duration(consumable.CooldownHours) * time.Hour)
	if time.Now().Before(cooldownEnd) {
		return true, &cooldownEnd, nil
	}
	return false, nil, nil
}
