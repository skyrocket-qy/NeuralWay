package systems

import (
	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// TransactionEvent represents a currency transaction.
type TransactionEvent struct {
	Entity   ecs.Entity
	Type     TransactionType
	Currency components.CurrencyType
	Amount   int64
	ItemID   string
	Success  bool
}

// TransactionType represents the type of transaction.
type TransactionType string

const (
	TransactionPurchase TransactionType = "purchase"
	TransactionSell     TransactionType = "sell"
	TransactionReward   TransactionType = "reward"
	TransactionSpend    TransactionType = "spend"
)

// EconomySystem manages currency and transactions.
type EconomySystem struct {
	currencyFilter *ecs.Filter1[components.Currency]
	scoreFilter    *ecs.Filter1[components.Score]
	upgradeFilter  *ecs.Filter1[components.UpgradeManager]
	shops          map[string]*components.Shop
	transactionLog []TransactionEvent
	onTransaction  func(TransactionEvent)
}

// NewEconomySystem creates an economy system.
func NewEconomySystem(world *ecs.World) *EconomySystem {
	return &EconomySystem{
		currencyFilter: ecs.NewFilter1[components.Currency](world),
		scoreFilter:    ecs.NewFilter1[components.Score](world),
		upgradeFilter:  ecs.NewFilter1[components.UpgradeManager](world),
		shops:          make(map[string]*components.Shop),
		transactionLog: make([]TransactionEvent, 0),
	}
}

// SetOnTransaction sets the transaction callback.
func (s *EconomySystem) SetOnTransaction(fn func(TransactionEvent)) {
	s.onTransaction = fn
}

// RegisterShop registers a shop.
func (s *EconomySystem) RegisterShop(shop components.Shop) {
	s.shops[shop.ID] = &shop
}

// GetShop returns a shop by ID.
func (s *EconomySystem) GetShop(id string) *components.Shop {
	return s.shops[id]
}

// AddCurrency adds currency to an entity.
func (s *EconomySystem) AddCurrency(
	world *ecs.World,
	entity ecs.Entity,
	currencyType components.CurrencyType,
	amount int64,
) bool {
	query := s.currencyFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			currency := query.Get()
			currency.Add(currencyType, amount)

			event := TransactionEvent{
				Entity:   entity,
				Type:     TransactionReward,
				Currency: currencyType,
				Amount:   amount,
				Success:  true,
			}

			s.transactionLog = append(s.transactionLog, event)
			if s.onTransaction != nil {
				s.onTransaction(event)
			}

			return true
		}
	}

	return false
}

// RemoveCurrency removes currency from an entity.
func (s *EconomySystem) RemoveCurrency(
	world *ecs.World,
	entity ecs.Entity,
	currencyType components.CurrencyType,
	amount int64,
) bool {
	query := s.currencyFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			currency := query.Get()
			if !currency.Remove(currencyType, amount) {
				return false
			}

			event := TransactionEvent{
				Entity:   entity,
				Type:     TransactionSpend,
				Currency: currencyType,
				Amount:   amount,
				Success:  true,
			}

			s.transactionLog = append(s.transactionLog, event)
			if s.onTransaction != nil {
				s.onTransaction(event)
			}

			return true
		}
	}

	return false
}

// GetCurrency returns currency amount for an entity.
func (s *EconomySystem) GetCurrency(
	world *ecs.World,
	entity ecs.Entity,
	currencyType components.CurrencyType,
) int64 {
	query := s.currencyFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			currency := query.Get()

			return currency.Get(currencyType)
		}
	}

	return 0
}

// Purchase attempts to buy an item from a shop.
func (s *EconomySystem) Purchase(
	world *ecs.World,
	entity ecs.Entity,
	shopID, itemID string,
	playerLevel int,
) bool {
	shop := s.shops[shopID]
	if shop == nil {
		return false
	}

	item, ok := shop.Items[itemID]
	if !ok {
		return false
	}

	// Get player currency
	var currency *components.Currency

	query := s.currencyFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			currency = query.Get()

			break
		}
	}

	if currency == nil {
		return false
	}

	if !item.CanPurchase(currency, playerLevel) {
		event := TransactionEvent{
			Entity:   entity,
			Type:     TransactionPurchase,
			Currency: item.Currency,
			Amount:   item.GetFinalPrice(),
			ItemID:   itemID,
			Success:  false,
		}
		s.transactionLog = append(s.transactionLog, event)

		return false
	}

	item.Purchase(currency)

	event := TransactionEvent{
		Entity:   entity,
		Type:     TransactionPurchase,
		Currency: item.Currency,
		Amount:   item.GetFinalPrice(),
		ItemID:   itemID,
		Success:  true,
	}

	s.transactionLog = append(s.transactionLog, event)
	if s.onTransaction != nil {
		s.onTransaction(event)
	}

	return true
}

// Sell sells an item for currency.
func (s *EconomySystem) Sell(
	world *ecs.World,
	entity ecs.Entity,
	item *components.Item,
	sellMultiplier float64,
) bool {
	value := int64(float64(item.Value) * sellMultiplier)
	if value <= 0 {
		return false
	}

	return s.AddCurrency(world, entity, components.CurrencyGold, value)
}

// PurchaseUpgrade attempts to purchase an upgrade.
func (s *EconomySystem) PurchaseUpgrade(world *ecs.World, entity ecs.Entity, upgradeID string) bool {
	var (
		currency *components.Currency
		upgrades *components.UpgradeManager
	)

	currQuery := s.currencyFilter.Query()
	for currQuery.Next() {
		e := currQuery.Entity()
		if e == entity {
			currency = currQuery.Get()

			break
		}
	}

	upgradeQuery := s.upgradeFilter.Query()
	for upgradeQuery.Next() {
		e := upgradeQuery.Entity()
		if e == entity {
			upgrades = upgradeQuery.Get()

			break
		}
	}

	if currency == nil || upgrades == nil {
		return false
	}

	upgrade, ok := upgrades.Upgrades[upgradeID]
	if !ok {
		return false
	}

	if !upgrade.CanUpgrade(currency) {
		return false
	}

	return upgrade.DoUpgrade(currency)
}

// AddScore adds to an entity's score.
func (s *EconomySystem) AddScore(world *ecs.World, entity ecs.Entity, units int64) int64 {
	query := s.scoreFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			score := query.Get()

			return score.AddScore(units)
		}
	}

	return 0
}

// GetScore returns an entity's current and best score.
func (s *EconomySystem) GetScore(world *ecs.World, entity ecs.Entity) (current, best int64) {
	query := s.scoreFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			score := query.Get()

			return score.Current, score.Best
		}
	}

	return 0, 0
}

// BreakStreak breaks an entity's score streak.
func (s *EconomySystem) BreakStreak(world *ecs.World, entity ecs.Entity) {
	query := s.scoreFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			score := query.Get()
			score.BreakStreak()

			return
		}
	}
}

// SetScoreMultiplier sets the score multiplier for an entity.
func (s *EconomySystem) SetScoreMultiplier(world *ecs.World, entity ecs.Entity, multiplier float64) {
	query := s.scoreFilter.Query()
	for query.Next() {
		e := query.Entity()
		if e == entity {
			score := query.Get()
			score.Multiplier = multiplier

			return
		}
	}
}

// GetTransactionLog returns recent transactions.
func (s *EconomySystem) GetTransactionLog() []TransactionEvent {
	return s.transactionLog
}

// ClearTransactionLog clears the transaction log.
func (s *EconomySystem) ClearTransactionLog() {
	s.transactionLog = s.transactionLog[:0]
}

// Update updates multiplier timers.
func (s *EconomySystem) Update(world *ecs.World, dt float64) {
	// Future: Update timed discounts, sales, etc.
}
