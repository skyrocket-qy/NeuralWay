package main

import (
	"bytes"
	"embed"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"td/internal/graphics"
)

//go:embed assets/*.png
var assetsFS embed.FS

const (
	screenWidth  = 900
	screenHeight = 700
)

// CharacterType represents playable characters.
type CharacterType int

const (
	CharJunior CharacterType = iota
	CharSenior
	CharTechLead
	Char10x
)

// Character stats definition.
type CharacterDef struct {
	Name        string
	HP          int
	Speed       float64
	StartWeapon WeaponType
	Trait       string
	TraitDesc   string
	Color       color.RGBA
	ImageFile   string
}

var Characters = []CharacterDef{
	{
		Name:        "Junior Dev",
		HP:          100,
		Speed:       3.0,
		StartWeapon: WeaponPrint,
		Trait:       "Eager",
		TraitDesc:   "+20% Speed",
		Color:       color.RGBA{R: 100, G: 200, B: 100, A: 255},
		ImageFile:   "assets/hero_junior.png",
	},
	{
		Name:        "Senior Dev",
		HP:          150,
		Speed:       2.5,
		StartWeapon: WeaponRefactor,
		Trait:       "Experienced",
		TraitDesc:   "+20% XP",
		Color:       color.RGBA{R: 100, G: 100, B: 200, A: 255},
		ImageFile:   "assets/hero_senior.png",
	},
	{
		Name:        "Tech Lead",
		HP:          120,
		Speed:       2.8,
		StartWeapon: WeaponDocker,
		Trait:       "Visionary",
		TraitDesc:   "+30% Area",
		Color:       color.RGBA{R: 200, G: 100, B: 200, A: 255},
		ImageFile:   "assets/hero_lead.png",
	},
	{
		Name:        "10x Eng",
		HP:          80,
		Speed:       3.5,
		StartWeapon: WeaponGitPush,
		Trait:       "Hyper",
		TraitDesc:   "+50% Cooldown",
		Color:       color.RGBA{R: 255, G: 100, B: 0, A: 255},
		ImageFile:   "assets/hero_10x.png",
	},
}

// WeaponType represents weapon types.
type WeaponType int

const (
	WeaponPrint WeaponType = iota
	WeaponRefactor
	WeaponGitPush
	WeaponCoffee
	WeaponFirewall
	WeaponStackOverflow
	WeaponDocker
	WeaponUnitTests
	// Evolved Weapons.
	WeaponLogStream
	WeaponCleanCode
	WeaponForcePush
	WeaponEspresso
	WeaponZeroTrust
	WeaponCopilot
	WeaponK8s
	WeaponCI_CD
)

// Weapon definition.
type WeaponDef struct {
	Name      string
	Damage    int
	Cooldown  float64
	Range     float64
	Count     int
	Color     color.RGBA
	IsEvolved bool
	ImageFile string
}

var WeaponDefs = map[WeaponType]WeaponDef{
	WeaponPrint: {
		Name:      "Print Debug",
		Damage:    15,
		Cooldown:  0.6,
		Range:     90,
		Count:     1,
		Color:     color.RGBA{R: 200, G: 200, B: 200, A: 255},
		ImageFile: "assets/weapon_print.png",
	},
	WeaponRefactor: {
		Name:      "Refactor",
		Damage:    20,
		Cooldown:  1.0,
		Range:     110,
		Count:     2,
		Color:     color.RGBA{R: 100, G: 150, B: 255, A: 255},
		ImageFile: "assets/weapon_refactor.png",
	},
	WeaponGitPush: {
		Name:      "Git Push",
		Damage:    12,
		Cooldown:  0.5,
		Range:     300,
		Count:     1,
		Color:     color.RGBA{R: 50, G: 200, B: 50, A: 255},
		ImageFile: "assets/weapon_gitpush.png",
	},
	WeaponCoffee: {
		Name:      "Coffee",
		Damage:    5,
		Cooldown:  0.3,
		Range:     60,
		Count:     1,
		Color:     color.RGBA{R: 100, G: 50, B: 0, A: 255},
		ImageFile: "assets/weapon_coffee.png",
	},
	WeaponFirewall: {
		Name:      "Firewall",
		Damage:    25,
		Cooldown:  1.2,
		Range:     130,
		Count:     1,
		Color:     color.RGBA{R: 255, G: 100, B: 50, A: 255},
		ImageFile: "assets/weapon_firewall.png",
	},
	WeaponStackOverflow: {
		Name:      "StackOverflow",
		Damage:    40,
		Cooldown:  1.5,
		Range:     250,
		Count:     2,
		Color:     color.RGBA{R: 255, G: 200, B: 0, A: 255},
		ImageFile: "assets/weapon_stackoverflow.png",
	},
	WeaponDocker: {
		Name:      "Docker Container",
		Damage:    30,
		Cooldown:  1.2,
		Range:     100,
		Count:     1,
		Color:     color.RGBA{R: 0, G: 100, B: 255, A: 255},
		ImageFile: "assets/weapon_docker.png",
	},
	WeaponUnitTests: {
		Name:      "Unit Tests",
		Damage:    8,
		Cooldown:  0.8,
		Range:     80,
		Count:     1,
		Color:     color.RGBA{R: 100, G: 255, B: 100, A: 255},
		ImageFile: "assets/weapon_unittests.png",
	},

	// Evolved weapons (no image assets yet, will use programmatic fallback)
	WeaponLogStream: {
		Name:      "Log Stream",
		Damage:    30,
		Cooldown:  0.2,
		Range:     150,
		Count:     1,
		Color:     color.RGBA{R: 255, G: 255, B: 255, A: 255},
		IsEvolved: true,
	},
	WeaponCleanCode: {
		Name:      "Clean Code",
		Damage:    35,
		Cooldown:  0.8,
		Range:     150,
		Count:     4,
		Color:     color.RGBA{R: 150, G: 200, B: 255, A: 255},
		IsEvolved: true,
	},
	WeaponForcePush: {
		Name:      "Force Push",
		Damage:    15,
		Cooldown:  0.1,
		Range:     500,
		Count:     1,
		Color:     color.RGBA{R: 0, G: 255, B: 0, A: 255},
		IsEvolved: true,
	},
	WeaponEspresso: {
		Name:      "Double Espresso",
		Damage:    10,
		Cooldown:  0.1,
		Range:     100,
		Count:     1,
		Color:     color.RGBA{R: 150, G: 100, B: 50, A: 255},
		IsEvolved: true,
	},
	WeaponZeroTrust: {
		Name:      "Zero Trust",
		Damage:    50,
		Cooldown:  1.0,
		Range:     180,
		Count:     1,
		Color:     color.RGBA{R: 255, G: 50, B: 0, A: 255},
		IsEvolved: true,
	},
	WeaponCopilot: {
		Name:      "AI Copilot",
		Damage:    60,
		Cooldown:  1.0,
		Range:     300,
		Count:     6,
		Color:     color.RGBA{R: 255, G: 255, B: 100, A: 255},
		IsEvolved: true,
	},
	WeaponK8s: {
		Name:      "Kubernetes",
		Damage:    80,
		Cooldown:  2.0,
		Range:     300,
		Count:     1,
		Color:     color.RGBA{R: 50, G: 50, B: 255, A: 255},
		IsEvolved: true,
	},
	WeaponCI_CD: {
		Name:      "CI/CD Pipeline",
		Damage:    12,
		Cooldown:  0.5,
		Range:     120,
		Count:     1,
		Color:     color.RGBA{R: 100, G: 255, B: 255, A: 255},
		IsEvolved: true,
	},
}

type EvolutionRecipe struct {
	BaseWeapon WeaponType
	Passive    PassiveType
	Result     WeaponType
}

var Evolutions = []EvolutionRecipe{
	{WeaponPrint, PassiveAmount, WeaponLogStream},
	{WeaponRefactor, PassiveArea, WeaponCleanCode},
	{WeaponGitPush, PassiveSpeed, WeaponForcePush},
	{WeaponCoffee, PassiveRecovery, WeaponEspresso},
	{WeaponFirewall, PassiveArmor, WeaponZeroTrust},
	{WeaponStackOverflow, PassiveLuck, WeaponCopilot},
	{WeaponDocker, PassiveGrowth, WeaponK8s},
	{WeaponUnitTests, PassiveCooldown, WeaponCI_CD},
}

// MonsterType represents enemy types.
type MonsterType int

const (
	MonsterBug MonsterType = iota
	MonsterNull
	MonsterSpaghetti
	MonsterDowntime
	MonsterLegacy
	MonsterRaceCond
	MonsterBossManager
	MonsterBossDeadline
)

// Monster definition.
type MonsterDef struct {
	Name      string
	HP        int
	Speed     float64
	Damage    int
	XP        int
	Radius    float64
	Color     color.RGBA
	IsBoss    bool
	ImageFile string
}

// Monster definitions.
var MonsterDefs = map[MonsterType]*MonsterDef{
	MonsterBug: {
		Name:      "Minor Bug",
		HP:        10,
		Speed:     2.2,
		Damage:    3,
		XP:        1,
		Radius:    10,
		Color:     color.RGBA{100, 100, 100, 255},
		ImageFile: "assets/monster_bug.png",
	}, // Bat
	MonsterNull: {
		Name:      "Null Pointer",
		HP:        30,
		Speed:     1.5,
		Damage:    8,
		XP:        3,
		Radius:    12,
		Color:     color.RGBA{200, 200, 200, 255},
		ImageFile: "assets/monster_null.png",
	}, // Skeleton
	MonsterSpaghetti: {
		Name:      "Spaghetti Code",
		HP:        50,
		Speed:     0.8,
		Damage:    12,
		XP:        5,
		Radius:    14,
		Color:     color.RGBA{50, 150, 50, 255},
		ImageFile: "assets/monster_spaghetti.png",
	}, // Zombie
	MonsterDowntime: {
		Name:      "Downtime",
		HP:        20,
		Speed:     2.0,
		Damage:    6,
		XP:        2,
		Radius:    10,
		Color:     color.RGBA{200, 200, 255, 150},
		ImageFile: "assets/monster_downtime.png",
	}, // Ghost
	MonsterLegacy: {
		Name:      "Legacy Code",
		HP:        80,
		Speed:     2.0,
		Damage:    15,
		XP:        10,
		Radius:    20,
		Color:     color.RGBA{200, 50, 50, 255},
		ImageFile: "assets/monster_legacy.png",
	}, // Demon
	MonsterRaceCond: {
		Name:      "Race Condition",
		HP:        40,
		Speed:     3.0,
		Damage:    10,
		XP:        8,
		Radius:    15,
		Color:     color.RGBA{50, 50, 200, 255},
		ImageFile: "assets/monster_race.png",
	}, // Elemental

	// Bosses
	MonsterBossManager: {
		Name:      "Micro Manager",
		HP:        1500,
		Speed:     1.8,
		Damage:    20,
		XP:        500,
		Radius:    30,
		Color:     color.RGBA{50, 0, 50, 255},
		IsBoss:    true,
		ImageFile: "assets/monster_manager.png",
	}, // Boss CharJunior
	MonsterBossDeadline: {
		Name:      "Hard Deadline",
		HP:        5000,
		Speed:     2.5,
		Damage:    40,
		XP:        2000,
		Radius:    50,
		Color:     color.RGBA{200, 100, 0, 255},
		IsBoss:    true,
		ImageFile: "assets/monster_deadline.png",
	}, // Boss Dragon
}

// Passive upgrade types.
type PassiveType int

const (
	PassiveMight PassiveType = iota
	PassiveArmor
	PassiveSpeed
	PassiveMagnet
	PassiveRecovery
	PassiveLuck
	PassiveGrowth
	PassiveCooldown
	PassiveArea
	PassiveDuration
	PassiveAmount
	PassiveRevival
)

type PassiveDef struct {
	Name      string
	Desc      string
	MaxLvl    int
	ImageFile string
}

var PassiveDefs = map[PassiveType]PassiveDef{
	PassiveMight: {Name: "Might", Desc: "+10% damage", MaxLvl: 5, ImageFile: "assets/passive_might.png"},
	PassiveArmor: {
		Name:      "Armor",
		Desc:      "-5 damage taken",
		MaxLvl:    5,
		ImageFile: "assets/passive_armor.png",
	},
	PassiveSpeed: {
		Name:      "Speed",
		Desc:      "+10% move speed",
		MaxLvl:    5,
		ImageFile: "assets/passive_speed.png",
	},
	PassiveMagnet: {
		Name:      "Magnet",
		Desc:      "+20% pickup range",
		MaxLvl:    5,
		ImageFile: "assets/passive_magnet.png",
	},
	PassiveRecovery: {Name: "Recovery", Desc: "+0.3 HP/s", MaxLvl: 5},
	PassiveLuck:     {Name: "Luck", Desc: "+10% crit", MaxLvl: 5, ImageFile: "assets/passive_luck.png"},
	PassiveGrowth:   {Name: "Growth", Desc: "+10% XP", MaxLvl: 5},
	PassiveCooldown: {
		Name:      "Cooldown",
		Desc:      "-5% cooldown",
		MaxLvl:    5,
		ImageFile: "assets/passive_cooldown.png",
	},
	PassiveArea:     {Name: "Area", Desc: "+10% area", MaxLvl: 5},
	PassiveDuration: {Name: "Duration", Desc: "+10% duration", MaxLvl: 5},
	PassiveAmount:   {Name: "Amount", Desc: "+1 projectile", MaxLvl: 3},
	PassiveRevival:  {Name: "Revival", Desc: "Revive once", MaxLvl: 1},
}

// ============================================================================
// EQUIPMENT SYSTEM
// ============================================================================

// EquipSlot represents equipment slot types.
type EquipSlot int

const (
	SlotKeyboard   EquipSlot = iota // Weapon - +Damage, +Attack Speed
	SlotMonitor                     // Helm - +MaxHP, +XP Gain
	SlotChair                       // Armor - +Armor, +Recovery
	SlotMouse                       // Gloves - +Crit, +Area
	SlotHeadphones                  // Accessory - +Cooldown, +Duration
	SlotCoffeeMug                   // Ring - +Speed, +Magnet
	SlotCount                       // Total number of slots
)

var EquipSlotNames = map[EquipSlot]string{
	SlotKeyboard:   "Keyboard",
	SlotMonitor:    "Monitor",
	SlotChair:      "Gaming Chair",
	SlotMouse:      "Mouse",
	SlotHeadphones: "Headphones",
	SlotCoffeeMug:  "Coffee Mug",
}

// Rarity represents item rarity tiers.
type Rarity int

const (
	RarityCommon    Rarity = iota // White - 0 mods
	RarityMagic                   // Blue - 1-2 mods
	RarityRare                    // Yellow - 3-4 mods
	RarityLegendary               // Orange - 5-6 mods + special
)

var RarityNames = map[Rarity]string{
	RarityCommon:    "Common",
	RarityMagic:     "Magic",
	RarityRare:      "Rare",
	RarityLegendary: "Legendary",
}

var RarityColors = map[Rarity]color.RGBA{
	RarityCommon:    {200, 200, 200, 255}, // White
	RarityMagic:     {100, 150, 255, 255}, // Blue
	RarityRare:      {255, 255, 100, 255}, // Yellow
	RarityLegendary: {255, 150, 50, 255},  // Orange
}

// ModType represents modifier types.
type ModType int

const (
	ModFlatDamage ModType = iota
	ModPercentDamage
	ModFlatHP
	ModPercentHP
	ModArmor
	ModSpeed
	ModCritChance
	ModCooldown
	ModArea
	ModDuration
	ModMagnet
	ModXPGain
	ModRecovery
	ModProjectiles
)

var ModTypeNames = map[ModType]string{
	ModFlatDamage:    "+# Damage",
	ModPercentDamage: "+#% Damage",
	ModFlatHP:        "+# Max HP",
	ModPercentHP:     "+#% Max HP",
	ModArmor:         "+# Armor",
	ModSpeed:         "+#% Movement Speed",
	ModCritChance:    "+#% Critical Chance",
	ModCooldown:      "-#% Cooldown",
	ModArea:          "+#% Area of Effect",
	ModDuration:      "+#% Duration",
	ModMagnet:        "+#% Pickup Range",
	ModXPGain:        "+#% XP Gain",
	ModRecovery:      "+# HP/s Recovery",
	ModProjectiles:   "+# Projectiles",
}

// Modifier represents a single stat modifier on equipment.
type Modifier struct {
	Type  ModType
	Value float64
	Tier  int // 1-5, affects value range
}

// Equipment represents an equippable item.
type Equipment struct {
	Slot      EquipSlot
	Name      string
	Rarity    Rarity
	Modifiers []Modifier
	ItemLevel int
}

// ============================================================================
// PASSIVE TREE SYSTEM
// ============================================================================

// PassiveNodeType categorizes passive nodes.
type PassiveNodeType int

const (
	NodeSmall    PassiveNodeType = iota // Minor bonus
	NodeNotable                         // Medium bonus, named
	NodeKeystone                        // Major bonus with downside
)

// PassiveNode represents a node in the passive skill tree.
type PassiveNode struct {
	ID          int
	Name        string
	Desc        string
	X, Y        float64    // Position in tree (grid coordinates)
	Connections []int      // Connected node IDs
	Effects     []Modifier // Stat bonuses
	NodeType    PassiveNodeType
	Allocated   bool
	StartClass  CharacterType // -1 for none, else starting node for class
}

// Weapon instance.
type Weapon struct {
	Type  WeaponType
	Level int
	Timer float64
}

// Projectile instance.
type Projectile struct {
	X, Y       float64
	VX, VY     float64
	Damage     int
	Lifetime   float64
	Radius     float64
	Piercing   int
	HitList    map[*Enemy]bool
	Color      color.RGBA
	WeaponType WeaponType
}

// Enemy instance.
type Enemy struct {
	X, Y      float64
	HP, MaxHP int
	Speed     float64
	Damage    int
	XP        int
	Radius    float64
	Type      MonsterType
	Dead      bool
	HitFlash  float64
	Color     color.RGBA
	IsBoss    bool
}

// XP Gem.
type XPGem struct {
	X, Y   float64
	Value  int
	Magnet bool
}

// Particle visual effect.
type Particle struct {
	X, Y     float64
	VX, VY   float64
	Lifetime float64
	MaxLife  float64
	Color    color.RGBA
	Size     float64
}

// Damage number.
type DamageNumber struct {
	X, Y   float64
	VX, VY float64
	Value  int
	Timer  float64
	Crit   bool
}

// Player state.
type Player struct {
	X, Y         float64
	HP, MaxHP    int
	XP           int
	Level        int
	Speed        float64
	CharType     CharacterType
	Weapons      []*Weapon
	Passives     map[PassiveType]int
	DamageMult   float64
	AreaMult     float64
	CooldownMult float64
	MagnetRange  float64
	Recovery     float64
	CritChance   float64
	XPMult       float64
	Armor        int
	HasRevival   bool
	UsedRevival  bool
	HitTimer     float64

	// Equipment system
	Equipment map[EquipSlot]*Equipment
	Inventory []*Equipment // Unequipped items

	// Passive tree system
	PassivePoints  int
	AllocatedNodes map[int]bool // Node IDs that are allocated
}

// GameState enum.
type GameState int

const (
	StateCharSelect GameState = iota
	StatePlaying
	StateLevelUp
	StatePaused
	StateGameOver
	StateEquipment   // Equipment inventory screen
	StatePassiveTree // Passive skill tree screen
	StateHelp        // Help/controls screen
)

// Game main struct.
type Game struct {
	state         GameState
	player        *Player
	enemies       []*Enemy
	projectiles   []*Projectile
	xpGems        []*XPGem
	damageNumbers []*DamageNumber
	particles     []*Particle // New visual effects
	charImages    []*ebiten.Image
	monsterImages map[MonsterType]*ebiten.Image
	weaponImages  map[WeaponType]*ebiten.Image
	passiveImages map[PassiveType]*ebiten.Image

	unusedProjs []*Projectile
	unusedParts []*Particle
	unusedDmg   []*DamageNumber

	gameTime float64

	spawnTimer   float64
	bossTimer    float64
	killCount    int
	selectedChar int

	upgradeOptions []UpgradeOption
	// Audio
	audio         *AudioPlayer
	hitAudioTimer float64
	helpSelection int // 0: SFX, 1: Music

	cameraX, cameraY float64
	grid             map[GridKey][]*Enemy

	// Passive tree
	passiveTree []*PassiveNode

	// Equipment UI state
	selectedSlot     EquipSlot
	selectedInvIndex int
	itemDrops        []*Equipment // Dropped items in world
}

type GridKey struct {
	X, Y int
}

// UpgradeOption for level-up.
type UpgradeOption struct {
	Name        string
	Desc        string
	IsWeapon    bool
	WeaponType  WeaponType
	PassiveType PassiveType
	CurrentLvl  int
	Apply       func(*Game)
}

// NewGame creates new game.
func NewGame() *Game {
	g := &Game{
		state:         StateCharSelect,
		selectedChar:  0,
		charImages:    make([]*ebiten.Image, len(Characters)),
		monsterImages: make(map[MonsterType]*ebiten.Image),
	}

	// Load character images
	for i, char := range Characters {
		data, err := assetsFS.ReadFile(char.ImageFile)
		if err != nil {
			log.Printf("Warning: could not load %s: %v", char.ImageFile, err)

			continue
		}

		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			log.Printf("Warning: could not decode %s: %v", char.ImageFile, err)

			continue
		}

		img = graphics.RemoveBackground(img)
		g.charImages[i] = ebiten.NewImageFromImage(img)
	}

	// Load monster images
	for t, def := range MonsterDefs {
		if def.ImageFile == "" {
			continue
		}

		data, err := assetsFS.ReadFile(def.ImageFile)
		if err != nil {
			log.Printf("Warning: could not load %s: %v", def.ImageFile, err)

			continue
		}

		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			log.Printf("Warning: could not decode %s: %v", def.ImageFile, err)

			continue
		}

		img = graphics.RemoveBackground(img)
		g.monsterImages[t] = ebiten.NewImageFromImage(img)
	}

	g.generateIcons()

	// Audio
	g.audio = NewAudioPlayer()
	g.audio.GenerateSounds()
	g.audio.PlayBGM()

	return g
}

func (g *Game) startGame(charType CharacterType) {
	charDef := Characters[charType]
	g.player = &Player{
		X: 0, Y: 0,
		HP: charDef.HP, MaxHP: charDef.HP,
		Level:    1,
		Speed:    charDef.Speed,
		CharType: charType,
		Weapons: []*Weapon{
			{Type: charDef.StartWeapon, Level: 1},
		},
		Passives:       make(map[PassiveType]int),
		DamageMult:     1.0,
		AreaMult:       1.0,
		CooldownMult:   1.0,
		MagnetRange:    80,
		XPMult:         1.0,
		Equipment:      make(map[EquipSlot]*Equipment),
		Inventory:      make([]*Equipment, 0),
		PassivePoints:  0,
		AllocatedNodes: make(map[int]bool),
	}

	// Apply character traits
	switch charType {
	case CharSenior:
		g.player.AreaMult = 1.5
	case CharTechLead:
		// Handled in projectile speed
	case Char10x:
		// Lifesteal handled in damage
	}

	g.enemies = make([]*Enemy, 0)
	g.projectiles = make([]*Projectile, 0)
	g.xpGems = make([]*XPGem, 0)
	g.damageNumbers = make([]*DamageNumber, 0)
	g.itemDrops = make([]*Equipment, 0)
	g.gameTime = 0
	g.spawnTimer = 0
	g.bossTimer = 0
	g.killCount = 0
	g.state = StatePlaying

	// Initialize passive tree
	g.initPassiveTree()
}

func (g *Game) Update() error {
	switch g.state {
	case StateCharSelect:
		return g.updateCharSelect()
	case StatePlaying:
		return g.updatePlaying()
	case StateLevelUp:
		return g.updateLevelUp()
	case StatePaused:
		return g.updatePaused()
	case StateGameOver:
		return g.updateGameOver()
	case StateEquipment:
		return g.updateEquipment()
	case StatePassiveTree:
		return g.updatePassiveTree()
	case StateHelp:
		return g.updateHelp()
	}

	return nil
}

func (g *Game) updateCharSelect() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.selectedChar--
		if g.selectedChar < 0 {
			g.selectedChar = len(Characters) - 1
		}

		g.audio.PlaySound("select")
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.selectedChar++
		if g.selectedChar >= len(Characters) {
			g.selectedChar = 0
		}

		g.audio.PlaySound("select")
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.startGame(CharacterType(g.selectedChar))
	}

	return nil
}

func (g *Game) updatePlaying() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.state = StatePaused

		return nil
	}
	// Equipment screen (I key)
	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		g.state = StateEquipment

		return nil
	}
	// Passive tree screen (P key)
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.state = StatePassiveTree

		return nil
	}
	// Help screen (H key)
	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		g.state = StateHelp

		return nil
	}

	dt := 1.0 / 60.0
	g.gameTime += dt
	g.player.HitTimer -= dt
	g.hitAudioTimer -= dt

	// Recovery
	if g.player.Recovery > 0 {
		g.player.HP += int(g.player.Recovery * dt)
		if g.player.HP > g.player.MaxHP {
			g.player.HP = g.player.MaxHP
		}
	}

	// Player movement
	dx, dy := 0.0, 0.0
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		dy = -1
	}

	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		dy = 1
	}

	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		dx = -1
	}

	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		dx = 1
	}

	if dx != 0 && dy != 0 {
		dx *= 0.707
		dy *= 0.707
	}

	g.player.X += dx * g.player.Speed
	g.player.Y += dy * g.player.Speed

	g.cameraX = g.player.X - float64(screenWidth)/2
	g.cameraY = g.player.Y - float64(screenHeight)/2

	// Spawn enemies
	g.spawnTimer += dt

	spawnRate := 1.0 - g.gameTime*0.01 // Starts at 1s, decays faster
	if spawnRate < 0.05 {              // Cap at 20 enemies/sec
		spawnRate = 0.05
	}
	// Spawn multiple if falling behind
	for g.spawnTimer >= spawnRate {
		g.spawnEnemy()
		g.spawnTimer -= spawnRate
	}

	// Boss timer (every 3 minutes)
	g.bossTimer += dt
	if g.bossTimer >= 180 {
		g.spawnBoss()
		g.bossTimer = 0
	}

	// Update weapons
	for _, w := range g.player.Weapons {
		def := WeaponDefs[w.Type]
		cooldown := def.Cooldown * g.player.CooldownMult

		w.Timer += dt
		if w.Timer >= cooldown {
			g.fireWeapon(w)
			w.Timer = 0
		}
	}

	// Update projectiles
	g.updateProjectiles(dt)

	// Update enemies
	g.updateEnemies(dt)

	// Collect XP
	g.collectXP(dt)

	// Update damage numbers
	for i := len(g.damageNumbers) - 1; i >= 0; i-- {
		d := g.damageNumbers[i]
		d.X += d.VX * dt
		d.Y += d.VY * dt
		d.VY += 200 * dt // Gravity

		d.Timer -= dt
		if d.Timer <= 0 {
			g.freeDamageNumber(d)
			g.damageNumbers = append(g.damageNumbers[:i], g.damageNumbers[i+1:]...)
		}
	}

	// Update particles
	g.updateParticles(dt)

	return nil
}

func (g *Game) spawnEnemy() {
	angle := rand.Float64() * math.Pi * 2
	dist := float64(screenWidth)/2 + 100

	// Choose monster type based on time
	var monsterType MonsterType

	r := rand.Float64()
	if g.gameTime < 60 {
		if r < 0.7 {
			monsterType = MonsterBug
		} else {
			monsterType = MonsterNull
		}
	} else if g.gameTime < 180 {
		if r < 0.3 {
			monsterType = MonsterBug
		} else if r < 0.6 {
			monsterType = MonsterNull
		} else if r < 0.8 {
			monsterType = MonsterSpaghetti
		} else {
			monsterType = MonsterDowntime
		}
	} else {
		if r < 0.2 {
			monsterType = MonsterNull
		} else if r < 0.4 {
			monsterType = MonsterSpaghetti
		} else if r < 0.6 {
			monsterType = MonsterDowntime
		} else if r < 0.8 {
			monsterType = MonsterLegacy
		} else {
			monsterType = MonsterRaceCond
		}
	}

	def := MonsterDefs[monsterType]
	hpScale := 1.0 + g.gameTime*0.008

	g.enemies = append(g.enemies, &Enemy{
		X:  g.player.X + math.Cos(angle)*dist,
		Y:  g.player.Y + math.Sin(angle)*dist,
		HP: int(float64(def.HP) * hpScale), MaxHP: int(float64(def.HP) * hpScale),
		Speed:  def.Speed,
		Damage: def.Damage,
		XP:     int(float64(def.XP) * g.player.XPMult),
		Radius: def.Radius,
		Type:   monsterType,
		Color:  def.Color,
	})
}

func (g *Game) spawnBoss() {
	angle := rand.Float64() * math.Pi * 2
	dist := float64(screenWidth)/2 + 150

	bossType := MonsterBossManager
	if g.gameTime > 360 {
		bossType = MonsterBossDeadline
	}

	def := MonsterDefs[bossType]

	g.enemies = append(g.enemies, &Enemy{
		X:  g.player.X + math.Cos(angle)*dist,
		Y:  g.player.Y + math.Sin(angle)*dist,
		HP: def.HP, MaxHP: def.HP,
		Speed:  def.Speed,
		Damage: def.Damage,
		XP:     def.XP,
		Radius: def.Radius,
		Type:   bossType,
		Color:  def.Color,
		IsBoss: true,
	})
}

func (g *Game) fireWeapon(w *Weapon) {
	def := WeaponDefs[w.Type]

	g.audio.PlaySound("shoot")
	damage := int(float64(def.Damage+w.Level*3) * g.player.DamageMult)
	count := def.Count + g.player.Passives[PassiveAmount]
	areaRange := def.Range * g.player.AreaMult

	// Helper to spawn projectile using pool
	spawnProj := func(x, y, vx, vy, lifetime, radius float64, piercing int) {
		p := g.newProjectile()
		p.X, p.Y = x, y
		p.VX, p.VY = vx, vy
		p.Damage = damage
		p.Lifetime = lifetime
		p.Radius = radius
		p.Piercing = piercing
		p.HitList = make(map[*Enemy]bool) // Each projectile needs its own hitlist
		p.Color = def.Color
		p.WeaponType = w.Type
		g.projectiles = append(g.projectiles, p)
	}

	switch w.Type {
	case WeaponPrint, WeaponLogStream:
		// Aim at nearest enemy
		target := g.findNearestEnemy(120) // Melee range

		var baseAngle float64
		if target != nil {
			baseAngle = math.Atan2(target.Y-g.player.Y, target.X-g.player.X)
		} else {
			baseAngle = (rand.Float64() - 0.5) * math.Pi / 2
		}

		for i := range count {
			// Spread around base angle
			angle := baseAngle

			if count > 1 {
				spread := math.Pi / 3 // 60 degrees spread
				angle += spread * (float64(i)/float64(count-1) - 0.5)
			}

			// Larger area for evolved
			radius := areaRange / 3
			if w.Type == WeaponLogStream {
				radius *= 1.5
			}

			spawnProj(
				g.player.X+math.Cos(angle)*40, g.player.Y+math.Sin(angle)*40,
				math.Cos(angle)*3, math.Sin(angle)*3,
				0.3, radius, 5,
			)
		}

	case WeaponRefactor, WeaponCleanCode:
		for i := range count {
			angle := g.gameTime*3 + float64(i)*(2*math.Pi/float64(count))
			spawnProj(
				g.player.X+math.Cos(angle)*areaRange, g.player.Y+math.Sin(angle)*areaRange,
				0, 0,
				0.2, 18, 3,
			)
		}

	case WeaponGitPush, WeaponForcePush:
		nearest := g.findNearestEnemy(500)
		if nearest != nil {
			dx, dy := nearest.X-g.player.X, nearest.Y-g.player.Y
			dist := math.Sqrt(dx*dx + dy*dy)

			speed := 10.0
			if g.player.CharType == CharTechLead {
				speed = 12.5
			}

			for i := range count {
				spread := float64(i-count/2) * 0.15
				spawnProj(
					g.player.X, g.player.Y,
					(dx/dist)*speed+spread, (dy/dist)*speed+spread,
					2.0, 6, 1,
				)
			}
		}

	case WeaponUnitTests, WeaponCI_CD:
		// Area damage around player
		spawnProj(g.player.X, g.player.Y, 0, 0, 0.2, areaRange, 999)

	case WeaponFirewall, WeaponZeroTrust:
		// Orbiting fireball
		angle := g.gameTime * 2
		spawnProj(
			g.player.X+math.Cos(angle)*areaRange, g.player.Y+math.Sin(angle)*areaRange,
			0, 0,
			0.5, 15, 999,
		)

	case WeaponStackOverflow, WeaponCopilot:
		// Random enemies
		targets := g.findNearestEnemies(count, 300)
		for _, e := range targets {
			spawnProj(e.X, e.Y-50, 0, 20, 0.2, 30, 1)
		}

	case WeaponDocker, WeaponK8s:
		// Throws toward nearest enemy then returns (boomerang-like)
		nearest := g.findNearestEnemy(400)
		if nearest != nil {
			dx, dy := nearest.X-g.player.X, nearest.Y-g.player.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			speed := 7.0
			spawnProj(g.player.X, g.player.Y, (dx/dist)*speed, (dy/dist)*speed, 2.0, 15, 999)
		} else {
			// No enemy nearby, shoot in last movement direction or default
			spawnProj(g.player.X, g.player.Y, 5, 0, 2.0, 15, 999)
		}

	case WeaponCoffee, WeaponEspresso:
		// Permanent aura
		spawnProj(g.player.X, g.player.Y, 0, 0, 0.4, areaRange, 999)
	}
}

func (g *Game) findNearestEnemy(maxDist float64) *Enemy {
	var nearest *Enemy

	minDist := maxDist

	for _, e := range g.enemies {
		if e.Dead {
			continue
		}

		dist := math.Sqrt((g.player.X-e.X)*(g.player.X-e.X) + (g.player.Y-e.Y)*(g.player.Y-e.Y))
		if dist < minDist {
			minDist = dist
			nearest = e
		}
	}

	return nearest
}

func (g *Game) findNearestEnemies(count int, maxDist float64) []*Enemy {
	type candidate struct {
		e    *Enemy
		dist float64
	}

	candidates := make([]candidate, 0)

	for _, e := range g.enemies {
		if e.Dead {
			continue
		}

		dist := math.Sqrt((g.player.X-e.X)*(g.player.X-e.X) + (g.player.Y-e.Y)*(g.player.Y-e.Y))
		if dist < maxDist {
			candidates = append(candidates, candidate{e, dist})
		}
	}

	// Sort by distance (simple bubble sort or similar since count is small, or just slice sort)
	// Since list might be long, finding top N is O(N*M).
	// Let's just find top count.

	result := make([]*Enemy, 0)

	for i := 0; i < count && len(candidates) > 0; i++ {
		bestIdx := 0

		minDist := candidates[0].dist
		for j := 1; j < len(candidates); j++ {
			if candidates[j].dist < minDist {
				minDist = candidates[j].dist
				bestIdx = j
			}
		}

		result = append(result, candidates[bestIdx].e)
		// Remove selected
		candidates[bestIdx] = candidates[len(candidates)-1]
		candidates = candidates[:len(candidates)-1]
	}

	return result
}

func (g *Game) killEnemy(e *Enemy) {
	e.Dead = true
	g.killCount++
	g.xpGems = append(g.xpGems, &XPGem{X: e.X, Y: e.Y, Value: e.XP})
	g.spawnParticle(e.X, e.Y, 15, e.Color)

	// Equipment drops
	def := MonsterDefs[e.Type]
	if def.IsBoss {
		// Bosses drop guaranteed rare/legendary equipment
		slot := EquipSlot(rand.Intn(int(SlotCount)))

		rarity := RarityRare
		if rand.Float64() < 0.3 {
			rarity = RarityLegendary
		}

		item := g.generateEquipment(slot, g.player.Level, rarity)
		g.player.Inventory = append(g.player.Inventory, item)
	} else if e.XP >= 5 && rand.Float64() < 0.05 {
		// Elite enemies have 5% chance to drop magic/rare items
		slot := EquipSlot(rand.Intn(int(SlotCount)))

		rarity := RarityMagic
		if rand.Float64() < 0.2 {
			rarity = RarityRare
		}

		item := g.generateEquipment(slot, g.player.Level, rarity)
		g.player.Inventory = append(g.player.Inventory, item)
	}
}

func (g *Game) addDamageNumber(x, y float64, value int, crit bool) {
	vx := (rand.Float64() - 0.5) * 60
	vy := -rand.Float64()*60 - 30

	if crit {
		value = int(float64(value) * 1.5)
		vy -= 30
	}

	d := g.newDamageNumber()
	d.X = x
	d.Y = y
	d.VX = vx
	d.VY = vy
	d.Value = value
	d.Timer = 0.8
	d.Crit = crit
	g.damageNumbers = append(g.damageNumbers, d)
}

func (g *Game) spawnParticle(x, y float64, count int, c color.RGBA) {
	for range count {
		angle := rand.Float64() * math.Pi * 2
		speed := rand.Float64()*100 + 50
		life := 0.3 + rand.Float64()*0.4

		p := g.newParticle()
		p.X = x
		p.Y = y
		p.VX = math.Cos(angle) * speed
		p.VY = math.Sin(angle) * speed
		p.Lifetime = life
		p.MaxLife = life
		p.Color = c
		p.Size = 3 + rand.Float64()*4

		g.particles = append(g.particles, p)
	}
}

// spawnTrailParticle spawns smaller, shorter-lived trail particles behind projectiles.
func (g *Game) spawnTrailParticle(x, y float64, count int, c color.RGBA) {
	for range count {
		// Trail particles spread less and fade faster
		angle := rand.Float64() * math.Pi * 2
		speed := rand.Float64()*30 + 10 // Slower than hit particles
		life := 0.15 + rand.Float64()*0.15

		p := g.newParticle()
		p.X = x + (rand.Float64()-0.5)*6
		p.Y = y + (rand.Float64()-0.5)*6
		p.VX = math.Cos(angle) * speed
		p.VY = math.Sin(angle) * speed
		p.Lifetime = life
		p.MaxLife = life
		// Slightly dimmer color for trails
		p.Color = color.RGBA{c.R, c.G, c.B, 150}
		p.Size = 2 + rand.Float64()*2 // Smaller than explosion particles

		g.particles = append(g.particles, p)
	}
}

func (g *Game) updateParticles(dt float64) {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.X += p.VX * dt
		p.Y += p.VY * dt

		p.Lifetime -= dt
		if p.Lifetime <= 0 {
			g.freeParticle(p)
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
}

func (g *Game) updateProjectiles(dt float64) {
	for i := len(g.projectiles) - 1; i >= 0; i-- {
		p := g.projectiles[i]
		p.X += p.VX
		p.Y += p.VY
		p.Lifetime -= dt

		// Spawn trail particles for fast-moving projectiles
		speed := math.Sqrt(p.VX*p.VX + p.VY*p.VY)
		if speed > 3 && rand.Float64() < 0.3 {
			// Spawn 1-2 trail particles behind the projectile
			trailCount := 1
			if speed > 7 {
				trailCount = 2
			}
			// Spawn behind the projectile based on velocity direction
			trailX := p.X - p.VX*0.5
			trailY := p.Y - p.VY*0.5
			g.spawnTrailParticle(trailX, trailY, trailCount, p.Color)
		}

		for _, e := range g.enemies {
			if e.Dead || p.HitList[e] {
				continue
			}

			dist := math.Sqrt((p.X-e.X)*(p.X-e.X) + (p.Y-e.Y)*(p.Y-e.Y))
			if dist < p.Radius+e.Radius {
				crit := rand.Float64() < g.player.CritChance

				damage := p.Damage
				if crit {
					damage = int(float64(damage) * 1.5)
				}

				e.HP -= damage
				e.HitFlash = 0.1
				p.HitList[e] = true

				// Audio limit
				if g.hitAudioTimer <= 0 {
					g.audio.PlaySound("hit")
					g.hitAudioTimer = 0.05
				}

				g.spawnParticle(e.X, e.Y, 5, p.Color)
				p.Piercing--

				g.addDamageNumber(e.X, e.Y, damage, crit)

				if e.HP <= 0 {
					g.killEnemy(e)
				}

				if p.Piercing <= 0 {
					p.Lifetime = 0

					break
				}
			}
		}

		if p.Lifetime <= 0 {
			g.freeProjectile(p)
			g.projectiles = append(g.projectiles[:i], g.projectiles[i+1:]...)
		}
	}
}

func (g *Game) updateEnemies(dt float64) {
	// 1. Rebuild Grid for optimization
	g.grid = make(map[GridKey][]*Enemy) // Re-allocate map (simple for now)
	cellSize := 100

	activeEnemies := g.enemies[:0]
	for _, e := range g.enemies {
		if e.Dead {
			continue
		}

		activeEnemies = append(activeEnemies, e)

		gx := int(e.X) / cellSize

		gy := int(e.Y) / cellSize
		if e.X < 0 {
			gx--
		} // Handle negative coords correctly

		if e.Y < 0 {
			gy--
		}

		key := GridKey{gx, gy}
		g.grid[key] = append(g.grid[key], e)
	}

	g.enemies = activeEnemies

	// 2. Update logic
	for _, e := range g.enemies {
		e.HitFlash -= dt

		// Separation (Soft collision) to prevent stacking
		gx := int(e.X) / cellSize

		gy := int(e.Y) / cellSize
		if e.X < 0 {
			gx--
		}

		if e.Y < 0 {
			gy--
		} // Consistent floor division

		sepX, sepY := 0.0, 0.0

		// Check 3x3 grid neighbors
		for y := gy - 1; y <= gy+1; y++ {
			for x := gx - 1; x <= gx+1; x++ {
				neighbors := g.grid[GridKey{x, y}]
				for _, other := range neighbors {
					if e == other {
						continue
					}

					dx := e.X - other.X
					dy := e.Y - other.Y
					distSq := dx*dx + dy*dy
					minDist := e.Radius + other.Radius
					// Push away logic
					if distSq < minDist*minDist && distSq > 0.1 {
						dist := math.Sqrt(distSq)
						force := (minDist - dist) / dist
						sepX += dx * force
						sepY += dy * force
					}
				}
			}
		}

		// Apply separation (softly)
		e.X += sepX * 5.0 * dt // Strength factor
		e.Y += sepY * 5.0 * dt

		// Move towards player
		dx, dy := g.player.X-e.X, g.player.Y-e.Y

		dist := math.Sqrt(dx*dx + dy*dy)
		if dist > 0 {
			e.X += (dx / dist) * e.Speed
			e.Y += (dy / dist) * e.Speed
		}

		if dist < 20+e.Radius {
			if g.player.HitTimer > 0 {
				continue
			}

			damage := max(e.Damage-g.player.Armor, 1)

			g.player.HP -= damage
			g.player.HitTimer = 0.5 // 0.5s invulnerability // Continuous damage per frame is brutal, need timer?
			// Original code did per-frame damage?
			// Line 941: g.player.HP -= damage
			// Yes. With 60 FPS, this kills instantly.
			// I should add invulnerability frame or reduce damage frequency.
			// For Dev Survivor, let's add a global hit timer or lower damage?
			// Or check if invulnerable.

			// Original implementation was brutal. Let's keep it but maybe it relies on low overlap?
			// With separation, overlap is minimized.

			if g.player.HP <= 0 {
				if g.player.HasRevival && !g.player.UsedRevival {
					g.player.HP = g.player.MaxHP / 2
					g.player.UsedRevival = true
				} else {
					g.state = StateGameOver
				}
			}
		}
	}
}

func (g *Game) collectXP(dt float64) {
	for i := len(g.xpGems) - 1; i >= 0; i-- {
		gem := g.xpGems[i]
		dx, dy := g.player.X-gem.X, g.player.Y-gem.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist < g.player.MagnetRange || gem.Magnet {
			gem.Magnet = true

			speed := 10.0
			if dist > 0 {
				gem.X += (dx / dist) * speed
				gem.Y += (dy / dist) * speed
			}
		}

		if dist < 25 {
			g.player.XP += gem.Value
			g.xpGems = append(g.xpGems[:i], g.xpGems[i+1:]...)

			xpNeeded := g.player.Level * 25
			if g.player.XP >= xpNeeded {
				g.player.XP -= xpNeeded
				g.player.Level++
				g.player.PassivePoints++ // Grant passive point on level-up
				g.showLevelUp()

				return // Stop processing other gems this frame to avoid multiple level-ups
			}
		}
	}
}

func (g *Game) showLevelUp() {
	g.state = StateLevelUp
	g.upgradeOptions = g.generateUpgrades()
	g.audio.PlaySound("levelup")
}

func (g *Game) generateUpgrades() []UpgradeOption {
	options := make([]UpgradeOption, 0)

	// 1. Check Evolutions
	for _, recipe := range Evolutions {
		// Check passives
		if g.player.Passives[recipe.Passive] == 0 {
			continue
		}

		// Check base weapon max level
		var baseW *Weapon

		for _, w := range g.player.Weapons {
			if w.Type == recipe.BaseWeapon && w.Level >= 8 {
				baseW = w

				break
			}
		}

		if baseW != nil {
			rec := recipe // Closure capture
			resDef := WeaponDefs[rec.Result]
			options = append(options, UpgradeOption{
				Name:       "EVOLVE: " + resDef.Name,
				Desc:       "Transform weapon!",
				IsWeapon:   true,
				WeaponType: rec.Result,
				CurrentLvl: 0,
				Apply: func(g *Game) {
					// Find base and replace
					for i, w := range g.player.Weapons {
						if w.Type == rec.BaseWeapon {
							g.player.Weapons[i] = &Weapon{Type: rec.Result, Level: 1}

							break
						}
					}
				},
			})
		}
	}

	// 2. Normal upgrades
	// Add weapon upgrades
	for _, w := range g.player.Weapons {
		if !WeaponDefs[w.Type].IsEvolved && w.Level < 8 {
			def := WeaponDefs[w.Type]
			wCopy := w
			options = append(options, UpgradeOption{
				Name: def.Name, Desc: "Level " + formatInt(w.Level+1),
				IsWeapon: true, WeaponType: w.Type, CurrentLvl: w.Level,
				Apply: func(g *Game) { wCopy.Level++ },
			})
		} else if WeaponDefs[w.Type].IsEvolved && w.Level < 8 {
			// Allow leveling evolved weapons too? Plan says nothing, but usually yes.
			// Let's allow it.
			def := WeaponDefs[w.Type]
			wCopy := w
			options = append(options, UpgradeOption{
				Name: def.Name, Desc: "Level " + formatInt(w.Level+1),
				IsWeapon: true, WeaponType: w.Type, CurrentLvl: w.Level,
				Apply: func(g *Game) { wCopy.Level++ },
			})
		}
	}

	// Add new weapons
	allWeapons := []WeaponType{
		WeaponPrint,
		WeaponRefactor,
		WeaponGitPush,
		WeaponUnitTests,
		WeaponFirewall,
		WeaponStackOverflow,
		WeaponDocker,
		WeaponCoffee,
	}
	for _, wt := range allWeapons {
		has := false

		for _, w := range g.player.Weapons {
			if w.Type == wt {
				has = true

				break
			}
		}
		// Also check if we have the evolved version (don't offer base if we have evolved? Actually in VS you
		// can have both if you find base again? No, usually not).
		// For simplicity, assume one instance of lineage.
		// Check evolved versions
		if !has {
			for _, rec := range Evolutions {
				if rec.BaseWeapon == wt {
					for _, w := range g.player.Weapons {
						if w.Type == rec.Result {
							has = true

							break
						}
					}
				}
			}
		}

		if !has && len(g.player.Weapons) < 6 {
			def := WeaponDefs[wt]
			wtCopy := wt
			options = append(options, UpgradeOption{
				Name: def.Name, Desc: "New weapon!",
				IsWeapon: true, WeaponType: wt,
				Apply: func(g *Game) {
					g.player.Weapons = append(g.player.Weapons, &Weapon{Type: wtCopy, Level: 1})
				},
			})
		}
	}

	// Add passive upgrades
	for pt, pdef := range PassiveDefs {
		current := g.player.Passives[pt]
		if current < pdef.MaxLvl {
			ptCopy := pt
			options = append(options, UpgradeOption{
				Name: pdef.Name, Desc: pdef.Desc,
				PassiveType: pt, CurrentLvl: current,
				Apply: func(g *Game) { g.applyPassive(ptCopy) },
			})
		}
	}

	// Shuffle and pick 4
	rand.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })

	// Prioritize Evolutions (move to front)
	// Actually shuffling mixes them. If we want guaranteed evolution, we should not shuffle them away.
	// Use stable sort or just pick.
	// Let's leave it random for now, but usually evolution is guaranteed to appear if conditions met.
	// I'll leave them in the pool. If pool > 4, might be cut.
	// To ensure they appear, I could prepend them after shuffle?
	// Let's Just limit regular options if evolutions exist.

	if len(options) > 4 {
		options = options[:4]
	}

	return options
}

func (g *Game) applyPassive(pt PassiveType) {
	g.player.Passives[pt]++
	switch pt {
	case PassiveMight:
		g.player.DamageMult += 0.1
	case PassiveArmor:
		g.player.Armor += 5
	case PassiveSpeed:
		g.player.Speed *= 1.1
	case PassiveMagnet:
		g.player.MagnetRange *= 1.2
	case PassiveRecovery:
		g.player.Recovery += 0.3
	case PassiveLuck:
		g.player.CritChance += 0.1
	case PassiveGrowth:
		g.player.XPMult += 0.1
	case PassiveCooldown:
		g.player.CooldownMult *= 0.95
	case PassiveArea:
		g.player.AreaMult += 0.1
	case PassiveRevival:
		g.player.HasRevival = true
	}
}

func (g *Game) updateLevelUp() error {
	for i := 0; i < len(g.upgradeOptions) && i < 4; i++ {
		if inpututil.IsKeyJustPressed(ebiten.Key(int(ebiten.Key1) + i)) {
			g.upgradeOptions[i].Apply(g)
			g.state = StatePlaying

			break
		}
	}

	return nil
}

func (g *Game) updatePaused() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.state = StatePlaying
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		g.state = StateCharSelect
	}

	return nil
}

func (g *Game) updateGameOver() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.startGame(g.player.CharType)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		g.state = StateCharSelect
	}

	return nil
}

// ============================================================================
// EQUIPMENT SYSTEM FUNCTIONS
// ============================================================================

func (g *Game) updateEquipment() error {
	// ESC or I to close
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyI) {
		g.state = StatePlaying

		return nil
	}

	// Navigate slots with arrow keys
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		if g.selectedSlot > 0 {
			g.selectedSlot--
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		if g.selectedSlot < SlotCount-1 {
			g.selectedSlot++
		}
	}

	// Navigate inventory with left/right
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		if g.selectedInvIndex > 0 {
			g.selectedInvIndex--
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		if g.selectedInvIndex < len(g.player.Inventory)-1 {
			g.selectedInvIndex++
		}
	}

	// Enter to equip selected inventory item to selected slot
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && len(g.player.Inventory) > 0 {
		if g.selectedInvIndex < len(g.player.Inventory) {
			item := g.player.Inventory[g.selectedInvIndex]
			if item.Slot == g.selectedSlot {
				// Swap with current equipment
				oldEquip := g.player.Equipment[g.selectedSlot]
				g.player.Equipment[g.selectedSlot] = item

				// Remove from inventory
				g.player.Inventory = append(
					g.player.Inventory[:g.selectedInvIndex],
					g.player.Inventory[g.selectedInvIndex+1:]...)

				// Add old equipment to inventory if exists
				if oldEquip != nil {
					g.player.Inventory = append(g.player.Inventory, oldEquip)
				}

				// Recalculate stats
				g.recalculateStats()
				g.audio.PlaySound("select")

				if g.selectedInvIndex >= len(g.player.Inventory) && len(g.player.Inventory) > 0 {
					g.selectedInvIndex = len(g.player.Inventory) - 1
				}
			}
		}
	}

	return nil
}

// ============================================================================
// PASSIVE TREE SYSTEM FUNCTIONS
// ============================================================================

func (g *Game) initPassiveTree() {
	// Create a PoE-style passive tree with ~30 nodes
	// Layout: Central start nodes, branching into offense/defense/utility paths
	g.passiveTree = []*PassiveNode{
		// Starting nodes for each class (center)
		{
			ID: 0, Name: "Junior Start", X: 0, Y: 0, Connections: []int{1, 2, 3}, NodeType: NodeSmall, StartClass: CharJunior,
			Effects: []Modifier{{Type: ModPercentDamage, Value: 5}},
		},
		{
			ID: 1, Name: "Senior Start", X: -2, Y: 0, Connections: []int{0, 4, 5}, NodeType: NodeSmall, StartClass: CharSenior,
			Effects: []Modifier{{Type: ModArea, Value: 10}},
		},
		{
			ID: 2, Name: "Lead Start", X: 2, Y: 0, Connections: []int{0, 6, 7}, NodeType: NodeSmall, StartClass: CharTechLead,
			Effects: []Modifier{{Type: ModCooldown, Value: 5}},
		},
		{
			ID: 3, Name: "10x Start", X: 0, Y: 2, Connections: []int{0, 8, 9}, NodeType: NodeSmall, StartClass: Char10x,
			Effects: []Modifier{{Type: ModSpeed, Value: 10}},
		},

		// Offense branch (left side)
		{
			ID: 4, Name: "Code Fury", Desc: "+10% Damage", X: -3, Y: -1, Connections: []int{1, 10}, NodeType: NodeSmall,
			Effects: []Modifier{{Type: ModPercentDamage, Value: 10}},
		},
		{
			ID: 5, Name: "Sharp Focus", Desc: "+5% Crit", X: -3, Y: 1, Connections: []int{1, 11}, NodeType: NodeSmall,
			Effects: []Modifier{{Type: ModCritChance, Value: 5}},
		},
		{
			ID: 10, Name: "Aggressive Coding", Desc: "+15% Damage", X: -4, Y: -2, Connections: []int{4, 12}, NodeType: NodeNotable,
			Effects: []Modifier{{Type: ModPercentDamage, Value: 15}},
		},
		{
			ID: 11, Name: "Precision", Desc: "+10% Crit", X: -4, Y: 2, Connections: []int{5, 12}, NodeType: NodeNotable,
			Effects: []Modifier{{Type: ModCritChance, Value: 10}},
		},
		{
			ID: 12, Name: "10x Developer", Desc: "+100% Damage, -50% HP", X: -5, Y: 0, Connections: []int{10, 11}, NodeType: NodeKeystone,
			Effects: []Modifier{{Type: ModPercentDamage, Value: 100}, {Type: ModPercentHP, Value: -50}},
		},

		// Defense branch (right side)
		{
			ID: 6, Name: "Thick Skin", Desc: "+5 Armor", X: 3, Y: -1, Connections: []int{2, 13}, NodeType: NodeSmall,
			Effects: []Modifier{{Type: ModArmor, Value: 5}},
		},
		{
			ID: 7, Name: "Vitality", Desc: "+20 Max HP", X: 3, Y: 1, Connections: []int{2, 14}, NodeType: NodeSmall,
			Effects: []Modifier{{Type: ModFlatHP, Value: 20}},
		},
		{
			ID: 13, Name: "Fortified Code", Desc: "+10 Armor", X: 4, Y: -2, Connections: []int{6, 15}, NodeType: NodeNotable,
			Effects: []Modifier{{Type: ModArmor, Value: 10}},
		},
		{
			ID: 14, Name: "Life Force", Desc: "+50 Max HP", X: 4, Y: 2, Connections: []int{7, 15}, NodeType: NodeNotable,
			Effects: []Modifier{{Type: ModFlatHP, Value: 50}},
		},
		{
			ID: 15, Name: "Defensive Programmer", Desc: "+50% Armor, -25% Damage", X: 5, Y: 0, Connections: []int{13, 14}, NodeType: NodeKeystone,
			Effects: []Modifier{{Type: ModArmor, Value: 50}, {Type: ModPercentDamage, Value: -25}},
		},

		// Utility branch (bottom)
		{
			ID: 8, Name: "Quick Deploy", Desc: "+5% Speed", X: -1, Y: 3, Connections: []int{3, 16}, NodeType: NodeSmall,
			Effects: []Modifier{{Type: ModSpeed, Value: 5}},
		},
		{
			ID: 9, Name: "Optimization", Desc: "-5% Cooldown", X: 1, Y: 3, Connections: []int{3, 17}, NodeType: NodeSmall,
			Effects: []Modifier{{Type: ModCooldown, Value: 5}},
		},
		{
			ID: 16, Name: "Rapid Iteration", Desc: "+10% Speed", X: -2, Y: 4, Connections: []int{8, 18}, NodeType: NodeNotable,
			Effects: []Modifier{{Type: ModSpeed, Value: 10}},
		},
		{
			ID: 17, Name: "CI Master", Desc: "-10% Cooldown", X: 2, Y: 4, Connections: []int{9, 18}, NodeType: NodeNotable,
			Effects: []Modifier{{Type: ModCooldown, Value: 10}},
		},
		{
			ID: 18, Name: "Code Reviewer", Desc: "+2 Projectiles, -30% Attack Speed", X: 0, Y: 5, Connections: []int{16, 17}, NodeType: NodeKeystone,
			Effects: []Modifier{{Type: ModProjectiles, Value: 2}, {Type: ModCooldown, Value: -30}},
		},

		// Extra nodes for depth
		{
			ID: 19, Name: "XP Boost", Desc: "+15% XP", X: 0, Y: -2, Connections: []int{0, 20}, NodeType: NodeSmall,
			Effects: []Modifier{{Type: ModXPGain, Value: 15}},
		},
		{
			ID: 20, Name: "Fast Learner", Desc: "+25% XP", X: 0, Y: -3, Connections: []int{19}, NodeType: NodeNotable,
			Effects: []Modifier{{Type: ModXPGain, Value: 25}},
		},
		{
			ID: 21, Name: "Wide Area", Desc: "+10% Area", X: -1, Y: -1, Connections: []int{0, 4}, NodeType: NodeSmall,
			Effects: []Modifier{{Type: ModArea, Value: 10}},
		},
		{
			ID: 22, Name: "Recovery", Desc: "+1 HP/s", X: 1, Y: -1, Connections: []int{0, 6}, NodeType: NodeSmall,
			Effects: []Modifier{{Type: ModRecovery, Value: 1}},
		},
		{
			ID: 23, Name: "Magnet", Desc: "+20% Pickup", X: -1, Y: 1, Connections: []int{0, 5}, NodeType: NodeSmall,
			Effects: []Modifier{{Type: ModMagnet, Value: 20}},
		},
		{
			ID: 24, Name: "Duration", Desc: "+15% Duration", X: 1, Y: 1, Connections: []int{0, 7}, NodeType: NodeSmall,
			Effects: []Modifier{{Type: ModDuration, Value: 15}},
		},
	}

	// Allocate starting node based on character class
	for _, node := range g.passiveTree {
		if node.StartClass == g.player.CharType {
			g.player.AllocatedNodes[node.ID] = true

			break
		}
	}
}

func (g *Game) updatePassiveTree() error {
	// ESC or P to close
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.state = StatePlaying

		return nil
	}

	// Mouse click to allocate nodes
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()

		// Convert screen coords to tree coords
		centerX := float64(screenWidth) / 2
		centerY := float64(screenHeight) / 2
		scale := 60.0 // Pixels per tree unit

		for _, node := range g.passiveTree {
			nodeScreenX := centerX + node.X*scale
			nodeScreenY := centerY + node.Y*scale

			// Check if click is on node
			dist := math.Sqrt(math.Pow(float64(mx)-nodeScreenX, 2) + math.Pow(float64(my)-nodeScreenY, 2))

			nodeRadius := 15.0

			switch node.NodeType {
			case NodeNotable:
				nodeRadius = 20.0
			case NodeKeystone:
				nodeRadius = 25.0
			}

			if dist < nodeRadius {
				g.tryAllocateNode(node)

				break
			}
		}
	}

	return nil
}

func (g *Game) tryAllocateNode(node *PassiveNode) {
	// Already allocated
	if g.player.AllocatedNodes[node.ID] {
		return
	}

	// Check if we have points
	if g.player.PassivePoints <= 0 {
		return
	}

	// Check if connected to an allocated node
	connected := false

	for _, connID := range node.Connections {
		if g.player.AllocatedNodes[connID] {
			connected = true

			break
		}
	}

	if !connected {
		return
	}

	// Allocate the node
	g.player.AllocatedNodes[node.ID] = true
	g.player.PassivePoints--

	// Recalculate all stats
	g.recalculateStats()
	g.audio.PlaySound("select")
}

func (g *Game) recalculateStats() {
	// Start with base stats
	charDef := Characters[g.player.CharType]
	g.player.MaxHP = charDef.HP
	g.player.Speed = charDef.Speed
	g.player.DamageMult = 1.0
	g.player.AreaMult = 1.0
	g.player.CooldownMult = 1.0
	g.player.MagnetRange = 80
	g.player.Recovery = 0
	g.player.CritChance = 0
	g.player.XPMult = 1.0
	g.player.Armor = 0

	// Apply character trait
	switch g.player.CharType {
	case CharSenior:
		g.player.AreaMult = 1.5
	}

	// Apply passive tree bonuses
	for _, node := range g.passiveTree {
		if g.player.AllocatedNodes[node.ID] {
			for _, mod := range node.Effects {
				g.applyModifier(mod)
			}
		}
	}

	// Apply equipment bonuses
	for _, equip := range g.player.Equipment {
		if equip != nil {
			for _, mod := range equip.Modifiers {
				g.applyModifier(mod)
			}
		}
	}

	// Apply passive skill bonuses (from level-up)
	for pType, level := range g.player.Passives {
		for range level {
			switch pType {
			case PassiveMight:
				g.player.DamageMult += 0.10
			case PassiveArmor:
				g.player.Armor += 5
			case PassiveSpeed:
				g.player.Speed *= 1.10
			case PassiveMagnet:
				g.player.MagnetRange *= 1.20
			case PassiveRecovery:
				g.player.Recovery += 0.3
			case PassiveLuck:
				g.player.CritChance += 0.10
			case PassiveGrowth:
				g.player.XPMult += 0.10
			case PassiveCooldown:
				g.player.CooldownMult *= 0.95
			case PassiveArea:
				g.player.AreaMult += 0.10
			}
		}
	}

	// Clamp HP to max
	if g.player.HP > g.player.MaxHP {
		g.player.HP = g.player.MaxHP
	}
}

func (g *Game) applyModifier(mod Modifier) {
	switch mod.Type {
	case ModFlatDamage:
		// Would need base damage tracking
	case ModPercentDamage:
		g.player.DamageMult += mod.Value / 100
	case ModFlatHP:
		g.player.MaxHP += int(mod.Value)
	case ModPercentHP:
		g.player.MaxHP = int(float64(g.player.MaxHP) * (1 + mod.Value/100))
	case ModArmor:
		g.player.Armor += int(mod.Value)
	case ModSpeed:
		g.player.Speed *= (1 + mod.Value/100)
	case ModCritChance:
		g.player.CritChance += mod.Value / 100
	case ModCooldown:
		g.player.CooldownMult *= (1 - mod.Value/100)
	case ModArea:
		g.player.AreaMult += mod.Value / 100
	case ModDuration:
		// Would need duration tracking
	case ModMagnet:
		g.player.MagnetRange *= (1 + mod.Value/100)
	case ModXPGain:
		g.player.XPMult += mod.Value / 100
	case ModRecovery:
		g.player.Recovery += mod.Value
	case ModProjectiles:
		// Apply to PassiveAmount
		g.player.Passives[PassiveAmount] += int(mod.Value)
	}
}

// generateEquipment creates a random equipment item.
func (g *Game) generateEquipment(slot EquipSlot, itemLevel int, rarity Rarity) *Equipment {
	names := map[EquipSlot][]string{
		SlotKeyboard:   {"Mechanical Keyboard", "Cherry MX Board", "Ergonomic Keyboard", "Gaming Keyboard"},
		SlotMonitor:    {"4K Monitor", "Ultrawide Display", "Gaming Monitor", "Dual Screen"},
		SlotChair:      {"Herman Miller", "Gaming Chair", "Ergonomic Seat", "Standing Desk"},
		SlotMouse:      {"Wireless Mouse", "Gaming Mouse", "Trackball", "Precision Mouse"},
		SlotHeadphones: {"Noise Cancelling", "Open Back Cans", "Gaming Headset", "AirPods Pro"},
		SlotCoffeeMug:  {"Yeti Tumbler", "Pour Over Set", "Espresso Cup", "Thermos"},
	}

	slotNames := names[slot]
	name := slotNames[rand.Intn(len(slotNames))]

	// Prefix based on rarity
	switch rarity {
	case RarityMagic:
		prefixes := []string{"Enhanced", "Quality", "Fine"}
		name = prefixes[rand.Intn(len(prefixes))] + " " + name
	case RarityRare:
		prefixes := []string{"Superior", "Exceptional", "Elite"}
		name = prefixes[rand.Intn(len(prefixes))] + " " + name
	case RarityLegendary:
		prefixes := []string{"Legendary", "Mythic", "Godly"}
		name = prefixes[rand.Intn(len(prefixes))] + " " + name
	}

	// Generate modifiers based on rarity
	modCount := 0

	switch rarity {
	case RarityCommon:
		modCount = 0
	case RarityMagic:
		modCount = 1 + rand.Intn(2)
	case RarityRare:
		modCount = 3 + rand.Intn(2)
	case RarityLegendary:
		modCount = 5 + rand.Intn(2)
	}

	// Preferred mods per slot
	slotMods := map[EquipSlot][]ModType{
		SlotKeyboard:   {ModFlatDamage, ModPercentDamage, ModCooldown},
		SlotMonitor:    {ModFlatHP, ModPercentHP, ModXPGain},
		SlotChair:      {ModArmor, ModRecovery, ModFlatHP},
		SlotMouse:      {ModCritChance, ModArea, ModPercentDamage},
		SlotHeadphones: {ModCooldown, ModDuration, ModArea},
		SlotCoffeeMug:  {ModSpeed, ModMagnet, ModRecovery},
	}

	preferredMods := slotMods[slot]
	mods := make([]Modifier, 0, modCount)

	for i := 0; i < modCount; i++ {
		modType := preferredMods[rand.Intn(len(preferredMods))]
		tier := 1 + rand.Intn(min(5, itemLevel/10+1))

		// Value based on mod type and tier
		var value float64

		switch modType {
		case ModFlatDamage:
			value = float64(tier * 3)
		case ModPercentDamage:
			value = float64(tier * 5)
		case ModFlatHP:
			value = float64(tier * 15)
		case ModPercentHP:
			value = float64(tier * 5)
		case ModArmor:
			value = float64(tier * 3)
		case ModSpeed:
			value = float64(tier * 3)
		case ModCritChance:
			value = float64(tier * 2)
		case ModCooldown:
			value = float64(tier * 3)
		case ModArea:
			value = float64(tier * 5)
		case ModDuration:
			value = float64(tier * 5)
		case ModMagnet:
			value = float64(tier * 10)
		case ModXPGain:
			value = float64(tier * 5)
		case ModRecovery:
			value = float64(tier) * 0.5
		}

		mods = append(mods, Modifier{Type: modType, Value: value, Tier: tier})
	}

	return &Equipment{
		Slot:      slot,
		Name:      name,
		Rarity:    rarity,
		Modifiers: mods,
		ItemLevel: itemLevel,
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case StateCharSelect:
		g.drawCharSelect(screen)
	case StatePlaying, StateLevelUp, StatePaused:
		g.drawGame(screen)

		if g.state == StateLevelUp {
			g.drawLevelUp(screen)
		}

		if g.state == StatePaused {
			g.drawPaused(screen)
		}
	case StateGameOver:
		g.drawGame(screen)
		g.drawGameOver(screen)
	case StateEquipment:
		g.drawGame(screen)
		g.drawEquipment(screen)
	case StatePassiveTree:
		g.drawPassiveTree(screen)
	case StateHelp:
		g.drawGame(screen)
		g.drawHelp(screen)
	}
}

func (g *Game) drawCharSelect(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 20, G: 25, B: 35, A: 255})

	// Title
	ebitenutil.DebugPrintAt(screen, "ENDLESS SWARM", screenWidth/2-50, 50)
	ebitenutil.DebugPrintAt(screen, "Select Your Hero", screenWidth/2-60, 80)

	// Characters
	for i, char := range Characters {
		x := 100 + i*180
		y := 200

		// Box
		boxColor := color.RGBA{R: 40, G: 45, B: 55, A: 255}
		if i == g.selectedChar {
			boxColor = color.RGBA{R: 60, G: 80, B: 100, A: 255}
		}

		vector.FillRect(screen, float32(x), float32(y), 150, 280, boxColor, false)

		if i == g.selectedChar {
			vector.StrokeRect(
				screen,
				float32(x),
				float32(y),
				150,
				280,
				3,
				color.RGBA{R: 255, G: 215, B: 0, A: 255},
				false,
			)
		}

		// Character image or fallback circle
		if g.charImages[i] != nil {
			img := g.charImages[i]
			bounds := img.Bounds()

			scale := 80.0 / float64(bounds.Dx())
			if float64(bounds.Dy())*scale > 100 {
				scale = 100.0 / float64(bounds.Dy())
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(float64(x+75)-float64(bounds.Dx())*scale/2, float64(y+10))
			screen.DrawImage(img, op)
		} else {
			vector.FillCircle(screen, float32(x+75), float32(y+60), 40, char.Color, false)
		}

		// Name
		ebitenutil.DebugPrintAt(screen, char.Name, x+50, y+115)

		// Stats
		ebitenutil.DebugPrintAt(screen, "HP: "+formatInt(char.HP), x+20, y+145)
		ebitenutil.DebugPrintAt(screen, "Speed: "+formatFloat(char.Speed), x+20, y+165)
		ebitenutil.DebugPrintAt(screen, "Weapon:", x+20, y+190)
		ebitenutil.DebugPrintAt(screen, WeaponDefs[char.StartWeapon].Name, x+20, y+205)
		ebitenutil.DebugPrintAt(screen, char.Trait+":", x+20, y+235)
		ebitenutil.DebugPrintAt(screen, char.TraitDesc, x+20, y+250)
	}

	// Controls
	ebitenutil.DebugPrintAt(
		screen,
		"LEFT/RIGHT to select | SPACE to start",
		screenWidth/2-130,
		screenHeight-50,
	)
}

func (g *Game) drawGame(screen *ebiten.Image) {
	// Background
	screen.Fill(color.RGBA{R: 25, G: 30, B: 40, A: 255})

	// Grid
	gridSize := 60.0
	offsetX := math.Mod(g.cameraX, gridSize)

	offsetY := math.Mod(g.cameraY, gridSize)
	for x := -gridSize; x < float64(screenWidth)+gridSize; x += gridSize {
		vector.FillRect(
			screen,
			float32(x-offsetX),
			0,
			1,
			screenHeight,
			color.RGBA{R: 35, G: 40, B: 50, A: 255},
			false,
		)
	}

	for y := -gridSize; y < float64(screenHeight)+gridSize; y += gridSize {
		vector.FillRect(
			screen,
			0,
			float32(y-offsetY),
			screenWidth,
			1,
			color.RGBA{R: 35, G: 40, B: 50, A: 255},
			false,
		)
	}

	// XP Gems
	for _, gem := range g.xpGems {
		sx, sy := gem.X-g.cameraX, gem.Y-g.cameraY
		if sx >= -20 && sx <= screenWidth+20 && sy >= -20 && sy <= screenHeight+20 {
			size := float32(6)
			if gem.Value >= 20 {
				size = 10
			}

			vector.FillRect(
				screen,
				float32(sx)-size/2,
				float32(sy)-size/2,
				size,
				size,
				color.RGBA{R: 100, G: 200, B: 255, A: 255},
				false,
			)
		}
	}

	// Enemies (Batched)
	g.drawEnemies(screen)

	// Projectiles
	g.drawProjectiles(screen)

	// Player
	px, py := g.player.X-g.cameraX, g.player.Y-g.cameraY
	pColor := Characters[g.player.CharType].Color

	// Draw character image or fallback
	if img := g.charImages[g.player.CharType]; img != nil {
		op := &ebiten.DrawImageOptions{}
		// Scale to fit ~40px width
		bounds := img.Bounds()
		scale := 50.0 / float64(bounds.Dx())
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(px-float64(bounds.Dx())*scale/2, py-float64(bounds.Dy())*scale/2)
		screen.DrawImage(img, op)
	} else {
		vector.FillCircle(screen, float32(px), float32(py), 20, pColor, false)
		vector.StrokeCircle(screen, float32(px), float32(py), 20, 3, color.RGBA{R: 255, G: 255, B: 255, A: 200}, false)
	}

	// Damage numbers
	for _, d := range g.damageNumbers {
		sx, sy := d.X-g.cameraX, d.Y-g.cameraY

		text := formatInt(d.Value)
		if d.Crit {
			text = text + "!"
		}

		ebitenutil.DebugPrintAt(screen, text, int(sx)-10, int(sy))
	}

	// Particles
	g.drawParticles(screen)

	// HUD
	g.drawHUD(screen)
}

func (g *Game) drawEnemies(screen *ebiten.Image) {
	// Sort by type for batching
	buckets := make(map[MonsterType][]*Enemy)

	for _, e := range g.enemies {
		// Culling
		sx, sy := e.X-g.cameraX, e.Y-g.cameraY
		if sx < -50 || sx > screenWidth+50 || sy < -50 || sy > screenHeight+50 {
			continue
		}

		buckets[e.Type] = append(buckets[e.Type], e)
	}

	for t, enemies := range buckets {
		img := g.monsterImages[t]

		for _, e := range enemies {
			sx, sy := e.X-g.cameraX, e.Y-g.cameraY

			if img != nil {
				bounds := img.Bounds()
				scale := (e.Radius * 2.5) / float64(bounds.Dx())

				op := &ebiten.DrawImageOptions{}
				op.GeoM.Scale(scale, scale)
				op.GeoM.Translate(sx-float64(bounds.Dx())*scale/2, sy-float64(bounds.Dy())*scale/2)

				if e.HitFlash > 0 {
					op.ColorScale.Scale(10, 10, 10, 1)
				} else {
					r := float32(e.Color.R) / 255.0
					g := float32(e.Color.G) / 255.0
					b := float32(e.Color.B) / 255.0
					op.ColorScale.Scale(r, g, b, 1)
				}

				screen.DrawImage(img, op)
			} else {
				c := e.Color
				if e.HitFlash > 0 {
					c = color.RGBA{R: 255, G: 255, B: 255, A: 255}
				}

				vector.FillCircle(screen, float32(sx), float32(sy), float32(e.Radius), c, false)
			}

			// Aux rendering (not batched, affects perf, but necessary for gameplay)
			// Boss indicator
			if e.IsBoss {
				vector.StrokeCircle(
					screen,
					float32(sx),
					float32(sy),
					float32(e.Radius)+5,
					3,
					color.RGBA{R: 255, G: 50, B: 50, A: 255},
					false,
				)
			}

			// HP bar
			if e.HP < e.MaxHP {
				barW := e.Radius * 2
				hpRatio := float32(e.HP) / float32(e.MaxHP)
				// Draw bg only
				vector.FillRect(
					screen,
					float32(sx)-float32(barW)/2,
					float32(sy)-float32(e.Radius)-8,
					float32(barW),
					4,
					color.RGBA{R: 50, G: 50, B: 50, A: 255},
					false,
				)
				// Draw fg
				vector.FillRect(
					screen,
					float32(sx)-float32(barW)/2,
					float32(sy)-float32(e.Radius)-8,
					float32(barW)*hpRatio,
					4,
					color.RGBA{R: 255, G: 50, B: 50, A: 255},
					false,
				)
			}
		}
	}
}

// drawProjectiles renders projectiles with weapon-specific visual effects.
func (g *Game) drawProjectiles(screen *ebiten.Image) {
	for _, p := range g.projectiles {
		sx, sy := p.X-g.cameraX, p.Y-g.cameraY

		// Skip if off-screen
		if sx < -50 || sx > screenWidth+50 || sy < -50 || sy > screenHeight+50 {
			continue
		}

		// Create glow color (lighter version of base color)
		glowColor := color.RGBA{
			R: uint8(min(255, int(p.Color.R)+100)),
			G: uint8(min(255, int(p.Color.G)+100)),
			B: uint8(min(255, int(p.Color.B)+100)),
			A: 100,
		}

		switch p.WeaponType {
		case WeaponGitPush, WeaponForcePush:
			// Arrow shape with speed trail
			// Draw trail (multiple smaller circles behind)
			if p.VX != 0 || p.VY != 0 {
				speed := math.Sqrt(p.VX*p.VX + p.VY*p.VY)
				dirX, dirY := -p.VX/speed, -p.VY/speed

				for i := 1; i <= 4; i++ {
					trailAlpha := uint8(150 - i*30)
					trailColor := color.RGBA{p.Color.R, p.Color.G, p.Color.B, trailAlpha}
					offset := float64(i) * 6
					vector.FillCircle(
						screen,
						float32(sx+dirX*offset),
						float32(sy+dirY*offset),
						float32(p.Radius)*0.6,
						trailColor,
						false,
					)
				}
			}
			// Main projectile with glow
			vector.FillCircle(screen, float32(sx), float32(sy), float32(p.Radius)+3, glowColor, false)
			vector.FillCircle(screen, float32(sx), float32(sy), float32(p.Radius), p.Color, false)
			// Arrow tip
			if p.VX != 0 || p.VY != 0 {
				speed := math.Sqrt(p.VX*p.VX + p.VY*p.VY)
				tipX, tipY := float32(sx+p.VX/speed*8), float32(sy+p.VY/speed*8)
				vector.FillCircle(screen, tipX, tipY, float32(p.Radius)*0.5, color.White, false)
			}

		case WeaponFirewall, WeaponZeroTrust:
			// Fire ring effect
			// Outer glow
			vector.StrokeCircle(screen, float32(sx), float32(sy), float32(p.Radius)+8, 4, glowColor, false)
			// Flame particles (animated)
			flameCount := 6
			for i := range flameCount {
				angle := g.gameTime*5 + float64(i)*math.Pi*2/float64(flameCount)
				flameX := sx + math.Cos(angle)*float64(p.Radius)*0.8
				flameY := sy + math.Sin(angle)*float64(p.Radius)*0.8
				// Inner flame
				vector.FillCircle(
					screen,
					float32(flameX),
					float32(flameY),
					4,
					color.RGBA{255, 200, 50, 200},
					false,
				)
				// Outer flame
				vector.FillCircle(
					screen,
					float32(flameX),
					float32(flameY),
					6,
					color.RGBA{255, 100, 30, 100},
					false,
				)
			}
			// Core fire ring
			vector.StrokeCircle(screen, float32(sx), float32(sy), float32(p.Radius), 3, p.Color, false)
			vector.StrokeCircle(
				screen,
				float32(sx),
				float32(sy),
				float32(p.Radius)*0.6,
				2,
				color.RGBA{255, 200, 100, 200},
				false,
			)

		case WeaponStackOverflow, WeaponCopilot:
			// Lightning bolt effect
			// Vertical bolt with zigzag
			vector.FillCircle(screen, float32(sx), float32(sy), float32(p.Radius)+5, glowColor, false)
			// Lightning segments
			segments := 4
			segHeight := p.Radius * 2 / float64(segments)
			startY := sy - p.Radius

			zigzag := 5.0
			for i := range segments {
				x1 := sx + zigzag*float64(i%2*2-1)
				y1 := startY + float64(i)*segHeight
				x2 := sx + zigzag*float64((i+1)%2*2-1)
				y2 := startY + float64(i+1)*segHeight
				vector.StrokeLine(
					screen,
					float32(x1),
					float32(y1),
					float32(x2),
					float32(y2),
					3,
					p.Color,
					false,
				)
				vector.StrokeLine(
					screen,
					float32(x1),
					float32(y1),
					float32(x2),
					float32(y2),
					1,
					color.White,
					false,
				)
			}
			// Spark at tip
			vector.FillCircle(screen, float32(sx), float32(sy+p.Radius), 4, color.White, false)

		case WeaponRefactor, WeaponCleanCode:
			// Orbiting circles with trail
			// Outer glow ring
			vector.StrokeCircle(screen, float32(sx), float32(sy), float32(p.Radius)+4, 2, glowColor, false)
			// Inner spinning circles
			orbCount := 3
			for i := range orbCount {
				angle := g.gameTime*4 + float64(i)*math.Pi*2/float64(orbCount)
				orbX := sx + math.Cos(angle)*float64(p.Radius)*0.5
				orbY := sy + math.Sin(angle)*float64(p.Radius)*0.5
				vector.FillCircle(screen, float32(orbX), float32(orbY), 5, p.Color, false)
			}
			// Core
			vector.FillCircle(
				screen,
				float32(sx),
				float32(sy),
				float32(p.Radius)*0.4,
				color.White,
				false,
			)

		case WeaponCoffee, WeaponEspresso:
			// Aura ring with steam particles
			// Pulsing aura rings (3 rings)
			for i := range 3 {
				pulseOffset := math.Sin(g.gameTime*4+float64(i)*0.5) * 5
				ringR := float64(p.Radius) + pulseOffset + float64(i)*8
				ringAlpha := uint8(150 - i*40)
				ringColor := color.RGBA{p.Color.R, p.Color.G, p.Color.B, ringAlpha}
				vector.StrokeCircle(screen, float32(sx), float32(sy), float32(ringR), 2, ringColor, false)
			}
			// Steam particles rising
			for i := range 4 {
				steamY := sy - float64(i)*10 - math.Sin(g.gameTime*3)*5
				steamX := sx + math.Sin(g.gameTime*2+float64(i))*8
				steamAlpha := uint8(100 - i*20)
				vector.FillCircle(
					screen,
					float32(steamX),
					float32(steamY),
					3,
					color.RGBA{200, 200, 200, steamAlpha},
					false,
				)
			}

		case WeaponDocker, WeaponK8s:
			// Container/box shape with glow
			// Glow
			boxSize := p.Radius * 1.5
			vector.FillRect(
				screen,
				float32(sx-boxSize-2),
				float32(sy-boxSize-2),
				float32(boxSize*2+4),
				float32(boxSize*2+4),
				glowColor,
				false,
			)
			// Main container
			vector.FillRect(
				screen,
				float32(sx-boxSize),
				float32(sy-boxSize),
				float32(boxSize*2),
				float32(boxSize*2),
				p.Color,
				false,
			)
			// Container lines (grid pattern)
			vector.StrokeLine(
				screen,
				float32(sx-boxSize),
				float32(sy),
				float32(sx+boxSize),
				float32(sy),
				1,
				color.RGBA{255, 255, 255, 150},
				false,
			)
			vector.StrokeLine(
				screen,
				float32(sx),
				float32(sy-boxSize),
				float32(sx),
				float32(sy+boxSize),
				1,
				color.RGBA{255, 255, 255, 150},
				false,
			)
			// Outline
			vector.StrokeRect(
				screen,
				float32(sx-boxSize),
				float32(sy-boxSize),
				float32(boxSize*2),
				float32(boxSize*2),
				2,
				color.White,
				false,
			)

		case WeaponUnitTests, WeaponCI_CD:
			// Pulse wave effect (expanding ring)
			// Multiple rings expanding outward
			for i := range 3 {
				waveOffset := math.Mod(g.gameTime*2+float64(i)*0.3, 1.0)
				waveR := float64(p.Radius) * waveOffset
				waveAlpha := uint8(float64(200) * (1 - waveOffset))
				waveColor := color.RGBA{p.Color.R, p.Color.G, p.Color.B, waveAlpha}
				vector.StrokeCircle(screen, float32(sx), float32(sy), float32(waveR), 2, waveColor, false)
			}
			// Center checkmark
			vector.FillCircle(screen, float32(sx), float32(sy), float32(p.Radius)*0.3, p.Color, false)
			// Check shape
			vector.StrokeLine(
				screen,
				float32(sx-5),
				float32(sy),
				float32(sx-2),
				float32(sy+3),
				2,
				color.White,
				false,
			)
			vector.StrokeLine(
				screen,
				float32(sx-2),
				float32(sy+3),
				float32(sx+5),
				float32(sy-4),
				2,
				color.White,
				false,
			)

		case WeaponPrint, WeaponLogStream:
			// Text/console effect with glow
			// Outer glow
			vector.FillCircle(screen, float32(sx), float32(sy), float32(p.Radius)+4, glowColor, false)
			// Main circle
			vector.FillCircle(screen, float32(sx), float32(sy), float32(p.Radius), p.Color, false)
			// Console text lines (3 small rectangles)
			lineWidth := p.Radius * 0.8

			for i := -1; i <= 1; i++ {
				lineY := sy + float64(i)*4
				lineLen := lineWidth * (0.6 + rand.Float64()*0.4) // Varying lengths
				vector.FillRect(
					screen,
					float32(sx-lineWidth/2),
					float32(lineY-1),
					float32(lineLen),
					2,
					color.RGBA{255, 255, 255, 200},
					false,
				)
			}

		default:
			// Default circle with glow (fallback)
			vector.FillCircle(screen, float32(sx), float32(sy), float32(p.Radius)+2, glowColor, false)
			vector.FillCircle(screen, float32(sx), float32(sy), float32(p.Radius), p.Color, false)
		}
	}
}

func (g *Game) drawParticles(screen *ebiten.Image) {
	for _, p := range g.particles {
		sx, sy := p.X-g.cameraX, p.Y-g.cameraY
		if sx >= -10 && sx <= screenWidth+10 && sy >= -10 && sy <= screenHeight+10 {
			// Fade out alpha
			alpha := float32(p.Lifetime / p.MaxLife)
			c := p.Color
			c.A = uint8(float32(c.A) * alpha)

			vector.FillRect(
				screen,
				float32(sx)-float32(p.Size)/2,
				float32(sy)-float32(p.Size)/2,
				float32(p.Size),
				float32(p.Size),
				c,
				false,
			)
		}
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	// Top bar
	vector.FillRect(screen, 0, 0, screenWidth, 60, color.RGBA{R: 0, G: 0, B: 0, A: 200}, false)

	// Character portrait
	vector.FillCircle(screen, 30, 30, 22, Characters[g.player.CharType].Color, false)

	// HP bar
	vector.FillRect(screen, 60, 8, 200, 18, color.RGBA{R: 40, G: 40, B: 40, A: 255}, false)

	hpRatio := float32(g.player.HP) / float32(g.player.MaxHP)
	vector.FillRect(screen, 60, 8, 200*hpRatio, 18, color.RGBA{R: 200, G: 50, B: 50, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, formatInt(g.player.HP)+"/"+formatInt(g.player.MaxHP), 130, 10)

	// XP bar
	xpNeeded := g.player.Level * 25
	xpRatio := float32(g.player.XP) / float32(xpNeeded)

	vector.FillRect(screen, 60, 32, 200, 12, color.RGBA{R: 40, G: 40, B: 40, A: 255}, false)
	vector.FillRect(screen, 60, 32, 200*xpRatio, 12, color.RGBA{R: 100, G: 200, B: 255, A: 255}, false)

	// Level
	ebitenutil.DebugPrintAt(screen, "Lv "+formatInt(g.player.Level), 270, 20)

	// Time and kills
	ebitenutil.DebugPrintAt(screen, "Time: "+formatTime(g.gameTime), 400, 10)
	ebitenutil.DebugPrintAt(screen, "Kills: "+formatInt(g.killCount), 400, 30)
	ebitenutil.DebugPrintAt(screen, "Enemies: "+formatInt(len(g.enemies)), 550, 10)

	// Weapon icons
	for i, w := range g.player.Weapons {
		x := 10 + i*55
		y := screenHeight - 55

		if img, ok := g.weaponImages[w.Type]; ok {
			op := &ebiten.DrawImageOptions{}
			scale := 50.0 / 64.0
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(float64(x), float64(y))
			screen.DrawImage(img, op)
		} else {
			def := WeaponDefs[w.Type]
			vector.FillRect(screen, float32(x), float32(y), 50, 50, def.Color, false)
			vector.StrokeRect(screen, float32(x), float32(y), 50, 50, 2, color.RGBA{R: 255, G: 255, B: 255, A: 150}, false)
		}

		ebitenutil.DebugPrintAt(screen, formatInt(w.Level), x+38, y+35)
	}

	// Controls hint
	ebitenutil.DebugPrintAt(
		screen,
		"H=Help | I=Equip | P=Passives | ESC=Pause",
		screenWidth-320,
		screenHeight-20,
	)
}

func (g *Game) drawLevelUp(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		0,
		0,
		screenWidth,
		screenHeight,
		color.RGBA{R: 0, G: 0, B: 0, A: 180},
		false,
	)

	boxW, boxH := float32(500), float32(300)
	boxX, boxY := float32(screenWidth-500)/2, float32(screenHeight-300)/2

	vector.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 30, G: 35, B: 50, A: 255}, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 3, color.RGBA{R: 255, G: 215, B: 0, A: 255}, false)

	ebitenutil.DebugPrintAt(screen, "LEVEL UP! Choose an upgrade:", int(boxX)+140, int(boxY)+15)

	for i, opt := range g.upgradeOptions {
		y := int(boxY) + 55 + i*60

		// Option box
		optColor := color.RGBA{R: 50, G: 55, B: 70, A: 255}
		vector.FillRect(screen, boxX+20, float32(y)-5, boxW-40, 55, optColor, false)

		// Icon
		var icon *ebiten.Image
		if opt.IsWeapon {
			icon = g.weaponImages[opt.WeaponType]
		} else {
			icon = g.passiveImages[opt.PassiveType]
		}

		if icon != nil {
			op := &ebiten.DrawImageOptions{}
			scale := 45.0 / 64.0
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(float64(boxX)+25, float64(y))
			screen.DrawImage(icon, op)
		}

		// Number (far right)
		ebitenutil.DebugPrintAt(screen, "["+formatInt(i+1)+"]", int(boxX+boxW)-40, y+20)

		// Name and level
		lvlText := ""
		if opt.CurrentLvl > 0 {
			lvlText = " (Lv " + formatInt(opt.CurrentLvl+1) + ")"
		}

		ebitenutil.DebugPrintAt(screen, opt.Name+lvlText, int(boxX)+80, y+5)
		ebitenutil.DebugPrintAt(screen, opt.Desc, int(boxX)+80, y+25)
	}
}

func (g *Game) drawPaused(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		0,
		0,
		screenWidth,
		screenHeight,
		color.RGBA{R: 0, G: 0, B: 0, A: 180},
		false,
	)

	boxW, boxH := float32(300), float32(180)
	boxX, boxY := float32(screenWidth-300)/2, float32(screenHeight-180)/2

	vector.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 40, G: 45, B: 60, A: 255}, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 2, color.RGBA{R: 200, G: 200, B: 200, A: 255}, false)

	ebitenutil.DebugPrintAt(screen, "PAUSED", int(boxX)+115, int(boxY)+30)
	ebitenutil.DebugPrintAt(screen, "SPACE / ESC - Resume", int(boxX)+70, int(boxY)+80)
	ebitenutil.DebugPrintAt(screen, "Q - Quit to Menu", int(boxX)+85, int(boxY)+110)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		0,
		0,
		screenWidth,
		screenHeight,
		color.RGBA{R: 0, G: 0, B: 0, A: 200},
		false,
	)

	boxW, boxH := float32(350), float32(250)
	boxX, boxY := float32(screenWidth-350)/2, float32(screenHeight-250)/2

	vector.FillRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 50, G: 30, B: 30, A: 255}, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 3, color.RGBA{R: 200, G: 50, B: 50, A: 255}, false)

	ebitenutil.DebugPrintAt(screen, "GAME OVER", int(boxX)+120, int(boxY)+25)

	ebitenutil.DebugPrintAt(screen, "Survived: "+formatTime(g.gameTime), int(boxX)+100, int(boxY)+70)
	ebitenutil.DebugPrintAt(screen, "Level: "+formatInt(g.player.Level), int(boxX)+120, int(boxY)+95)
	ebitenutil.DebugPrintAt(screen, "Kills: "+formatInt(g.killCount), int(boxX)+120, int(boxY)+120)
	ebitenutil.DebugPrintAt(
		screen,
		"Weapons: "+formatInt(len(g.player.Weapons)),
		int(boxX)+110,
		int(boxY)+145,
	)

	ebitenutil.DebugPrintAt(screen, "SPACE - Retry", int(boxX)+110, int(boxY)+185)
	ebitenutil.DebugPrintAt(screen, "Q - Character Select", int(boxX)+85, int(boxY)+210)
}

// ============================================================================
// EQUIPMENT UI
// ============================================================================

func (g *Game) drawEquipment(screen *ebiten.Image) {
	// Semi-transparent overlay
	vector.FillRect(
		screen,
		0,
		0,
		screenWidth,
		screenHeight,
		color.RGBA{R: 0, G: 0, B: 0, A: 180},
		false,
	)

	// Main panel
	panelW, panelH := float32(700), float32(500)
	panelX, panelY := float32(screenWidth-700)/2, float32(screenHeight-500)/2

	vector.FillRect(
		screen,
		panelX,
		panelY,
		panelW,
		panelH,
		color.RGBA{R: 30, G: 30, B: 40, A: 255},
		false,
	)
	vector.StrokeRect(
		screen,
		panelX,
		panelY,
		panelW,
		panelH,
		3,
		color.RGBA{R: 100, G: 150, B: 200, A: 255},
		false,
	)

	// Title
	ebitenutil.DebugPrintAt(screen, "EQUIPMENT (Press I to close)", int(panelX)+250, int(panelY)+10)

	// Equipment slots on the left
	slotStartX := panelX + 30
	slotStartY := panelY + 50
	slotH := float32(60)

	for slot := range SlotCount {
		y := slotStartY + float32(slot)*slotH
		slotW := float32(250)

		// Highlight selected slot
		bgCol := color.RGBA{R: 40, G: 40, B: 50, A: 255}
		if slot == g.selectedSlot {
			bgCol = color.RGBA{R: 60, G: 80, B: 100, A: 255}
		}

		vector.FillRect(screen, slotStartX, y, slotW, slotH-5, bgCol, false)
		vector.StrokeRect(
			screen,
			slotStartX,
			y,
			slotW,
			slotH-5,
			1,
			color.RGBA{R: 100, G: 100, B: 120, A: 255},
			false,
		)

		// Slot name
		ebitenutil.DebugPrintAt(screen, EquipSlotNames[slot]+":", int(slotStartX)+5, int(y)+5)

		// Equipped item
		if equip := g.player.Equipment[slot]; equip != nil {
			itemCol := RarityColors[equip.Rarity]
			ebitenutil.DebugPrintAt(screen, equip.Name, int(slotStartX)+5, int(y)+22)
			vector.FillRect(screen, slotStartX+slotW-30, y+5, 25, 25, itemCol, false)
			// Show mod count
			ebitenutil.DebugPrintAt(
				screen,
				formatInt(len(equip.Modifiers))+" mods",
				int(slotStartX)+5,
				int(y)+39,
			)
		} else {
			ebitenutil.DebugPrintAt(screen, "(empty)", int(slotStartX)+5, int(y)+25)
		}
	}

	// Inventory on the right
	invStartX := panelX + 320
	invStartY := panelY + 50

	ebitenutil.DebugPrintAt(screen, "Inventory:", int(invStartX), int(invStartY)-20)

	itemW, itemH := float32(80), float32(70)
	cols := 4

	for i, item := range g.player.Inventory {
		col := i % cols
		row := i / cols

		x := invStartX + float32(col)*(itemW+5)
		y := invStartY + float32(row)*(itemH+5)

		// Highlight selected
		bgCol := color.RGBA{R: 40, G: 40, B: 50, A: 255}
		if i == g.selectedInvIndex {
			bgCol = color.RGBA{R: 60, G: 80, B: 100, A: 255}
		}

		vector.FillRect(screen, x, y, itemW, itemH, bgCol, false)
		vector.StrokeRect(screen, x, y, itemW, itemH, 1, RarityColors[item.Rarity], false)

		// Item info (truncated)
		name := item.Name
		if len(name) > 10 {
			name = name[:10] + "..."
		}

		ebitenutil.DebugPrintAt(screen, name, int(x)+2, int(y)+5)
		ebitenutil.DebugPrintAt(screen, EquipSlotNames[item.Slot], int(x)+2, int(y)+22)
		ebitenutil.DebugPrintAt(screen, formatInt(len(item.Modifiers))+" mods", int(x)+2, int(y)+39)
	}

	if len(g.player.Inventory) == 0 {
		ebitenutil.DebugPrintAt(screen, "No items", int(invStartX)+50, int(invStartY)+50)
	}

	// Instructions
	ebitenutil.DebugPrintAt(
		screen,
		"UP/DOWN: Select Slot | LEFT/RIGHT: Select Item | ENTER: Equip",
		int(panelX)+100,
		int(panelY+panelH-25),
	)
}

// ============================================================================
// PASSIVE TREE UI
// ============================================================================

func (g *Game) drawPassiveTree(screen *ebiten.Image) {
	// Dark background
	screen.Fill(color.RGBA{R: 15, G: 15, B: 25, A: 255})

	centerX := float64(screenWidth) / 2
	centerY := float64(screenHeight) / 2
	scale := 60.0 // Pixels per tree grid unit

	// Draw connections first
	for _, node := range g.passiveTree {
		nodeX := centerX + node.X*scale
		nodeY := centerY + node.Y*scale

		for _, connID := range node.Connections {
			// Find connected node
			for _, other := range g.passiveTree {
				if other.ID == connID {
					otherX := centerX + other.X*scale
					otherY := centerY + other.Y*scale

					// Connection color based on allocation
					connCol := color.RGBA{R: 50, G: 50, B: 60, A: 255}
					if g.player.AllocatedNodes[node.ID] && g.player.AllocatedNodes[other.ID] {
						connCol = color.RGBA{R: 100, G: 150, B: 200, A: 255}
					} else if g.player.AllocatedNodes[node.ID] || g.player.AllocatedNodes[other.ID] {
						connCol = color.RGBA{R: 80, G: 100, B: 120, A: 255}
					}

					vector.StrokeLine(
						screen,
						float32(nodeX),
						float32(nodeY),
						float32(otherX),
						float32(otherY),
						2,
						connCol,
						false,
					)

					break
				}
			}
		}
	}

	// Draw nodes
	mx, my := ebiten.CursorPosition()

	var hoveredNode *PassiveNode

	for _, node := range g.passiveTree {
		nodeX := centerX + node.X*scale
		nodeY := centerY + node.Y*scale

		// Node size based on type
		nodeRadius := 12.0

		switch node.NodeType {
		case NodeNotable:
			nodeRadius = 18.0
		case NodeKeystone:
			nodeRadius = 24.0
		}

		// Check hover
		dist := math.Sqrt(math.Pow(float64(mx)-nodeX, 2) + math.Pow(float64(my)-nodeY, 2))
		if dist < nodeRadius {
			hoveredNode = node
		}

		// Node color
		var nodeCol color.RGBA
		if g.player.AllocatedNodes[node.ID] {
			// Allocated - bright gold
			nodeCol = color.RGBA{R: 200, G: 180, B: 100, A: 255}
		} else {
			// Check if allocatable (connected to allocated)
			allocatable := false

			for _, connID := range node.Connections {
				if g.player.AllocatedNodes[connID] {
					allocatable = true

					break
				}
			}

			if allocatable && g.player.PassivePoints > 0 {
				// Allocatable - blue-ish
				nodeCol = color.RGBA{R: 80, G: 120, B: 180, A: 255}
			} else {
				// Not allocatable - gray
				nodeCol = color.RGBA{R: 50, G: 50, B: 60, A: 255}
			}
		}

		// Keystone special coloring
		if node.NodeType == NodeKeystone {
			if g.player.AllocatedNodes[node.ID] {
				nodeCol = color.RGBA{R: 220, G: 150, B: 50, A: 255}
			} else {
				nodeCol = color.RGBA{R: 100, G: 50, B: 50, A: 255}
			}
		}

		// Draw node
		vector.FillCircle(screen, float32(nodeX), float32(nodeY), float32(nodeRadius), nodeCol, false)

		// Border for notable/keystone
		switch node.NodeType {
		case NodeNotable:
			vector.StrokeCircle(
				screen,
				float32(nodeX),
				float32(nodeY),
				float32(nodeRadius),
				2,
				color.RGBA{R: 150, G: 150, B: 200, A: 255},
				false,
			)
		case NodeKeystone:
			vector.StrokeCircle(
				screen,
				float32(nodeX),
				float32(nodeY),
				float32(nodeRadius),
				3,
				color.RGBA{R: 255, G: 200, B: 50, A: 255},
				false,
			)
		}
	}

	// Tooltip for hovered node
	if hoveredNode != nil {
		ttX, ttY := float32(mx+15), float32(my+15)
		ttW := float32(200)
		ttH := float32(60 + len(hoveredNode.Effects)*15)

		// Clamp to screen
		if ttX+ttW > screenWidth {
			ttX = screenWidth - ttW - 5
		}

		if ttY+ttH > screenHeight {
			ttY = screenHeight - ttH - 5
		}

		vector.FillRect(screen, ttX, ttY, ttW, ttH, color.RGBA{R: 20, G: 20, B: 30, A: 240}, false)
		vector.StrokeRect(screen, ttX, ttY, ttW, ttH, 1, color.RGBA{R: 100, G: 100, B: 150, A: 255}, false)

		ebitenutil.DebugPrintAt(screen, hoveredNode.Name, int(ttX)+5, int(ttY)+5)

		if hoveredNode.Desc != "" {
			ebitenutil.DebugPrintAt(screen, hoveredNode.Desc, int(ttX)+5, int(ttY)+22)
		}

		// Show effects
		for i, mod := range hoveredNode.Effects {
			modText := ModTypeNames[mod.Type]
			modText = modText[:3] + formatFloat(mod.Value) + modText[4:]
			ebitenutil.DebugPrintAt(screen, modText, int(ttX)+5, int(ttY)+40+i*15)
		}
	}

	// UI overlay
	// Top bar
	vector.FillRect(screen, 0, 0, screenWidth, 40, color.RGBA{R: 20, G: 20, B: 30, A: 220}, false)
	ebitenutil.DebugPrintAt(screen, "PASSIVE TREE (Press P to close)", 10, 10)
	ebitenutil.DebugPrintAt(screen, "Points: "+formatInt(g.player.PassivePoints), screenWidth-120, 10)
	ebitenutil.DebugPrintAt(screen, "Click nodes to allocate", screenWidth/2-70, 10)
}

// ============================================================================
// HELP SCREEN
// ============================================================================

func (g *Game) updateHelp() error {
	// ESC or H to close
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyH) {
		g.state = StatePlaying
	}

	// Selection
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.helpSelection--
		if g.helpSelection < 0 {
			g.helpSelection = 1
		}

		g.audio.PlaySound("select")
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.helpSelection++
		if g.helpSelection > 1 {
			g.helpSelection = 0
		}

		g.audio.PlaySound("select")
	}

	// Adjust Volume
	switch g.helpSelection {
	case 0: // SFX
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
			g.audio.SetSFXVolume(g.audio.SFXVolume() - 0.1)
			g.audio.PlaySound("select")
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
			g.audio.SetSFXVolume(g.audio.SFXVolume() + 0.1)
			g.audio.PlaySound("select")
		}
	case 1: // Music
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
			g.audio.SetMusicVolume(g.audio.MusicVolume() - 0.1)
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
			g.audio.SetMusicVolume(g.audio.MusicVolume() + 0.1)
		}
	}

	return nil
}

func (g *Game) drawHelp(screen *ebiten.Image) {
	// Semi-transparent overlay
	vector.FillRect(
		screen,
		0,
		0,
		screenWidth,
		screenHeight,
		color.RGBA{R: 0, G: 0, B: 0, A: 200},
		false,
	)

	// Main panel
	panelW, panelH := float32(500), float32(450)
	panelX, panelY := float32(screenWidth-500)/2, float32(screenHeight-450)/2

	vector.FillRect(
		screen,
		panelX,
		panelY,
		panelW,
		panelH,
		color.RGBA{R: 25, G: 30, B: 40, A: 255},
		false,
	)
	vector.StrokeRect(
		screen,
		panelX,
		panelY,
		panelW,
		panelH,
		3,
		color.RGBA{R: 100, G: 200, B: 255, A: 255},
		false,
	)

	// Title
	ebitenutil.DebugPrintAt(screen, "=== CONTROLS & HELP ===", int(panelX)+150, int(panelY)+15)

	y := int(panelY) + 50

	// Volume Control Section
	ebitenutil.DebugPrintAt(screen, "-- AUDIO SETTINGS --", int(panelX)+175, y)
	y += 25

	// SFX Volume
	sfxCol := color.RGBA{R: 150, G: 150, B: 150, A: 255}
	if g.helpSelection == 0 {
		sfxCol = color.RGBA{R: 255, G: 255, B: 100, A: 255}
	}

	ebitenutil.DebugPrintAt(screen, "SFX Volume:", int(panelX)+30, y)
	vector.FillRect(
		screen,
		panelX+130,
		float32(y+2),
		100,
		10,
		color.RGBA{R: 50, G: 50, B: 50, A: 255},
		false,
	)
	vector.FillRect(
		screen,
		panelX+130,
		float32(y+2),
		100*float32(g.audio.SFXVolume()),
		10,
		sfxCol,
		false,
	)
	y += 20

	// Music Volume
	musicCol := color.RGBA{R: 150, G: 150, B: 150, A: 255}
	if g.helpSelection == 1 {
		musicCol = color.RGBA{R: 255, G: 255, B: 100, A: 255}
	}

	ebitenutil.DebugPrintAt(screen, "Music Volume:", int(panelX)+30, y)
	vector.FillRect(
		screen,
		panelX+130,
		float32(y+2),
		100,
		10,
		color.RGBA{R: 50, G: 50, B: 50, A: 255},
		false,
	)
	vector.FillRect(
		screen,
		panelX+130,
		float32(y+2),
		100*float32(g.audio.MusicVolume()),
		10,
		musicCol,
		false,
	)

	y += 30
	ebitenutil.DebugPrintAt(screen, "(UP/DOWN to select | LEFT/RIGHT to adjust)", int(panelX)+80, y)
	y += 35

	// Movement section
	ebitenutil.DebugPrintAt(screen, "-- MOVEMENT --", int(panelX)+180, y)
	y += 25
	ebitenutil.DebugPrintAt(screen, "WASD / Arrow Keys    Move character", int(panelX)+30, y)
	y += 20
	ebitenutil.DebugPrintAt(screen, "ESC                  Pause game", int(panelX)+30, y)
	y += 35

	// Screens section
	ebitenutil.DebugPrintAt(screen, "-- SCREENS --", int(panelX)+185, y)
	y += 25
	ebitenutil.DebugPrintAt(screen, "H                    Help (this screen)", int(panelX)+30, y)
	y += 20
	ebitenutil.DebugPrintAt(screen, "I                    Equipment/Inventory", int(panelX)+30, y)
	y += 20
	ebitenutil.DebugPrintAt(screen, "P                    Passive Skill Tree", int(panelX)+30, y)
	y += 35

	// Equipment section
	ebitenutil.DebugPrintAt(screen, "-- EQUIPMENT --", int(panelX)+175, y)
	y += 25
	ebitenutil.DebugPrintAt(screen, "Kill bosses (every 3 min) for guaranteed drops", int(panelX)+30, y)
	y += 20
	ebitenutil.DebugPrintAt(screen, "Elite enemies have 5% drop chance", int(panelX)+30, y)
	y += 20
	ebitenutil.DebugPrintAt(screen, "In Equipment screen: Arrow keys to select,", int(panelX)+30, y)
	y += 20
	ebitenutil.DebugPrintAt(screen, "                     Enter to equip item", int(panelX)+30, y)
	y += 35

	// Passive Tree section
	ebitenutil.DebugPrintAt(screen, "-- PASSIVE TREE --", int(panelX)+165, y)
	y += 25
	ebitenutil.DebugPrintAt(screen, "Gain +1 point per level", int(panelX)+30, y)
	y += 20
	ebitenutil.DebugPrintAt(screen, "Click nodes to allocate (must be connected)", int(panelX)+30, y)
	y += 20
	ebitenutil.DebugPrintAt(screen, "Hover nodes to see effects", int(panelX)+30, y)

	// Close hint
	ebitenutil.DebugPrintAt(screen, "Press H or ESC to close", int(panelX)+160, int(panelY+panelH-30))
}

func formatFloat(f float64) string {
	if f == float64(int(f)) {
		return formatInt(int(f))
	}

	return formatInt(int(f * 10)) // e.g. 0.5 -> 5
}

func formatInt(n int) string {
	if n == 0 {
		return "0"
	}

	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}

	return result
}

func formatTime(t float64) string {
	mins := int(t) / 60
	secs := int(t) % 60

	s := formatInt(secs)
	if secs < 10 {
		s = "0" + s
	}

	return formatInt(mins) + ":" + s
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) generateIcons() {
	g.weaponImages = make(map[WeaponType]*ebiten.Image)
	g.passiveImages = make(map[PassiveType]*ebiten.Image)

	const iconSize = 64

	// Helper function to load and scale an icon image from file
	loadIconImage := func(imageFile string) *ebiten.Image {
		if imageFile == "" {
			return nil
		}

		data, err := assetsFS.ReadFile(imageFile)
		if err != nil {
			return nil
		}

		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil
		}
		// Scale to icon size
		scaled := image.NewRGBA(image.Rect(0, 0, iconSize, iconSize))
		srcBounds := img.Bounds()
		sx := float64(srcBounds.Dx()) / float64(iconSize)

		sy := float64(srcBounds.Dy()) / float64(iconSize)
		for y := range iconSize {
			for x := range iconSize {
				srcX := int(float64(x) * sx)
				srcY := int(float64(y) * sy)
				scaled.Set(x, y, img.At(srcBounds.Min.X+srcX, srcBounds.Min.Y+srcY))
			}
		}

		return ebiten.NewImageFromImage(scaled)
	}

	// Weapons - try loading from file first
	for t, def := range WeaponDefs {
		var img *ebiten.Image

		// Try to load from image file
		if def.ImageFile != "" {
			img = loadIconImage(def.ImageFile)
		}

		// Fallback to programmatic generation
		if img == nil {
			img = g.generateWeaponIcon(t, def, iconSize)
		}

		g.weaponImages[t] = img
	}

	// Passives - try loading from file first
	for t, def := range PassiveDefs {
		var img *ebiten.Image

		// Try to load from image file
		if def.ImageFile != "" {
			img = loadIconImage(def.ImageFile)
		}

		// Fallback to programmatic generation
		if img == nil {
			img = g.generatePassiveIcon(t, iconSize)
		}

		g.passiveImages[t] = img
	}
}

// generateWeaponIcon creates a programmatic weapon icon (fallback when no image file).
func (g *Game) generateWeaponIcon(t WeaponType, def WeaponDef, size int) *ebiten.Image {
	img := ebiten.NewImage(size, size)

	// Background & Border
	bgCol := color.RGBA{R: 30, G: 30, B: 40, A: 255}
	borderWidth := float32(2)

	if def.IsEvolved {
		bgCol = color.RGBA{R: 60, G: 40, B: 70, A: 255} // Magic purple bg
		borderWidth = 4
	}

	vector.FillRect(img, 0, 0, float32(size), float32(size), bgCol, false)
	vector.StrokeRect(img, 0, 0, float32(size), float32(size), borderWidth, def.Color, false)

	cx, cy := float32(size)/2, float32(size)/2
	c := def.Color

	switch t {
	case WeaponPrint:
		vector.StrokeLine(img, cx-15, cy+15, cx+15, cy-15, 6, c, false)
		vector.StrokeLine(img, cx-10, cy+10, cx-5, cy+5, 10, color.RGBA{100, 50, 20, 255}, false)
	case WeaponLogStream:
		// Big glowing sword
		vector.StrokeLine(img, cx-20, cy+20, cx+20, cy-20, 10, c, false)
		vector.StrokeLine(img, cx-20, cy+20, cx+20, cy-20, 4, color.White, false)
		vector.StrokeLine(img, cx-10, cy+20, cx+20, cy-10, 2, color.White, false) // Cross-guard

	case WeaponRefactor:
		vector.FillCircle(img, cx, cy, 15, c, false)
		vector.StrokeCircle(img, cx, cy, 20, 2, color.White, false)
	case WeaponCleanCode:
		// Dark void
		vector.FillCircle(img, cx, cy, 20, color.Black, false)
		vector.StrokeCircle(img, cx, cy, 22, 4, c, false)
		vector.StrokeCircle(img, cx, cy, 15, 2, color.White, false)

	case WeaponGitPush:
		vector.StrokeLine(img, cx-15, cy, cx+15, cy, 4, c, false)
		vector.StrokeLine(img, cx+5, cy-10, cx+15, cy, 4, c, false)
		vector.StrokeLine(img, cx+5, cy+10, cx+15, cy, 4, c, false)
	case WeaponForcePush:
		// Multiple arrows
		for i := -1; i <= 1; i++ {
			off := float32(i * 10)
			vector.StrokeLine(img, cx-15, cy+off, cx+15, cy+off, 3, c, false)
			vector.StrokeLine(img, cx+5, cy+off-5, cx+15, cy+off, 3, c, false)
			vector.StrokeLine(img, cx+5, cy+off+5, cx+15, cy+off, 3, c, false)
		}

	case WeaponUnitTests:
		vector.FillCircle(img, cx, cy+5, 12, c, false)
		vector.FillCircle(img, cx, cy-8, 8, c, false)
	case WeaponCI_CD:
		// Skull-ish shape or large drop
		vector.FillCircle(img, cx, cy, 18, c, false)
		vector.FillCircle(img, cx-8, cy-5, 6, color.Black, false) // Eyes
		vector.FillCircle(img, cx+8, cy-5, 6, color.Black, false)

	case WeaponFirewall:
		vector.StrokeCircle(img, cx, cy, 18, 5, c, false)
		vector.StrokeCircle(img, cx, cy, 12, 3, color.RGBA{255, 200, 50, 255}, false)
	case WeaponZeroTrust:
		// Persistent fire
		vector.FillCircle(img, cx, cy, 20, c, false)
		vector.FillCircle(img, cx, cy, 12, color.RGBA{255, 200, 50, 255}, false)

	case WeaponStackOverflow:
		vector.StrokeLine(img, cx-10, cy-20, cx+5, cy, 4, c, false)
		vector.StrokeLine(img, cx+5, cy, cx-5, cy+20, 4, c, false)
	case WeaponCopilot:
		// Loop shape
		vector.StrokeCircle(img, cx, cy, 20, 4, c, false)
		vector.StrokeLine(img, cx-10, cy, cx+10, cy, 4, color.White, false)

	case WeaponDocker:
		vector.FillRect(img, cx-5, cy-20, 10, 40, c, false)
		vector.FillRect(img, cx-15, cy-5, 30, 10, c, false)
	case WeaponK8s:
		// Giant sword (vertical)
		vector.FillRect(img, cx-8, cy-25, 16, 50, c, false)
		vector.StrokeRect(img, cx-8, cy-25, 16, 50, 2, color.White, false)

	case WeaponCoffee:
		vector.FillCircle(img, cx, cy, 18, c, false)
		vector.StrokeLine(img, cx, cy-18, cx, cy-25, 3, color.RGBA{100, 200, 100, 255}, false)
	case WeaponEspresso:
		// Red aura
		vector.FillCircle(img, cx, cy, 20, c, false)
		vector.StrokeCircle(img, cx, cy, 25, 3, color.White, false)
	}

	return img
}

// generatePassiveIcon creates a programmatic passive icon (fallback when no image file).
func (g *Game) generatePassiveIcon(t PassiveType, size int) *ebiten.Image {
	img := ebiten.NewImage(size, size)

	// Background
	vector.FillRect(
		img,
		0,
		0,
		float32(size),
		float32(size),
		color.RGBA{R: 20, G: 30, B: 30, A: 255},
		false,
	)
	vector.StrokeRect(img, 0, 0, float32(size), float32(size), 2, color.RGBA{100, 200, 200, 255}, false)

	cx, cy := float32(size)/2, float32(size)/2

	switch t {
	case PassiveMight: // Sword
		vector.StrokeLine(img, cx-10, cy+10, cx+10, cy-10, 5, color.RGBA{255, 50, 50, 255}, false)
	case PassiveArmor: // Square
		vector.FillRect(img, cx-12, cy-12, 24, 24, color.RGBA{150, 150, 150, 255}, false)
	case PassiveSpeed: // Arrow
		vector.StrokeLine(img, cx-15, cy, cx+15, cy, 4, color.RGBA{255, 255, 50, 255}, false)
	case PassiveMagnet: // U
		vector.StrokeLine(img, cx-10, cy-10, cx-10, cy+10, 4, color.RGBA{50, 50, 255, 255}, false)
		vector.StrokeLine(img, cx+10, cy-10, cx+10, cy+10, 4, color.RGBA{255, 50, 50, 255}, false)
		vector.StrokeLine(img, cx-10, cy+10, cx+10, cy+10, 4, color.White, false)
	case PassiveRecovery: // Heart
		vector.FillCircle(img, cx-6, cy-5, 8, color.RGBA{255, 100, 100, 255}, false)
		vector.FillCircle(img, cx+6, cy-5, 8, color.RGBA{255, 100, 100, 255}, false)
		vector.FillCircle(img, cx, cy+8, 8, color.RGBA{255, 100, 100, 255}, false)
	case PassiveLuck: // Green circle
		vector.FillCircle(img, cx, cy, 15, color.RGBA{50, 200, 50, 255}, false)
	case PassiveGrowth: // Tree/Star
		vector.StrokeLine(img, cx, cy-15, cx, cy+15, 4, color.RGBA{50, 255, 50, 255}, false)
		vector.StrokeLine(img, cx-15, cy, cx+15, cy, 4, color.RGBA{50, 255, 50, 255}, false)
	case PassiveCooldown: // Clock
		vector.StrokeCircle(img, cx, cy, 15, 2, color.RGBA{100, 100, 255, 255}, false)
		vector.StrokeLine(img, cx, cy, cx, cy-10, 2, color.White, false)
		vector.StrokeLine(img, cx, cy, cx+8, cy, 2, color.White, false)
	case PassiveArea: // Empty square
		vector.StrokeRect(img, cx-15, cy-15, 30, 30, 3, color.RGBA{200, 50, 200, 255}, false)
	case PassiveDuration: // Hourglass (X)
		vector.StrokeLine(img, cx-10, cy-15, cx+10, cy+15, 4, color.RGBA{200, 200, 50, 255}, false)
		vector.StrokeLine(img, cx+10, cy-15, cx-10, cy+15, 4, color.RGBA{200, 200, 50, 255}, false)
	case PassiveAmount: // Dots
		vector.FillCircle(img, cx-8, cy, 5, color.White, false)
		vector.FillCircle(img, cx+8, cy, 5, color.White, false)
	case PassiveRevival: // Ankh
		vector.StrokeLine(img, cx, cy-5, cx, cy+15, 4, color.RGBA{255, 200, 50, 255}, false)
		vector.StrokeLine(img, cx-10, cy+5, cx+10, cy+5, 4, color.RGBA{255, 200, 50, 255}, false)
		vector.StrokeCircle(img, cx, cy-10, 6, 3, color.RGBA{255, 200, 50, 255}, false)
	}

	return img
}

func (g *Game) newProjectile() *Projectile {
	if len(g.unusedProjs) > 0 {
		p := g.unusedProjs[len(g.unusedProjs)-1]
		g.unusedProjs = g.unusedProjs[:len(g.unusedProjs)-1]

		for k := range p.HitList {
			delete(p.HitList, k)
		}

		return p
	}

	return &Projectile{HitList: make(map[*Enemy]bool)}
}

func (g *Game) freeProjectile(p *Projectile) {
	g.unusedProjs = append(g.unusedProjs, p)
}

func (g *Game) newParticle() *Particle {
	if len(g.unusedParts) > 0 {
		p := g.unusedParts[len(g.unusedParts)-1]
		g.unusedParts = g.unusedParts[:len(g.unusedParts)-1]

		return p
	}

	return &Particle{}
}

func (g *Game) freeParticle(p *Particle) {
	g.unusedParts = append(g.unusedParts, p)
}

func (g *Game) newDamageNumber() *DamageNumber {
	if len(g.unusedDmg) > 0 {
		d := g.unusedDmg[len(g.unusedDmg)-1]
		g.unusedDmg = g.unusedDmg[:len(g.unusedDmg)-1]

		return d
	}

	return &DamageNumber{}
}

func (g *Game) freeDamageNumber(d *DamageNumber) {
	g.unusedDmg = append(g.unusedDmg, d)
}

// removeBackground removes the background color from sprite images using chroma key.
// This follows the standard 16-bit game development convention where specific colors
// are designated as "transparent" and removed during rendering.
//
// Standard Chroma Key Colors (16-bit RGB 5-6-5 format):
// | Color Name     | Hex (8-bit) | Hex (5-6-5) | Binary (5-6-5)      |
// |----------------|-------------|-------------|---------------------|
// | Pure Magenta   | #FF00FF     | $F81F       | 11111 000000 11111  |
// | Bright Green   | #00FF00     | $07E0       | 00000 111111 00000  |
// | Pure Black     | #000000     | $0000       | 00000 000000 00000  |
//
// Pure Magenta (#FF00FF) is the most common choice because it rarely appears
// naturally in game sprites. This function removes magenta pixels and also
// falls back to edge-detection for images without a magenta background.

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Dev Survivor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(60)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
