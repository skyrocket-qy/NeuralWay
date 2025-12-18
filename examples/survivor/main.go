package main

import (
	"bytes"
	"embed"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
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
	CharKnight CharacterType = iota
	CharMage
	CharArcher
	CharNecro
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
	{Name: "Knight", HP: 150, Speed: 2.5, StartWeapon: WeaponSword, Trait: "Armor", TraitDesc: "+30% Defense", Color: color.RGBA{R: 100, G: 150, B: 200, A: 255}, ImageFile: "assets/hero_knight.png"},
	{Name: "Mage", HP: 80, Speed: 3.0, StartWeapon: WeaponOrb, Trait: "Amplify", TraitDesc: "+50% Area", Color: color.RGBA{R: 150, G: 100, B: 200, A: 255}, ImageFile: "assets/hero_mage.png"},
	{Name: "Archer", HP: 100, Speed: 3.5, StartWeapon: WeaponArrow, Trait: "Swift", TraitDesc: "+25% Proj Speed", Color: color.RGBA{R: 100, G: 200, B: 100, A: 255}, ImageFile: "assets/hero_archer.png"},
	{Name: "Necro", HP: 90, Speed: 2.8, StartWeapon: WeaponLeech, Trait: "Lifesteal", TraitDesc: "10% Lifesteal", Color: color.RGBA{R: 80, G: 80, B: 120, A: 255}, ImageFile: "assets/hero_necro.png"},
}

// WeaponType represents weapon types.
type WeaponType int

const (
	WeaponSword WeaponType = iota
	WeaponOrb
	WeaponArrow
	WeaponLeech
	WeaponFireRing
	WeaponLightning
	WeaponCross
	WeaponGarlic
	// Evolved Weapons
	WeaponHolyBlade
	WeaponVoidOrb
	WeaponThousandEdge
	WeaponSoulEater
	WeaponHellfire
	WeaponThunderLoop
	WeaponHeavenSword
	WeaponCrimsonShroud
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
}

var WeaponDefs = map[WeaponType]WeaponDef{
	WeaponSword:     {Name: "Sword Swing", Damage: 20, Cooldown: 0.8, Range: 80, Count: 1, Color: color.RGBA{R: 200, G: 200, B: 200, A: 255}},
	WeaponOrb:       {Name: "Magic Orb", Damage: 15, Cooldown: 1.0, Range: 100, Count: 2, Color: color.RGBA{R: 100, G: 150, B: 255, A: 255}},
	WeaponArrow:     {Name: "Arrow Storm", Damage: 12, Cooldown: 0.5, Range: 300, Count: 1, Color: color.RGBA{R: 200, G: 150, B: 50, A: 255}},
	WeaponLeech:     {Name: "Soul Leech", Damage: 8, Cooldown: 0.8, Range: 60, Count: 1, Color: color.RGBA{R: 150, G: 50, B: 150, A: 255}},
	WeaponFireRing:  {Name: "Fire Ring", Damage: 25, Cooldown: 1.2, Range: 120, Count: 1, Color: color.RGBA{R: 255, G: 100, B: 50, A: 255}},
	WeaponLightning: {Name: "Lightning", Damage: 30, Cooldown: 1.5, Range: 200, Count: 3, Color: color.RGBA{R: 255, G: 255, B: 100, A: 255}},
	WeaponCross:     {Name: "Holy Cross", Damage: 18, Cooldown: 1.0, Range: 150, Count: 1, Color: color.RGBA{R: 255, G: 255, B: 200, A: 255}},
	WeaponGarlic:    {Name: "Garlic Aura", Damage: 5, Cooldown: 0.3, Range: 50, Count: 1, Color: color.RGBA{R: 200, G: 255, B: 200, A: 255}},

	// Evolved
	WeaponHolyBlade:     {Name: "Holy Blade", Damage: 40, Cooldown: 0.6, Range: 120, Count: 1, Color: color.RGBA{R: 255, G: 215, B: 0, A: 255}, IsEvolved: true},
	WeaponVoidOrb:       {Name: "Void Orbs", Damage: 25, Cooldown: 0.8, Range: 140, Count: 4, Color: color.RGBA{R: 100, G: 0, B: 150, A: 255}, IsEvolved: true},
	WeaponThousandEdge:  {Name: "Thousand Edge", Damage: 15, Cooldown: 0.1, Range: 400, Count: 1, Color: color.RGBA{R: 255, G: 100, B: 0, A: 255}, IsEvolved: true}, // Machine gun
	WeaponSoulEater:     {Name: "Soul Eater", Damage: 12, Cooldown: 0.5, Range: 100, Count: 1, Color: color.RGBA{R: 255, G: 0, B: 50, A: 255}, IsEvolved: true},
	WeaponHellfire:      {Name: "Hellfire", Damage: 40, Cooldown: 1.0, Range: 160, Count: 1, Color: color.RGBA{R: 255, G: 50, B: 0, A: 255}, IsEvolved: true},
	WeaponThunderLoop:   {Name: "Thunder Loop", Damage: 40, Cooldown: 1.0, Range: 250, Count: 5, Color: color.RGBA{R: 200, G: 255, B: 255, A: 255}, IsEvolved: true},
	WeaponHeavenSword:   {Name: "Heaven Sword", Damage: 50, Cooldown: 0.8, Range: 200, Count: 1, Color: color.RGBA{R: 200, G: 255, B: 255, A: 255}, IsEvolved: true},
	WeaponCrimsonShroud: {Name: "Crimson Shroud", Damage: 10, Cooldown: 0.2, Range: 80, Count: 1, Color: color.RGBA{R: 255, G: 100, B: 100, A: 255}, IsEvolved: true},
}

type EvolutionRecipe struct {
	BaseWeapon WeaponType
	Passive    PassiveType
	Result     WeaponType
}

var Evolutions = []EvolutionRecipe{
	{WeaponSword, PassiveArmor, WeaponHolyBlade},
	{WeaponOrb, PassiveMagnet, WeaponVoidOrb},
	{WeaponArrow, PassiveSpeed, WeaponThousandEdge},
	{WeaponLeech, PassiveRecovery, WeaponSoulEater},
	{WeaponFireRing, PassiveArea, WeaponHellfire},
	{WeaponLightning, PassiveCooldown, WeaponThunderLoop},
	{WeaponCross, PassiveLuck, WeaponHeavenSword},
	{WeaponGarlic, PassiveDuration, WeaponCrimsonShroud},
}

// MonsterType represents enemy types.
type MonsterType int

const (
	MonsterBat MonsterType = iota
	MonsterSkeleton
	MonsterZombie
	MonsterGhost
	MonsterDemon
	MonsterElemental
	MonsterBossKnight
	MonsterBossDragon
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
	MonsterBat:       {Name: "Bat", HP: 10, Speed: 2.2, Damage: 3, XP: 1, Radius: 10, Color: color.RGBA{100, 100, 100, 255}, ImageFile: "assets/monster_bat.png"},
	MonsterSkeleton:  {Name: "Skeleton", HP: 30, Speed: 1.5, Damage: 8, XP: 3, Radius: 12, Color: color.RGBA{200, 200, 200, 255}, ImageFile: "assets/monster_skeleton.png"},
	MonsterZombie:    {Name: "Zombie", HP: 50, Speed: 0.8, Damage: 12, XP: 5, Radius: 14, Color: color.RGBA{50, 150, 50, 255}, ImageFile: "assets/monster_zombie.png"},
	MonsterGhost:     {Name: "Ghost", HP: 20, Speed: 2.0, Damage: 6, XP: 2, Radius: 10, Color: color.RGBA{200, 200, 255, 150}, ImageFile: "assets/monster_ghost.png"},
	MonsterDemon:     {Name: "Demon", HP: 80, Speed: 2.0, Damage: 15, XP: 10, Radius: 20, Color: color.RGBA{200, 50, 50, 255}, ImageFile: "assets/monster_ghost.png"},    // Placeholder
	MonsterElemental: {Name: "Elemental", HP: 40, Speed: 3.0, Damage: 10, XP: 8, Radius: 15, Color: color.RGBA{50, 50, 200, 255}, ImageFile: "assets/monster_ghost.png"}, // Placeholder

	// Bosses
	MonsterBossKnight: {Name: "Death Knight", HP: 1500, Speed: 1.8, Damage: 20, XP: 500, Radius: 30, Color: color.RGBA{50, 0, 50, 255}, IsBoss: true, ImageFile: "assets/monster_skeleton.png"}, // Placeholder
	MonsterBossDragon: {Name: "Dragon", HP: 5000, Speed: 2.5, Damage: 40, XP: 2000, Radius: 50, Color: color.RGBA{200, 100, 0, 255}, IsBoss: true, ImageFile: "assets/monster_bat.png"},         // Placeholder
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
	Name   string
	Desc   string
	MaxLvl int
}

var PassiveDefs = map[PassiveType]PassiveDef{
	PassiveMight:    {Name: "Might", Desc: "+10% damage", MaxLvl: 5},
	PassiveArmor:    {Name: "Armor", Desc: "-5 damage taken", MaxLvl: 5},
	PassiveSpeed:    {Name: "Speed", Desc: "+10% move speed", MaxLvl: 5},
	PassiveMagnet:   {Name: "Magnet", Desc: "+20% pickup range", MaxLvl: 5},
	PassiveRecovery: {Name: "Recovery", Desc: "+0.3 HP/s", MaxLvl: 5},
	PassiveLuck:     {Name: "Luck", Desc: "+10% crit", MaxLvl: 5},
	PassiveGrowth:   {Name: "Growth", Desc: "+10% XP", MaxLvl: 5},
	PassiveCooldown: {Name: "Cooldown", Desc: "-5% cooldown", MaxLvl: 5},
	PassiveArea:     {Name: "Area", Desc: "+10% area", MaxLvl: 5},
	PassiveDuration: {Name: "Duration", Desc: "+10% duration", MaxLvl: 5},
	PassiveAmount:   {Name: "Amount", Desc: "+1 projectile", MaxLvl: 3},
	PassiveRevival:  {Name: "Revival", Desc: "Revive once", MaxLvl: 1},
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
}

// GameState enum.
type GameState int

const (
	StateCharSelect GameState = iota
	StatePlaying
	StateLevelUp
	StatePaused
	StateGameOver
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

	upgradeOptions   []UpgradeOption
	cameraX, cameraY float64
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
	rand.Seed(time.Now().UnixNano())
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
		img = removeBackground(img)
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
		img = removeBackground(img)
		g.monsterImages[t] = ebiten.NewImageFromImage(img)
	}

	g.generateIcons()
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
		Passives:     make(map[PassiveType]int),
		DamageMult:   1.0,
		AreaMult:     1.0,
		CooldownMult: 1.0,
		MagnetRange:  80,
		XPMult:       1.0,
	}

	// Apply character traits
	switch charType {
	case CharMage:
		g.player.AreaMult = 1.5
	case CharArcher:
		// Handled in projectile speed
	case CharNecro:
		// Lifesteal handled in damage
	}

	g.enemies = make([]*Enemy, 0)
	g.projectiles = make([]*Projectile, 0)
	g.xpGems = make([]*XPGem, 0)
	g.damageNumbers = make([]*DamageNumber, 0)
	g.gameTime = 0
	g.spawnTimer = 0
	g.bossTimer = 0
	g.killCount = 0
	g.state = StatePlaying
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
	}
	return nil
}

func (g *Game) updateCharSelect() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.selectedChar--
		if g.selectedChar < 0 {
			g.selectedChar = len(Characters) - 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.selectedChar++
		if g.selectedChar >= len(Characters) {
			g.selectedChar = 0
		}
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

	dt := 1.0 / 60.0
	g.gameTime += dt

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
	spawnRate := 1.5 - g.gameTime*0.005
	if spawnRate < 0.2 {
		spawnRate = 0.2
	}
	if g.spawnTimer >= spawnRate {
		g.spawnEnemy()
		g.spawnTimer = 0
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
			monsterType = MonsterBat
		} else {
			monsterType = MonsterSkeleton
		}
	} else if g.gameTime < 180 {
		if r < 0.3 {
			monsterType = MonsterBat
		} else if r < 0.6 {
			monsterType = MonsterSkeleton
		} else if r < 0.8 {
			monsterType = MonsterZombie
		} else {
			monsterType = MonsterGhost
		}
	} else {
		if r < 0.2 {
			monsterType = MonsterSkeleton
		} else if r < 0.4 {
			monsterType = MonsterZombie
		} else if r < 0.6 {
			monsterType = MonsterGhost
		} else if r < 0.8 {
			monsterType = MonsterDemon
		} else {
			monsterType = MonsterElemental
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

	bossType := MonsterBossKnight
	if g.gameTime > 360 {
		bossType = MonsterBossDragon
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
	case WeaponSword, WeaponHolyBlade:
		// Aim at nearest enemy
		target := g.findNearestEnemy(120) // Melee range
		baseAngle := 0.0
		if target != nil {
			baseAngle = math.Atan2(target.Y-g.player.Y, target.X-g.player.X)
		} else {
			baseAngle = (rand.Float64() - 0.5) * math.Pi / 2
		}

		for i := 0; i < count; i++ {
			// Spread around base angle
			angle := baseAngle
			if count > 1 {
				spread := math.Pi / 3 // 60 degrees spread
				angle += spread * (float64(i)/float64(count-1) - 0.5)
			}

			// Larger area for evolved
			radius := areaRange / 3
			if w.Type == WeaponHolyBlade {
				radius *= 1.5
			}

			spawnProj(
				g.player.X+math.Cos(angle)*40, g.player.Y+math.Sin(angle)*40,
				math.Cos(angle)*3, math.Sin(angle)*3,
				0.3, radius, 5,
			)
		}

	case WeaponOrb, WeaponVoidOrb:
		for i := 0; i < count; i++ {
			angle := g.gameTime*3 + float64(i)*(2*math.Pi/float64(count))
			spawnProj(
				g.player.X+math.Cos(angle)*areaRange, g.player.Y+math.Sin(angle)*areaRange,
				0, 0,
				0.2, 18, 3,
			)
		}

	case WeaponArrow, WeaponThousandEdge:
		nearest := g.findNearestEnemy(500)
		if nearest != nil {
			dx, dy := nearest.X-g.player.X, nearest.Y-g.player.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			speed := 10.0
			if g.player.CharType == CharArcher {
				speed = 12.5
			}
			for i := 0; i < count; i++ {
				spread := float64(i-count/2) * 0.15
				spawnProj(
					g.player.X, g.player.Y,
					(dx/dist)*speed+spread, (dy/dist)*speed+spread,
					2.0, 6, 1,
				)
			}
		}

	case WeaponLeech, WeaponSoulEater:
		// Area damage around player
		spawnProj(g.player.X, g.player.Y, 0, 0, 0.2, areaRange, 999)

	case WeaponFireRing, WeaponHellfire:
		// Orbiting fireball
		angle := g.gameTime * 2
		spawnProj(
			g.player.X+math.Cos(angle)*areaRange, g.player.Y+math.Sin(angle)*areaRange,
			0, 0,
			0.5, 15, 999,
		)

	case WeaponLightning, WeaponThunderLoop:
		// Random enemies
		targets := g.findNearestEnemies(count, 300)
		for _, e := range targets {
			spawnProj(e.X, e.Y-50, 0, 20, 0.2, 30, 1)
		}

	case WeaponCross, WeaponHeavenSword:
		// Throws forward then returns
		spawnProj(g.player.X, g.player.Y, 5, -5, 2.0, 15, 999)

	case WeaponGarlic, WeaponCrimsonShroud:
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
		dist := math.Sqrt(math.Pow(g.player.X-e.X, 2) + math.Pow(g.player.Y-e.Y, 2))
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
		dist := math.Sqrt(math.Pow(g.player.X-e.X, 2) + math.Pow(g.player.Y-e.Y, 2))
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
	for i := 0; i < count; i++ {
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

		for _, e := range g.enemies {
			if e.Dead || p.HitList[e] {
				continue
			}
			dist := math.Sqrt(math.Pow(p.X-e.X, 2) + math.Pow(p.Y-e.Y, 2))
			if dist < p.Radius+e.Radius {
				crit := rand.Float64() < g.player.CritChance
				damage := p.Damage
				if crit {
					damage = int(float64(damage) * 1.5)
				}
				e.HP -= damage
				e.HitFlash = 0.1
				p.HitList[e] = true
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
	for i := len(g.enemies) - 1; i >= 0; i-- {
		e := g.enemies[i]
		if e.Dead {
			g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
			continue
		}
		e.HitFlash -= dt

		dx, dy := g.player.X-e.X, g.player.Y-e.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist > 0 {
			e.X += (dx / dist) * e.Speed
			e.Y += (dy / dist) * e.Speed
		}

		if dist < 20+e.Radius {
			damage := e.Damage - g.player.Armor
			if damage < 1 {
				damage = 1
			}
			g.player.HP -= damage
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
				g.showLevelUp()
			}
		}
	}
}

func (g *Game) showLevelUp() {
	g.state = StateLevelUp
	g.upgradeOptions = g.generateUpgrades()
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
	allWeapons := []WeaponType{WeaponSword, WeaponOrb, WeaponArrow, WeaponLeech, WeaponFireRing, WeaponLightning, WeaponCross, WeaponGarlic}
	for _, wt := range allWeapons {
		has := false
		for _, w := range g.player.Weapons {
			if w.Type == wt {
				has = true
				break
			}
		}
		// Also check if we have the evolved version (don't offer base if we have evolved? Actually in VS you can have both if you find base again? No, usually not).
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
		vector.DrawFilledRect(screen, float32(x), float32(y), 150, 280, boxColor, false)

		if i == g.selectedChar {
			vector.StrokeRect(screen, float32(x), float32(y), 150, 280, 3, color.RGBA{R: 255, G: 215, B: 0, A: 255}, false)
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
			vector.DrawFilledCircle(screen, float32(x+75), float32(y+60), 40, char.Color, false)
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
	ebitenutil.DebugPrintAt(screen, "LEFT/RIGHT to select | SPACE to start", screenWidth/2-130, screenHeight-50)
}

func (g *Game) drawGame(screen *ebiten.Image) {
	// Background
	screen.Fill(color.RGBA{R: 25, G: 30, B: 40, A: 255})

	// Grid
	gridSize := 60.0
	offsetX := math.Mod(g.cameraX, gridSize)
	offsetY := math.Mod(g.cameraY, gridSize)
	for x := -gridSize; x < float64(screenWidth)+gridSize; x += gridSize {
		vector.DrawFilledRect(screen, float32(x-offsetX), 0, 1, screenHeight, color.RGBA{R: 35, G: 40, B: 50, A: 255}, false)
	}
	for y := -gridSize; y < float64(screenHeight)+gridSize; y += gridSize {
		vector.DrawFilledRect(screen, 0, float32(y-offsetY), screenWidth, 1, color.RGBA{R: 35, G: 40, B: 50, A: 255}, false)
	}

	// XP Gems
	for _, gem := range g.xpGems {
		sx, sy := gem.X-g.cameraX, gem.Y-g.cameraY
		if sx >= -20 && sx <= screenWidth+20 && sy >= -20 && sy <= screenHeight+20 {
			size := float32(6)
			if gem.Value >= 20 {
				size = 10
			}
			vector.DrawFilledRect(screen, float32(sx)-size/2, float32(sy)-size/2, size, size, color.RGBA{R: 100, G: 200, B: 255, A: 255}, false)
		}
	}

	// Enemies
	for _, e := range g.enemies {
		sx, sy := e.X-g.cameraX, e.Y-g.cameraY
		if sx >= -50 && sx <= screenWidth+50 && sy >= -50 && sy <= screenHeight+50 {
			// Render monster image or fallback circle
			if img := g.monsterImages[e.Type]; img != nil {
				bounds := img.Bounds()
				// Scale to fit radius * 2
				scale := (e.Radius * 2.5) / float64(bounds.Dx())

				op := &ebiten.DrawImageOptions{}
				op.GeoM.Scale(scale, scale)
				op.GeoM.Translate(sx-float64(bounds.Dx())*scale/2, sy-float64(bounds.Dy())*scale/2)

				// Tinting
				if e.HitFlash > 0 {
					op.ColorM.Scale(10, 10, 10, 1)
				} else {
					// Apply monster color tint
					r := float64(e.Color.R) / 255.0
					g := float64(e.Color.G) / 255.0
					b := float64(e.Color.B) / 255.0
					op.ColorM.Scale(r, g, b, 1)
				}

				screen.DrawImage(img, op)
			} else {
				c := e.Color
				if e.HitFlash > 0 {
					c = color.RGBA{R: 255, G: 255, B: 255, A: 255}
				}
				vector.DrawFilledCircle(screen, float32(sx), float32(sy), float32(e.Radius), c, false)
			}

			// Boss indicator
			if e.IsBoss {
				vector.StrokeCircle(screen, float32(sx), float32(sy), float32(e.Radius)+5, 3, color.RGBA{R: 255, G: 50, B: 50, A: 255}, false)
			}

			// HP bar
			if e.HP < e.MaxHP {
				barW := e.Radius * 2
				hpRatio := float32(e.HP) / float32(e.MaxHP)
				vector.DrawFilledRect(screen, float32(sx)-float32(barW)/2, float32(sy)-float32(e.Radius)-8, float32(barW), 4, color.RGBA{R: 50, G: 50, B: 50, A: 255}, false)
				vector.DrawFilledRect(screen, float32(sx)-float32(barW)/2, float32(sy)-float32(e.Radius)-8, float32(barW)*hpRatio, 4, color.RGBA{R: 255, G: 50, B: 50, A: 255}, false)
			}
		}
	}

	// Projectiles
	for _, p := range g.projectiles {
		sx, sy := p.X-g.cameraX, p.Y-g.cameraY
		vector.DrawFilledCircle(screen, float32(sx), float32(sy), float32(p.Radius), p.Color, false)
	}

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
		vector.DrawFilledCircle(screen, float32(px), float32(py), 20, pColor, false)
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

func (g *Game) drawParticles(screen *ebiten.Image) {
	for _, p := range g.particles {
		sx, sy := p.X-g.cameraX, p.Y-g.cameraY
		if sx >= -10 && sx <= screenWidth+10 && sy >= -10 && sy <= screenHeight+10 {
			// Fade out alpha
			alpha := float32(p.Lifetime / p.MaxLife)
			c := p.Color
			c.A = uint8(float32(c.A) * alpha)

			vector.DrawFilledRect(screen, float32(sx)-float32(p.Size)/2, float32(sy)-float32(p.Size)/2, float32(p.Size), float32(p.Size), c, false)
		}
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	// Top bar
	vector.DrawFilledRect(screen, 0, 0, screenWidth, 60, color.RGBA{R: 0, G: 0, B: 0, A: 200}, false)

	// Character portrait
	vector.DrawFilledCircle(screen, 30, 30, 22, Characters[g.player.CharType].Color, false)

	// HP bar
	vector.DrawFilledRect(screen, 60, 8, 200, 18, color.RGBA{R: 40, G: 40, B: 40, A: 255}, false)
	hpRatio := float32(g.player.HP) / float32(g.player.MaxHP)
	vector.DrawFilledRect(screen, 60, 8, 200*hpRatio, 18, color.RGBA{R: 200, G: 50, B: 50, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, formatInt(g.player.HP)+"/"+formatInt(g.player.MaxHP), 130, 10)

	// XP bar
	xpNeeded := g.player.Level * 25
	xpRatio := float32(g.player.XP) / float32(xpNeeded)
	vector.DrawFilledRect(screen, 60, 32, 200, 12, color.RGBA{R: 40, G: 40, B: 40, A: 255}, false)
	vector.DrawFilledRect(screen, 60, 32, 200*xpRatio, 12, color.RGBA{R: 100, G: 200, B: 255, A: 255}, false)

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
			vector.DrawFilledRect(screen, float32(x), float32(y), 50, 50, def.Color, false)
			vector.StrokeRect(screen, float32(x), float32(y), 50, 50, 2, color.RGBA{R: 255, G: 255, B: 255, A: 150}, false)
		}
		ebitenutil.DebugPrintAt(screen, formatInt(w.Level), x+38, y+35)
	}

	// Controls hint
	ebitenutil.DebugPrintAt(screen, "WASD = Move | ESC = Pause", screenWidth-200, screenHeight-20)
}

func (g *Game) drawLevelUp(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{R: 0, G: 0, B: 0, A: 180}, false)

	boxW, boxH := float32(500), float32(300)
	boxX, boxY := float32(screenWidth-500)/2, float32(screenHeight-300)/2

	vector.DrawFilledRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 30, G: 35, B: 50, A: 255}, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 3, color.RGBA{R: 255, G: 215, B: 0, A: 255}, false)

	ebitenutil.DebugPrintAt(screen, "LEVEL UP! Choose an upgrade:", int(boxX)+140, int(boxY)+15)

	for i, opt := range g.upgradeOptions {
		y := int(boxY) + 55 + i*60

		// Option box
		optColor := color.RGBA{R: 50, G: 55, B: 70, A: 255}
		vector.DrawFilledRect(screen, boxX+20, float32(y)-5, boxW-40, 55, optColor, false)

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
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{R: 0, G: 0, B: 0, A: 180}, false)

	boxW, boxH := float32(300), float32(180)
	boxX, boxY := float32(screenWidth-300)/2, float32(screenHeight-180)/2

	vector.DrawFilledRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 40, G: 45, B: 60, A: 255}, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 2, color.RGBA{R: 200, G: 200, B: 200, A: 255}, false)

	ebitenutil.DebugPrintAt(screen, "PAUSED", int(boxX)+115, int(boxY)+30)
	ebitenutil.DebugPrintAt(screen, "SPACE / ESC - Resume", int(boxX)+70, int(boxY)+80)
	ebitenutil.DebugPrintAt(screen, "Q - Quit to Menu", int(boxX)+85, int(boxY)+110)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{R: 0, G: 0, B: 0, A: 200}, false)

	boxW, boxH := float32(350), float32(250)
	boxX, boxY := float32(screenWidth-350)/2, float32(screenHeight-250)/2

	vector.DrawFilledRect(screen, boxX, boxY, boxW, boxH, color.RGBA{R: 50, G: 30, B: 30, A: 255}, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 3, color.RGBA{R: 200, G: 50, B: 50, A: 255}, false)

	ebitenutil.DebugPrintAt(screen, "GAME OVER", int(boxX)+120, int(boxY)+25)

	ebitenutil.DebugPrintAt(screen, "Survived: "+formatTime(g.gameTime), int(boxX)+100, int(boxY)+70)
	ebitenutil.DebugPrintAt(screen, "Level: "+formatInt(g.player.Level), int(boxX)+120, int(boxY)+95)
	ebitenutil.DebugPrintAt(screen, "Kills: "+formatInt(g.killCount), int(boxX)+120, int(boxY)+120)
	ebitenutil.DebugPrintAt(screen, "Weapons: "+formatInt(len(g.player.Weapons)), int(boxX)+110, int(boxY)+145)

	ebitenutil.DebugPrintAt(screen, "SPACE - Retry", int(boxX)+110, int(boxY)+185)
	ebitenutil.DebugPrintAt(screen, "Q - Character Select", int(boxX)+85, int(boxY)+210)
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

func formatFloat(f float64) string {
	return formatInt(int(f * 10)) // Simple approximation
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

	size := 64

	// Weapons
	for t, def := range WeaponDefs {
		img := ebiten.NewImage(size, size)

		// Background & Border
		bgCol := color.RGBA{R: 30, G: 30, B: 40, A: 255}
		borderWidth := float32(2)
		if def.IsEvolved {
			bgCol = color.RGBA{R: 60, G: 40, B: 70, A: 255} // Magic purple bg
			borderWidth = 4
		}

		vector.DrawFilledRect(img, 0, 0, float32(size), float32(size), bgCol, false)
		vector.StrokeRect(img, 0, 0, float32(size), float32(size), borderWidth, def.Color, false)

		cx, cy := float32(size)/2, float32(size)/2
		c := def.Color

		switch t {
		case WeaponSword:
			vector.StrokeLine(img, cx-15, cy+15, cx+15, cy-15, 6, c, false)
			vector.StrokeLine(img, cx-10, cy+10, cx-5, cy+5, 10, color.RGBA{100, 50, 20, 255}, false)
		case WeaponHolyBlade:
			// Big glowing sword
			vector.StrokeLine(img, cx-20, cy+20, cx+20, cy-20, 10, c, false)
			vector.StrokeLine(img, cx-20, cy+20, cx+20, cy-20, 4, color.White, false)
			vector.StrokeLine(img, cx-10, cy+20, cx+20, cy-10, 2, color.White, false) // Cross-guard

		case WeaponOrb:
			vector.DrawFilledCircle(img, cx, cy, 15, c, false)
			vector.StrokeCircle(img, cx, cy, 20, 2, color.White, false)
		case WeaponVoidOrb:
			// Dark void
			vector.DrawFilledCircle(img, cx, cy, 20, color.Black, false)
			vector.StrokeCircle(img, cx, cy, 22, 4, c, false)
			vector.StrokeCircle(img, cx, cy, 15, 2, color.White, false)

		case WeaponArrow:
			vector.StrokeLine(img, cx-15, cy, cx+15, cy, 4, c, false)
			vector.StrokeLine(img, cx+5, cy-10, cx+15, cy, 4, c, false)
			vector.StrokeLine(img, cx+5, cy+10, cx+15, cy, 4, c, false)
		case WeaponThousandEdge:
			// Multiple arrows
			for i := -1; i <= 1; i++ {
				off := float32(i * 10)
				vector.StrokeLine(img, cx-15, cy+off, cx+15, cy+off, 3, c, false)
				vector.StrokeLine(img, cx+5, cy+off-5, cx+15, cy+off, 3, c, false)
				vector.StrokeLine(img, cx+5, cy+off+5, cx+15, cy+off, 3, c, false)
			}

		case WeaponLeech:
			vector.DrawFilledCircle(img, cx, cy+5, 12, c, false)
			vector.DrawFilledCircle(img, cx, cy-8, 8, c, false)
		case WeaponSoulEater:
			// Skull-ish shape or large drop
			vector.DrawFilledCircle(img, cx, cy, 18, c, false)
			vector.DrawFilledCircle(img, cx-8, cy-5, 6, color.Black, false) // Eyes
			vector.DrawFilledCircle(img, cx+8, cy-5, 6, color.Black, false)

		case WeaponFireRing:
			vector.StrokeCircle(img, cx, cy, 18, 5, c, false)
			vector.StrokeCircle(img, cx, cy, 12, 3, color.RGBA{255, 200, 50, 255}, false)
		case WeaponHellfire:
			// Persistent fire
			vector.DrawFilledCircle(img, cx, cy, 20, c, false)
			vector.DrawFilledCircle(img, cx, cy, 12, color.RGBA{255, 200, 50, 255}, false)

		case WeaponLightning:
			vector.StrokeLine(img, cx-10, cy-20, cx+5, cy, 4, c, false)
			vector.StrokeLine(img, cx+5, cy, cx-5, cy+20, 4, c, false)
		case WeaponThunderLoop:
			// Loop shape
			vector.StrokeCircle(img, cx, cy, 20, 4, c, false)
			vector.StrokeLine(img, cx-10, cy, cx+10, cy, 4, color.White, false)

		case WeaponCross:
			vector.DrawFilledRect(img, cx-5, cy-20, 10, 40, c, false)
			vector.DrawFilledRect(img, cx-15, cy-5, 30, 10, c, false)
		case WeaponHeavenSword:
			// Giant sword (vertical)
			vector.DrawFilledRect(img, cx-8, cy-25, 16, 50, c, false)
			vector.StrokeRect(img, cx-8, cy-25, 16, 50, 2, color.White, false)

		case WeaponGarlic:
			vector.DrawFilledCircle(img, cx, cy, 18, c, false)
			vector.StrokeLine(img, cx, cy-18, cx, cy-25, 3, color.RGBA{100, 200, 100, 255}, false)
		case WeaponCrimsonShroud:
			// Red aura
			vector.DrawFilledCircle(img, cx, cy, 20, c, false)
			vector.StrokeCircle(img, cx, cy, 25, 3, color.White, false)
		}

		g.weaponImages[t] = img
	}

	// Passives (generic icons)
	for t := range PassiveDefs {
		img := ebiten.NewImage(size, size)

		// Background
		vector.DrawFilledRect(img, 0, 0, float32(size), float32(size), color.RGBA{R: 20, G: 30, B: 30, A: 255}, false)
		vector.StrokeRect(img, 0, 0, float32(size), float32(size), 2, color.RGBA{100, 200, 200, 255}, false)

		cx, cy := float32(size)/2, float32(size)/2

		switch t {
		case PassiveMight: // Sword
			vector.StrokeLine(img, cx-10, cy+10, cx+10, cy-10, 5, color.RGBA{255, 50, 50, 255}, false)
		case PassiveArmor: // Square
			vector.DrawFilledRect(img, cx-12, cy-12, 24, 24, color.RGBA{150, 150, 150, 255}, false)
		case PassiveSpeed: // Arrow
			vector.StrokeLine(img, cx-15, cy, cx+15, cy, 4, color.RGBA{255, 255, 50, 255}, false)
		case PassiveMagnet: // U
			vector.StrokeLine(img, cx-10, cy-10, cx-10, cy+10, 4, color.RGBA{50, 50, 255, 255}, false)
			vector.StrokeLine(img, cx+10, cy-10, cx+10, cy+10, 4, color.RGBA{255, 50, 50, 255}, false)
			vector.StrokeLine(img, cx-10, cy+10, cx+10, cy+10, 4, color.White, false)
		case PassiveRecovery: // Heart
			vector.DrawFilledCircle(img, cx-6, cy-5, 8, color.RGBA{255, 100, 100, 255}, false)
			vector.DrawFilledCircle(img, cx+6, cy-5, 8, color.RGBA{255, 100, 100, 255}, false)
			vector.DrawFilledCircle(img, cx, cy+8, 8, color.RGBA{255, 100, 100, 255}, false)
		case PassiveLuck: // Green circle
			vector.DrawFilledCircle(img, cx, cy, 15, color.RGBA{50, 200, 50, 255}, false)
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
			vector.DrawFilledCircle(img, cx-8, cy, 5, color.White, false)
			vector.DrawFilledCircle(img, cx+8, cy, 5, color.White, false)
		case PassiveRevival: // Ankh
			vector.StrokeLine(img, cx, cy-5, cx, cy+15, 4, color.RGBA{255, 200, 50, 255}, false)
			vector.StrokeLine(img, cx-10, cy+5, cx+10, cy+5, 4, color.RGBA{255, 200, 50, 255}, false)
			vector.StrokeCircle(img, cx, cy-10, 6, 3, color.RGBA{255, 200, 50, 255}, false)
		}

		g.passiveImages[t] = img
	}
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
func removeBackground(src image.Image) image.Image {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)

	// Standard chroma key colors (in 16-bit values from RGBA())
	// RGBA() returns 16-bit values (0-65535), so:
	// - Pure Magenta #FF00FF = R:65535, G:0, B:65535
	// - Bright Green #00FF00 = R:0, G:65535, B:0
	const (
		magentaR     uint32 = 65535
		magentaG     uint32 = 0
		magentaB     uint32 = 65535
		keyThreshold        = 8000 // Tolerance for compression artifacts
	)

	// First pass: try to detect if image uses standard magenta key
	hasMagentaKey := false
	c := src.At(bounds.Min.X, bounds.Min.Y)
	r, g, b, _ := c.RGBA()
	if abs(int(r)-int(magentaR)) < 5000 && g < 5000 && abs(int(b)-int(magentaB)) < 5000 {
		hasMagentaKey = true
	}

	if hasMagentaKey {
		// Use standard magenta chroma key removal
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				c := dst.At(x, y)
				r, g, b, _ := c.RGBA()

				// Check if pixel is close to magenta
				if abs(int(r)-int(magentaR)) < int(keyThreshold) &&
					g < keyThreshold &&
					abs(int(b)-int(magentaB)) < int(keyThreshold) {
					dst.Set(x, y, color.Transparent)
				}
			}
		}
	} else {
		// Fallback: sample background from edges and remove similar colors
		var avgR, avgG, avgB uint64
		var count uint64

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := src.At(x, bounds.Min.Y)
			r, g, b, _ := c.RGBA()
			avgR += uint64(r)
			avgG += uint64(g)
			avgB += uint64(b)
			count++

			c = src.At(x, bounds.Max.Y-1)
			r, g, b, _ = c.RGBA()
			avgR += uint64(r)
			avgG += uint64(g)
			avgB += uint64(b)
			count++
		}
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			c := src.At(bounds.Min.X, y)
			r, g, b, _ := c.RGBA()
			avgR += uint64(r)
			avgG += uint64(g)
			avgB += uint64(b)
			count++

			c = src.At(bounds.Max.X-1, y)
			r, g, b, _ = c.RGBA()
			avgR += uint64(r)
			avgG += uint64(g)
			avgB += uint64(b)
			count++
		}
		avgR /= count
		avgG /= count
		avgB /= count

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				c := dst.At(x, y)
				r, g, b, _ := c.RGBA()

				d := abs(int(r)-int(avgR)) + abs(int(g)-int(avgG)) + abs(int(b)-int(avgB))
				if d < 35000 {
					dst.Set(x, y, color.Transparent)
				}
			}
		}
	}
	return dst
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Endless Swarm - Survivor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
