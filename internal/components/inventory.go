package components

// ItemRarity represents the rarity tier of an item.
type ItemRarity int

const (
	RarityCommon ItemRarity = iota
	RarityUncommon
	RarityRare
	RarityEpic
	RarityLegendary
	RarityUnique
)

// String returns the rarity name.
func (r ItemRarity) String() string {
	names := []string{"Common", "Uncommon", "Rare", "Epic", "Legendary", "Unique"}
	if int(r) < len(names) {
		return names[r]
	}

	return "Unknown"
}

// Color returns a hex color for the rarity.
func (r ItemRarity) Color() uint32 {
	colors := []uint32{
		0xFFFFFF, // Common - White
		0x1EFF00, // Uncommon - Green
		0x0070DD, // Rare - Blue
		0xA335EE, // Epic - Purple
		0xFF8000, // Legendary - Orange
		0xE6CC80, // Unique - Gold
	}
	if int(r) < len(colors) {
		return colors[r]
	}

	return 0xFFFFFF
}

// Item represents a game item.
type Item struct {
	ID          string
	Name        string
	Description string
	Rarity      ItemRarity
	StackSize   int                // Max stack size (1 = non-stackable)
	Count       int                // Current stack count
	Value       int64              // Base sell value
	Stats       map[string]float64 // Stat modifiers
	Tags        []string           // Item tags for filtering
	IconID      string             // Asset ID for icon
}

// NewItem creates a new item.
func NewItem(id, name string, rarity ItemRarity) Item {
	return Item{
		ID:        id,
		Name:      name,
		Rarity:    rarity,
		StackSize: 1,
		Count:     1,
		Stats:     make(map[string]float64),
		Tags:      make([]string, 0),
	}
}

// CanStack returns true if this item can stack with another.
func (i *Item) CanStack(other *Item) bool {
	return i.ID == other.ID && i.StackSize > 1 && i.Count < i.StackSize
}

// AddToStack adds items to the stack, returns overflow.
func (i *Item) AddToStack(amount int) int {
	space := i.StackSize - i.Count
	if amount <= space {
		i.Count += amount

		return 0
	}

	i.Count = i.StackSize

	return amount - space
}

// InventorySlot represents a slot in an inventory.
type InventorySlot struct {
	Item   *Item
	Locked bool // Slot cannot be modified
}

// Inventory represents a container of item slots.
type Inventory struct {
	Slots    []InventorySlot
	MaxSlots int
	Gold     int64 // Built-in currency
}

// NewInventory creates an inventory with the given slot count.
func NewInventory(slots int) Inventory {
	return Inventory{
		Slots:    make([]InventorySlot, slots),
		MaxSlots: slots,
	}
}

// AddItem adds an item to the inventory, returns false if full.
func (inv *Inventory) AddItem(item Item) bool {
	// Try to stack first
	for i := range inv.Slots {
		if inv.Slots[i].Item != nil && inv.Slots[i].Item.CanStack(&item) {
			overflow := inv.Slots[i].Item.AddToStack(item.Count)
			if overflow == 0 {
				return true
			}

			item.Count = overflow
		}
	}

	// Find empty slot
	for i := range inv.Slots {
		if inv.Slots[i].Item == nil && !inv.Slots[i].Locked {
			inv.Slots[i].Item = &item

			return true
		}
	}

	return false
}

// RemoveItem removes an item by ID, returns the removed item.
func (inv *Inventory) RemoveItem(id string, count int) *Item {
	for i := range inv.Slots {
		if inv.Slots[i].Item != nil && inv.Slots[i].Item.ID == id {
			if inv.Slots[i].Item.Count <= count {
				item := inv.Slots[i].Item
				inv.Slots[i].Item = nil

				return item
			}
			// Partial removal
			removed := *inv.Slots[i].Item
			removed.Count = count
			inv.Slots[i].Item.Count -= count

			return &removed
		}
	}

	return nil
}

// GetItem returns an item by ID if it exists.
func (inv *Inventory) GetItem(id string) *Item {
	for i := range inv.Slots {
		if inv.Slots[i].Item != nil && inv.Slots[i].Item.ID == id {
			return inv.Slots[i].Item
		}
	}

	return nil
}

// GetItemCount returns total count of an item by ID.
func (inv *Inventory) GetItemCount(id string) int {
	total := 0

	for i := range inv.Slots {
		if inv.Slots[i].Item != nil && inv.Slots[i].Item.ID == id {
			total += inv.Slots[i].Item.Count
		}
	}

	return total
}

// IsFull returns true if no empty slots remain.
func (inv *Inventory) IsFull() bool {
	for i := range inv.Slots {
		if inv.Slots[i].Item == nil && !inv.Slots[i].Locked {
			return false
		}
	}

	return true
}

// EquipmentSlotType represents where equipment can be worn.
type EquipmentSlotType string

const (
	SlotHead      EquipmentSlotType = "head"
	SlotChest     EquipmentSlotType = "chest"
	SlotLegs      EquipmentSlotType = "legs"
	SlotFeet      EquipmentSlotType = "feet"
	SlotHands     EquipmentSlotType = "hands"
	SlotMainHand  EquipmentSlotType = "main_hand"
	SlotOffHand   EquipmentSlotType = "off_hand"
	SlotRing1     EquipmentSlotType = "ring1"
	SlotRing2     EquipmentSlotType = "ring2"
	SlotAmulet    EquipmentSlotType = "amulet"
	SlotAccessory EquipmentSlotType = "accessory"
)

// Equipment represents equipped gear on an entity.
type Equipment struct {
	Slots map[EquipmentSlotType]*Item
}

// NewEquipment creates an empty equipment set.
func NewEquipment() Equipment {
	return Equipment{
		Slots: make(map[EquipmentSlotType]*Item),
	}
}

// Equip places an item in a slot, returns the previously equipped item.
func (e *Equipment) Equip(slot EquipmentSlotType, item *Item) *Item {
	previous := e.Slots[slot]
	e.Slots[slot] = item

	return previous
}

// Unequip removes and returns an item from a slot.
func (e *Equipment) Unequip(slot EquipmentSlotType) *Item {
	item := e.Slots[slot]
	e.Slots[slot] = nil

	return item
}

// GetTotalStats returns combined stats from all equipped items.
func (e *Equipment) GetTotalStats() map[string]float64 {
	stats := make(map[string]float64)

	for _, item := range e.Slots {
		if item != nil {
			for stat, value := range item.Stats {
				stats[stat] += value
			}
		}
	}

	return stats
}

// Consumable represents a one-time use item.
type Consumable struct {
	ItemID      string
	Uses        int                // Remaining uses
	MaxUses     int                // Maximum uses (1 for single-use)
	Cooldown    float64            // Cooldown between uses
	CooldownEnd float64            // Time when usable again
	Effects     map[string]float64 // Effects when consumed
	Duration    float64            // Duration of effects (0 = instant)
}

// NewConsumable creates a consumable.
func NewConsumable(itemID string, uses int) Consumable {
	return Consumable{
		ItemID:  itemID,
		Uses:    uses,
		MaxUses: uses,
		Effects: make(map[string]float64),
	}
}

// CanUse returns true if the consumable can be used.
func (c *Consumable) CanUse(currentTime float64) bool {
	return c.Uses > 0 && currentTime >= c.CooldownEnd
}

// Use consumes one use, returns true if successful.
func (c *Consumable) Use(currentTime float64) bool {
	if !c.CanUse(currentTime) {
		return false
	}

	c.Uses--
	c.CooldownEnd = currentTime + c.Cooldown

	return true
}

// LootEntry represents an item in a loot table.
type LootEntry struct {
	ItemID   string
	Weight   float64 // Relative weight for selection
	MinCount int
	MaxCount int
	Rarity   ItemRarity // Minimum rarity to drop
}

// LootTable defines possible drops from an entity.
type LootTable struct {
	Entries       []LootEntry
	Guaranteed    []string // Items that always drop
	MinDrops      int      // Minimum items to drop
	MaxDrops      int      // Maximum items to drop
	NothingChance float64  // Chance to drop nothing (0.0-1.0)
}

// NewLootTable creates an empty loot table.
func NewLootTable(minDrops, maxDrops int) LootTable {
	return LootTable{
		Entries:    make([]LootEntry, 0),
		Guaranteed: make([]string, 0),
		MinDrops:   minDrops,
		MaxDrops:   maxDrops,
	}
}

// AddEntry adds a loot entry.
func (lt *LootTable) AddEntry(entry LootEntry) {
	lt.Entries = append(lt.Entries, entry)
}

// CraftingRecipe defines how to craft an item.
type CraftingRecipe struct {
	ID          string
	ResultID    string
	ResultCount int
	Ingredients map[string]int // Item ID -> count required
	CraftTime   float64        // Time to craft in seconds
	Discovered  bool           // Whether recipe is known
}

// NewCraftingRecipe creates a crafting recipe.
func NewCraftingRecipe(id, resultID string, resultCount int) CraftingRecipe {
	return CraftingRecipe{
		ID:          id,
		ResultID:    resultID,
		ResultCount: resultCount,
		Ingredients: make(map[string]int),
		Discovered:  true,
	}
}

// CanCraft checks if inventory has required ingredients.
func (cr *CraftingRecipe) CanCraft(inv *Inventory) bool {
	if !cr.Discovered {
		return false
	}

	for itemID, required := range cr.Ingredients {
		if inv.GetItemCount(itemID) < required {
			return false
		}
	}

	return true
}
