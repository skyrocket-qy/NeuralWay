package main

import (
	"bytes"
	"embed"
	"image"
	"image/color"
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
)

// Weapon definition.
type WeaponDef struct {
	Name     string
	Damage   int
	Cooldown float64
	Range    float64
	Count    int
	Color    color.RGBA
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
	Name   string
	HP     int
	Speed  float64
	Damage int
	XP     int
	Radius float64
	Color  color.RGBA
	IsBoss bool
}

var MonsterDefs = map[MonsterType]MonsterDef{
	MonsterBat:        {Name: "Bat", HP: 15, Speed: 4.0, Damage: 3, XP: 3, Radius: 10, Color: color.RGBA{R: 80, G: 60, B: 80, A: 255}},
	MonsterSkeleton:   {Name: "Skeleton", HP: 30, Speed: 1.5, Damage: 8, XP: 8, Radius: 14, Color: color.RGBA{R: 200, G: 200, B: 180, A: 255}},
	MonsterZombie:     {Name: "Zombie", HP: 60, Speed: 1.0, Damage: 12, XP: 12, Radius: 16, Color: color.RGBA{R: 100, G: 150, B: 100, A: 255}},
	MonsterGhost:      {Name: "Ghost", HP: 20, Speed: 2.5, Damage: 6, XP: 10, Radius: 12, Color: color.RGBA{R: 200, G: 200, B: 255, A: 180}},
	MonsterDemon:      {Name: "Demon", HP: 100, Speed: 2.0, Damage: 15, XP: 20, Radius: 18, Color: color.RGBA{R: 200, G: 50, B: 50, A: 255}},
	MonsterElemental:  {Name: "Elemental", HP: 50, Speed: 2.5, Damage: 10, XP: 15, Radius: 14, Color: color.RGBA{R: 100, G: 200, B: 255, A: 255}},
	MonsterBossKnight: {Name: "Death Knight", HP: 500, Speed: 1.5, Damage: 25, XP: 200, Radius: 35, Color: color.RGBA{R: 50, G: 50, B: 80, A: 255}, IsBoss: true},
	MonsterBossDragon: {Name: "Dragon", HP: 800, Speed: 2.0, Damage: 30, XP: 400, Radius: 45, Color: color.RGBA{R: 150, G: 50, B: 50, A: 255}, IsBoss: true},
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

// Damage number.
type DamageNumber struct {
	X, Y  float64
	Value int
	Timer float64
	Crit  bool
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
	charImages    []*ebiten.Image

	gameTime     float64
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
		state:        StateCharSelect,
		selectedChar: 0,
		charImages:   make([]*ebiten.Image, len(Characters)),
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
		g.charImages[i] = ebiten.NewImageFromImage(img)
	}

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
		d.Y -= 40 * dt
		d.Timer -= dt
		if d.Timer <= 0 {
			g.damageNumbers = append(g.damageNumbers[:i], g.damageNumbers[i+1:]...)
		}
	}

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
	hpScale := 1.0 + g.gameTime*0.01

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

	switch w.Type {
	case WeaponSword:
		// Arc in front (based on last movement or random)
		for i := 0; i < count; i++ {
			angle := rand.Float64()*math.Pi - math.Pi/2
			g.projectiles = append(g.projectiles, &Projectile{
				X:  g.player.X + math.Cos(angle)*40,
				Y:  g.player.Y + math.Sin(angle)*40,
				VX: math.Cos(angle) * 3, VY: math.Sin(angle) * 3,
				Damage: damage, Lifetime: 0.3, Radius: areaRange / 3, Piercing: 5,
				HitList: make(map[*Enemy]bool), Color: def.Color, WeaponType: w.Type,
			})
		}

	case WeaponOrb:
		for i := 0; i < count; i++ {
			angle := g.gameTime*3 + float64(i)*(2*math.Pi/float64(count))
			g.projectiles = append(g.projectiles, &Projectile{
				X:      g.player.X + math.Cos(angle)*areaRange,
				Y:      g.player.Y + math.Sin(angle)*areaRange,
				Damage: damage, Lifetime: 0.2, Radius: 18, Piercing: 3,
				HitList: make(map[*Enemy]bool), Color: def.Color, WeaponType: w.Type,
			})
		}

	case WeaponArrow:
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
				g.projectiles = append(g.projectiles, &Projectile{
					X: g.player.X, Y: g.player.Y,
					VX: (dx/dist)*speed + spread, VY: (dy/dist)*speed + spread,
					Damage: damage, Lifetime: 2.0, Radius: 6, Piercing: 1,
					HitList: make(map[*Enemy]bool), Color: def.Color, WeaponType: w.Type,
				})
			}
		}

	case WeaponLeech:
		for _, e := range g.enemies {
			if e.Dead {
				continue
			}
			dist := math.Sqrt(math.Pow(g.player.X-e.X, 2) + math.Pow(g.player.Y-e.Y, 2))
			if dist < areaRange {
				e.HP -= damage
				e.HitFlash = 0.1
				// Lifesteal
				if g.player.CharType == CharNecro {
					g.player.HP += damage / 10
					if g.player.HP > g.player.MaxHP {
						g.player.HP = g.player.MaxHP
					}
				}
				g.addDamageNumber(e.X, e.Y, damage, false)
				if e.HP <= 0 {
					g.killEnemy(e)
				}
			}
		}

	case WeaponFireRing:
		for i := 0; i < 8; i++ {
			angle := float64(i) * math.Pi / 4
			g.projectiles = append(g.projectiles, &Projectile{
				X:      g.player.X + math.Cos(angle)*areaRange,
				Y:      g.player.Y + math.Sin(angle)*areaRange,
				Damage: damage, Lifetime: 0.4, Radius: 20, Piercing: 10,
				HitList: make(map[*Enemy]bool), Color: def.Color, WeaponType: w.Type,
			})
		}

	case WeaponLightning:
		// Chain lightning to nearest enemies
		targets := g.findNearestEnemies(count, 300)
		for _, e := range targets {
			e.HP -= damage
			e.HitFlash = 0.15
			g.addDamageNumber(e.X, e.Y, damage, rand.Float64() < g.player.CritChance)
			if e.HP <= 0 {
				g.killEnemy(e)
			}
		}

	case WeaponCross:
		angles := []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2}
		for _, angle := range angles {
			g.projectiles = append(g.projectiles, &Projectile{
				X: g.player.X, Y: g.player.Y,
				VX: math.Cos(angle) * 6, VY: math.Sin(angle) * 6,
				Damage: damage, Lifetime: 1.5, Radius: 12, Piercing: 99,
				HitList: make(map[*Enemy]bool), Color: def.Color, WeaponType: w.Type,
			})
		}

	case WeaponGarlic:
		for _, e := range g.enemies {
			if e.Dead {
				continue
			}
			dist := math.Sqrt(math.Pow(g.player.X-e.X, 2) + math.Pow(g.player.Y-e.Y, 2))
			if dist < areaRange {
				e.HP -= damage
				// Knockback
				dx, dy := e.X-g.player.X, e.Y-g.player.Y
				if dist > 0 {
					e.X += (dx / dist) * 5
					e.Y += (dy / dist) * 5
				}
				if e.HP <= 0 {
					g.killEnemy(e)
				}
			}
		}
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
	result := make([]*Enemy, 0)
	for _, e := range g.enemies {
		if e.Dead {
			continue
		}
		dist := math.Sqrt(math.Pow(g.player.X-e.X, 2) + math.Pow(g.player.Y-e.Y, 2))
		if dist < maxDist {
			result = append(result, e)
			if len(result) >= count {
				break
			}
		}
	}
	return result
}

func (g *Game) killEnemy(e *Enemy) {
	e.Dead = true
	g.killCount++
	g.xpGems = append(g.xpGems, &XPGem{X: e.X, Y: e.Y, Value: e.XP})
}

func (g *Game) addDamageNumber(x, y float64, value int, crit bool) {
	if crit {
		value = int(float64(value) * 1.5)
	}
	g.damageNumbers = append(g.damageNumbers, &DamageNumber{
		X: x, Y: y - 20, Value: value, Timer: 0.6, Crit: crit,
	})
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

	// Add weapon upgrades
	for _, w := range g.player.Weapons {
		if w.Level < 8 {
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
			c := e.Color
			if e.HitFlash > 0 {
				c = color.RGBA{R: 255, G: 255, B: 255, A: 255}
			}
			vector.DrawFilledCircle(screen, float32(sx), float32(sy), float32(e.Radius), c, false)

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
	vector.DrawFilledCircle(screen, float32(px), float32(py), 20, pColor, false)
	vector.StrokeCircle(screen, float32(px), float32(py), 20, 3, color.RGBA{R: 255, G: 255, B: 255, A: 200}, false)

	// Damage numbers
	for _, d := range g.damageNumbers {
		sx, sy := d.X-g.cameraX, d.Y-g.cameraY
		text := formatInt(d.Value)
		if d.Crit {
			text = text + "!"
		}
		ebitenutil.DebugPrintAt(screen, text, int(sx)-10, int(sy))
	}

	// HUD
	g.drawHUD(screen)
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
		def := WeaponDefs[w.Type]
		vector.DrawFilledRect(screen, float32(x), float32(y), 50, 50, def.Color, false)
		vector.StrokeRect(screen, float32(x), float32(y), 50, 50, 2, color.RGBA{R: 255, G: 255, B: 255, A: 150}, false)
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

		// Number
		ebitenutil.DebugPrintAt(screen, "["+formatInt(i+1)+"]", int(boxX)+30, y+15)

		// Name and level
		lvlText := ""
		if opt.CurrentLvl > 0 {
			lvlText = " (Lv " + formatInt(opt.CurrentLvl+1) + ")"
		}
		ebitenutil.DebugPrintAt(screen, opt.Name+lvlText, int(boxX)+70, y+5)
		ebitenutil.DebugPrintAt(screen, opt.Desc, int(boxX)+70, y+25)
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

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Endless Swarm - Survivor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
