package components

// CurrencyType represents different currency types.
type CurrencyType string

const (
	CurrencyGold     CurrencyType = "gold"
	CurrencyGems     CurrencyType = "gems"
	CurrencyTokens   CurrencyType = "tokens"
	CurrencyPrestige CurrencyType = "prestige"
)

// Currency represents multiple currency pools.
type Currency struct {
	Amounts map[CurrencyType]int64
}

// NewCurrency creates an empty currency wallet.
func NewCurrency() Currency {
	return Currency{
		Amounts: make(map[CurrencyType]int64),
	}
}

// Add adds currency of a specific type.
func (c *Currency) Add(currencyType CurrencyType, amount int64) {
	c.Amounts[currencyType] += amount
}

// Remove removes currency, returns false if insufficient.
func (c *Currency) Remove(currencyType CurrencyType, amount int64) bool {
	if c.Amounts[currencyType] < amount {
		return false
	}

	c.Amounts[currencyType] -= amount

	return true
}

// Get returns the amount of a currency type.
func (c *Currency) Get(currencyType CurrencyType) int64 {
	return c.Amounts[currencyType]
}

// Has returns true if there's at least the specified amount.
func (c *Currency) Has(currencyType CurrencyType, amount int64) bool {
	return c.Amounts[currencyType] >= amount
}

// Score represents a game score with multipliers.
type Score struct {
	Current     int64
	Best        int64   // High score
	Multiplier  float64 // Current score multiplier
	BasePerUnit int64   // Base points per scoring action
	Streak      int     // Current streak count
	StreakBonus float64 // Bonus per streak
	StreakMax   int     // Maximum streak for bonus cap
}

// NewScore creates a score tracker.
func NewScore(basePerUnit int64) Score {
	return Score{
		BasePerUnit: basePerUnit,
		Multiplier:  1.0,
		StreakBonus: 0.1,
	}
}

// AddScore adds points with multipliers applied.
func (s *Score) AddScore(units int64) int64 {
	streakMult := 1.0 + (s.StreakBonus * float64(min(s.Streak, s.StreakMax)))
	points := int64(float64(s.BasePerUnit*units) * s.Multiplier * streakMult)

	s.Current += points
	if s.Current > s.Best {
		s.Best = s.Current
	}

	s.Streak++

	return points
}

// BreakStreak resets the streak.
func (s *Score) BreakStreak() {
	s.Streak = 0
}

// Reset resets the current score but keeps best.
func (s *Score) Reset() {
	s.Current = 0
	s.Streak = 0
	s.Multiplier = 1.0
}

// ShopItem represents an item available for purchase.
type ShopItem struct {
	ID            string
	Name          string
	Description   string
	ItemID        string // ID of item to give on purchase
	Price         int64
	Currency      CurrencyType
	Stock         int     // -1 for unlimited
	MaxPurchases  int     // Per-player limit, -1 for unlimited
	Purchased     int     // Times purchased by player
	Discount      float64 // 0.0-1.0 discount percentage
	Available     bool    // Currently available
	RequiredLevel int     // Level required to see/buy
}

// NewShopItem creates a shop item.
func NewShopItem(id, name, itemID string, price int64, currency CurrencyType) ShopItem {
	return ShopItem{
		ID:           id,
		Name:         name,
		ItemID:       itemID,
		Price:        price,
		Currency:     currency,
		Stock:        -1,
		MaxPurchases: -1,
		Available:    true,
	}
}

// GetFinalPrice returns the price after discount.
func (si *ShopItem) GetFinalPrice() int64 {
	return int64(float64(si.Price) * (1.0 - si.Discount))
}

// CanPurchase checks if the item can be bought.
func (si *ShopItem) CanPurchase(playerCurrency *Currency, playerLevel int) bool {
	if !si.Available {
		return false
	}

	if si.RequiredLevel > 0 && playerLevel < si.RequiredLevel {
		return false
	}

	if si.Stock == 0 {
		return false
	}

	if si.MaxPurchases > 0 && si.Purchased >= si.MaxPurchases {
		return false
	}

	return playerCurrency.Has(si.Currency, si.GetFinalPrice())
}

// Purchase processes a purchase, returns true if successful.
func (si *ShopItem) Purchase(playerCurrency *Currency) bool {
	price := si.GetFinalPrice()
	if !playerCurrency.Remove(si.Currency, price) {
		return false
	}

	if si.Stock > 0 {
		si.Stock--
	}

	si.Purchased++

	return true
}

// Shop represents a collection of purchasable items.
type Shop struct {
	ID    string
	Name  string
	Items map[string]*ShopItem
}

// NewShop creates an empty shop.
func NewShop(id, name string) Shop {
	return Shop{
		ID:    id,
		Name:  name,
		Items: make(map[string]*ShopItem),
	}
}

// AddItem adds an item to the shop.
func (s *Shop) AddItem(item ShopItem) {
	s.Items[item.ID] = &item
}

// Upgrade represents a permanent upgrade.
type Upgrade struct {
	ID           string
	Name         string
	Description  string
	CurrentLevel int
	MaxLevel     int
	BaseCost     int64
	CostScaling  float64 // Cost multiplier per level
	Currency     CurrencyType
	Effects      map[string]float64 // Effect per level
}

// NewUpgrade creates an upgrade.
func NewUpgrade(id, name string, maxLevel int, baseCost int64, currency CurrencyType) Upgrade {
	return Upgrade{
		ID:          id,
		Name:        name,
		MaxLevel:    maxLevel,
		BaseCost:    baseCost,
		CostScaling: 1.5,
		Currency:    currency,
		Effects:     make(map[string]float64),
	}
}

// GetCost returns the cost for the next level.
func (u *Upgrade) GetCost() int64 {
	if u.CurrentLevel >= u.MaxLevel {
		return 0
	}

	multiplier := 1.0
	for i := 0; i < u.CurrentLevel; i++ {
		multiplier *= u.CostScaling
	}

	return int64(float64(u.BaseCost) * multiplier)
}

// CanUpgrade checks if upgrade is possible.
func (u *Upgrade) CanUpgrade(currency *Currency) bool {
	if u.CurrentLevel >= u.MaxLevel {
		return false
	}

	return currency.Has(u.Currency, u.GetCost())
}

// DoUpgrade performs the upgrade, returns true if successful.
func (u *Upgrade) DoUpgrade(currency *Currency) bool {
	cost := u.GetCost()
	if !currency.Remove(u.Currency, cost) {
		return false
	}

	u.CurrentLevel++

	return true
}

// GetTotalEffect returns the cumulative effect value.
func (u *Upgrade) GetTotalEffect(stat string) float64 {
	if effect, ok := u.Effects[stat]; ok {
		return effect * float64(u.CurrentLevel)
	}

	return 0
}

// UpgradeManager holds all upgrades for an entity.
type UpgradeManager struct {
	Upgrades map[string]*Upgrade
}

// NewUpgradeManager creates an upgrade manager.
func NewUpgradeManager() UpgradeManager {
	return UpgradeManager{
		Upgrades: make(map[string]*Upgrade),
	}
}

// Add registers an upgrade.
func (um *UpgradeManager) Add(upgrade Upgrade) {
	um.Upgrades[upgrade.ID] = &upgrade
}

// GetAllEffects returns combined effects from all upgrades.
func (um *UpgradeManager) GetAllEffects() map[string]float64 {
	effects := make(map[string]float64)

	for _, upgrade := range um.Upgrades {
		for stat := range upgrade.Effects {
			effects[stat] += upgrade.GetTotalEffect(stat)
		}
	}

	return effects
}

// Multiplier represents a temporary or permanent multiplier.
type Multiplier struct {
	ID        string
	StatType  string
	Value     float64
	Duration  float64 // 0 = permanent
	Remaining float64
	Additive  bool // If true, adds to base; if false, multiplies
}

// NewMultiplier creates a multiplier.
func NewMultiplier(id, statType string, value, duration float64) Multiplier {
	return Multiplier{
		ID:        id,
		StatType:  statType,
		Value:     value,
		Duration:  duration,
		Remaining: duration,
	}
}

// IsActive returns true if the multiplier is still active.
func (m *Multiplier) IsActive() bool {
	return m.Duration == 0 || m.Remaining > 0
}

// MultiplierManager holds active multipliers.
type MultiplierManager struct {
	Multipliers []Multiplier
}

// NewMultiplierManager creates a multiplier manager.
func NewMultiplierManager() MultiplierManager {
	return MultiplierManager{
		Multipliers: make([]Multiplier, 0),
	}
}

// Add adds a multiplier.
func (mm *MultiplierManager) Add(mult Multiplier) {
	mm.Multipliers = append(mm.Multipliers, mult)
}

// GetTotal returns the total multiplier for a stat type.
func (mm *MultiplierManager) GetTotal(statType string) (additive, multiplicative float64) {
	multiplicative = 1.0

	for _, m := range mm.Multipliers {
		if m.StatType == statType && m.IsActive() {
			if m.Additive {
				additive += m.Value
			} else {
				multiplicative *= m.Value
			}
		}
	}

	return additive, multiplicative
}

// RemoveExpired removes inactive multipliers.
func (mm *MultiplierManager) RemoveExpired() {
	active := mm.Multipliers[:0]
	for _, m := range mm.Multipliers {
		if m.IsActive() {
			active = append(active, m)
		}
	}

	mm.Multipliers = active
}
