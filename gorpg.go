package main

import (
	"bytes"
	_ "embed"
	"gorpg/tilemaps"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"

	"time"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	screenWidth   = 1920 / 3
	screenHeight  = 1080 / 3
	imgSize       = 48
	SPEED         = time.Second / 4
	houseTileSize = 64
	SampleRate    = 44100
	wheat         = "wheat"
	tomato        = "tomato"
	chicken       = "chicken"
	egg           = "egg"
	chest         = 0
	fire          = 1
)

//go:embed assets/sound/LostVillage.ogg
var audioBG []byte

//go:embed assets/sound/Village.ogg
var audioVillage []byte

//go:embed assets/sound/Coin.ogg
var audioCoin []byte

//go:embed assets/sound/Fx.ogg
var audioFx []byte

//go:embed assets/sound/Chest.ogg
var audioChest []byte

//go:embed assets/sound/Secret.ogg
var audioSecret []byte

var (
	skyBlue         = color.RGBA{120, 180, 255, 255}
	red             = color.RGBA{255, 0, 0, 255}
	red_rect        = color.RGBA{255, 0, 0, 40}
	blue            = color.RGBA{0, 10, 60, 255}
	blue_transp     = color.RGBA{0, 10, 60, 100}
	blue_rect       = color.RGBA{0, 20, 120, 40}
	yellow          = color.RGBA{220, 200, 0, 255}
	green           = color.RGBA{0, 220, 0, 255}
	dark_green      = color.RGBA{0, 140, 20, 255}
	purple          = color.RGBA{200, 0, 200, 255}
	orange          = color.RGBA{180, 160, 0, 255}
	white           = color.RGBA{255, 255, 255, 255}
	black           = color.RGBA{0, 0, 0, 255}
	gameSpeed       = SPEED
	PlayerSpeed     = 3.0
	diagonalSpeed   = 0.8
	tileSize        = 16
	mplusFaceSource *text.GoTextFaceSource
	coin_anim       = 0
	plant_anim      = 0
	chicken_anim    = 0
)

type Game struct {
	Player            *Characters
	workers           []*Characters
	coins             []*Objects
	chickens          []*Objects
	eggs              []*Objects
	house             []*Objects
	plants            []*Objects
	buddaSpawnItems   []*Objects
	lastUpdate        time.Time
	tick              bool
	fullWindow        bool
	gamePause         bool
	village           *ebiten.Image
	bgImg             *ebiten.Image
	tilemapImg        *ebiten.Image
	tilemapImgWater   *ebiten.Image
	plantImg          *ebiten.Image
	workImg           *ebiten.Image
	workerIdleImg     *ebiten.Image
	coinImg           *ebiten.Image
	chickenImg        *ebiten.Image
	eggImg            *ebiten.Image
	infoBoxSpite      *Sprite
	addBottonImg      *widget.ButtonImage
	smokeSprite       *Sprite
	tilemapJSON1      *tilemaps.TilemapJSON
	tilemapJSON2      *tilemaps.TilemapJSON
	tilemapJSON3      *tilemaps.TilemapJSON
	scene             int
	exitGame          bool
	buddaAnimCounter  int
	buddaSpawnCounter int
}
type Sprite struct {
	img          *ebiten.Image
	pos          Point
	prePos       Point
	rectPos      image.Rectangle
	rectTop      Point // Sprite amination
	rectBot      Point // Sprite amination
	active       bool
	frameCounter int
	frame        int
}
type Characters struct {
	*Sprite
	Dir
	speed         float64
	dest          Point
	coin          int
	wallet        int
	basketSize    int
	tomatoBasket  int
	wheatBasket   int
	chicken       int
	chicken_count int
	egg           int
}
type Objects struct {
	*Sprite
	variety  string
	dest     Point
	picked   bool
	pickable bool
}
type Point struct {
	x, y float64
}
type Dir struct {
	down, up, right, left bool
}

// Idle workers faceing front animation
func (g *Game) idleWorkers(i int) {
	// show animation subImage
	if g.tick {
		g.workers[i].rectTop.x = imgSize - imgSize // 0
		g.workers[i].rectTop.y = imgSize - imgSize // 0
		g.workers[i].rectBot.x = imgSize           // 48
		g.workers[i].rectBot.y = imgSize           // 48
	} else {
		g.workers[i].rectTop.x = imgSize
		g.workers[i].rectTop.y = imgSize - imgSize
		g.workers[i].rectBot.x = imgSize * 2
		g.workers[i].rectBot.y = imgSize
	}
}

// return random point position
func randomPoint() Point {
	//rand.Seed(time.Now().UnixNano())
	x := rand.Intn(screenWidth)
	y := rand.Intn(screenHeight)
	pos := Point{float64(x), float64(y)}
	return pos
}

// Idle faceing front animation
func (g *Game) idle() {
	// show animation subImage
	if g.tick {
		g.Player.rectTop.x = imgSize - imgSize // 0
		g.Player.rectTop.y = imgSize - imgSize // 0
		g.Player.rectBot.x = imgSize           // 48
		g.Player.rectBot.y = imgSize           // 48
	} else {
		g.Player.rectTop.x = imgSize
		g.Player.rectTop.y = imgSize - imgSize
		g.Player.rectBot.x = imgSize * 2
		g.Player.rectBot.y = imgSize
	}
}

// set new position and player images(animation)
func (g *Game) dirDown() {
	if g.Player.Dir.right || g.Player.Dir.left {
		g.Player.speed = PlayerSpeed * diagonalSpeed
	}
	g.Player.pos.y += g.Player.speed
	// show animation subImage
	if g.tick {
		g.Player.rectTop.x = imgSize * 2
		g.Player.rectTop.y = imgSize - imgSize
		g.Player.rectBot.x = imgSize * 3
		g.Player.rectBot.y = imgSize
	} else {
		g.Player.rectTop.x = imgSize * 3
		g.Player.rectTop.y = imgSize - imgSize
		g.Player.rectBot.x = imgSize * 4
		g.Player.rectBot.y = imgSize
	}
	g.Player.Dir.down = false
	g.Player.Dir.right = false
	g.Player.Dir.left = false
	g.Player.speed = PlayerSpeed
	// if player moving, infoBox is closing
	g.infoBoxSpite.active = false
}
func (g *Game) dirUp() {
	if g.Player.Dir.right || g.Player.Dir.left {
		g.Player.speed = PlayerSpeed * diagonalSpeed
	}
	g.Player.pos.y -= g.Player.speed
	// show animation subImage
	if g.tick {
		g.Player.rectTop.x = imgSize * 2
		g.Player.rectTop.y = imgSize
		g.Player.rectBot.x = imgSize * 3
		g.Player.rectBot.y = imgSize * 2
	} else {
		g.Player.rectTop.x = imgSize * 3
		g.Player.rectTop.y = imgSize
		g.Player.rectBot.x = imgSize * 4
		g.Player.rectBot.y = imgSize * 2
	}
	g.Player.Dir.up = false
	g.Player.Dir.right = false
	g.Player.Dir.left = false
	g.Player.speed = PlayerSpeed
	// if player moving, infoBox is closing
	g.infoBoxSpite.active = false
}
func (g *Game) dirLeft() {
	if g.Player.Dir.up || g.Player.Dir.down {
		g.Player.speed = PlayerSpeed * diagonalSpeed
	}
	g.Player.pos.x -= g.Player.speed
	// show animation subImage
	if g.tick {
		g.Player.rectTop.x = imgSize * 2
		g.Player.rectTop.y = imgSize * 2
		g.Player.rectBot.x = imgSize * 3
		g.Player.rectBot.y = imgSize * 3
	} else {
		g.Player.rectTop.x = imgSize * 3
		g.Player.rectTop.y = imgSize * 2
		g.Player.rectBot.x = imgSize * 4
		g.Player.rectBot.y = imgSize * 3
	}
	g.Player.Dir.left = false
	g.Player.Dir.up = false
	g.Player.Dir.down = false
	g.Player.speed = PlayerSpeed
	// if player moving, infoBox is closing
	g.infoBoxSpite.active = false
}
func (g *Game) dirRight() {
	if g.Player.Dir.up || g.Player.Dir.down {
		g.Player.speed = PlayerSpeed * diagonalSpeed
	}
	g.Player.pos.x += g.Player.speed
	// show animation subImage
	if g.tick {
		g.Player.rectTop.x = imgSize - imgSize
		g.Player.rectTop.y = imgSize * 3
		g.Player.rectBot.x = imgSize
		g.Player.rectBot.y = imgSize * 4
	} else {
		g.Player.rectTop.x = imgSize * 2
		g.Player.rectTop.y = imgSize * 3
		g.Player.rectBot.x = imgSize * 3
		g.Player.rectBot.y = imgSize * 4
	}
	g.Player.Dir.right = false
	g.Player.Dir.up = false
	g.Player.Dir.down = false
	g.Player.speed = PlayerSpeed
	// if player moving, infoBox is closing
	g.infoBoxSpite.active = false
}

// check collision character with charakter: Player-worker
func (g *Game) Collision_Character_Character(obj Characters, char Characters) bool {
	// Player....
	character_position := image.Rect(
		int(char.pos.x+imgSize/4),
		int(char.pos.y+imgSize/4),
		int(char.pos.x+imgSize/2),
		int(char.pos.y+imgSize/2))

	// Worker
	object_position := image.Rect(
		int(obj.pos.x),
		int(obj.pos.y),
		int(obj.pos.x+imgSize/2),
		int(obj.pos.y+imgSize/2))

	return object_position.Overlaps(character_position)
}

// check collision char-obj: worker-plant
func (g *Game) Collision_worker_plant(worker Characters, plant Objects) bool {
	// Player....
	worker_position := image.Rect(
		int(worker.pos.x+imgSize/4),
		int(worker.pos.y+imgSize/4),
		int(worker.pos.x+imgSize/2),
		int(worker.pos.y+imgSize/2))

	// Worker
	plant_position := image.Rect(
		int(plant.pos.x),
		int(plant.pos.y),
		int(plant.pos.x+imgSize/2),
		int(plant.pos.y+imgSize/2))

	if worker_position.Overlaps(plant_position) {
		g.smokeSprite.active = true
		return true
	}
	return false
}

// check collision Objects with charakter: any obj-Player
func (g *Game) Collision_Object_Caracter(obj Objects, char Characters) bool {
	// Player....
	charakter_position := image.Rect(
		int(char.pos.x+imgSize/4),
		int(char.pos.y+imgSize/4),
		int(char.pos.x+imgSize/2),
		int(char.pos.y+imgSize/2))

	object_position := image.Rect(
		int(obj.pos.x),
		int(obj.pos.y),
		int(obj.pos.x+imgSize/2),
		int(obj.pos.y+imgSize/2))

	if obj.variety == "coin" || obj.variety == "wheat" || obj.variety == "tomato" {
		object_position = image.Rect(
			int(obj.pos.x-imgSize/4+10),
			int(obj.pos.y-imgSize/4+10),
			int(obj.pos.x+imgSize/4),
			int(obj.pos.y+imgSize/4))
	}
	if obj.variety == "budda" {
		object_position = image.Rect(
			int(obj.pos.x),
			int(obj.pos.y),
			int(obj.pos.x+imgSize/2),
			int(obj.pos.y+imgSize/2))
	}
	if obj.variety == "house" {
		object_position = image.Rect(
			int(obj.pos.x),
			int(obj.pos.y),
			int(obj.pos.x+houseTileSize-10),
			int(obj.pos.y+imgSize-10))
	}
	if obj.variety == "small_house" {
		object_position = image.Rect(
			int(obj.pos.x),
			int(obj.pos.y),
			int(obj.pos.x+imgSize-10),
			int(obj.pos.y+imgSize-10))
	}

	if object_position.Overlaps(charakter_position) {
		return true
	}
	return false
}

// TEST collision point - point
func (g *Game) checkCollision(p1 Point, p2 Point) bool {
	if p1.x >= p2.x-imgSize &&
		p1.x <= p2.x+imgSize &&
		p1.y >= p2.y-imgSize &&
		p1.y <= p2.y+imgSize {
		return true
	}
	return false
}

// collide with budda Action: set new player pos, span workers and Items, chose sceens ...
func (g *Game) buddaCollision() {
	// Portal Player to new pos
	g.Player.pos.x = screenWidth/2 + 20
	g.Player.pos.y = screenHeight/2 + 60
	// playSound
	playSound(audioFx)
	// Check if player has Tomatos and have a big wallet for the coins
	if g.Player.tomatoBasket > 0 && g.Player.coin < g.Player.wallet {
		g.Player.tomatoBasket--
		g.Player.coin += 2
		playSound(audioCoin)
		g.buddaSpawnCounter++ // count upp level
	}
	if g.Player.wheatBasket > 0 && g.Player.coin < g.Player.wallet {
		g.Player.wheatBasket--
		g.Player.coin += 1
		playSound(audioCoin)
		g.buddaSpawnCounter++ // count upp level
	}
	if g.Player.egg > 0 {
		g.Player.egg--
		g.eggs[0].active = false
		g.buddaSpawnItems[chest].active = true // show chest TEST
		g.buddaSpawnItems[chest].pickable = true
		g.buddaSpawnItems[chest].picked = false
		g.coins[0].active = true
		g.coins[0].picked = false
	}
	if g.buddaSpawnCounter > 3 {
		// add action: spawn workers
		g.workers[2].active = true
		if g.buddaSpawnCounter > 4 {
			g.workers[3].active = true
		}
		if g.buddaSpawnCounter > 5 {
			g.workers[4].active = true
		}
		if g.buddaSpawnCounter > 6 {
			g.workers[5].active = true
		}
		if g.buddaSpawnCounter > 7 {
			g.workers[6].active = true
		}
		if g.buddaSpawnCounter > 8 {
			g.workers[7].active = true
		}
		if g.buddaSpawnCounter > 9 {
			g.workers[8].active = true
		}
		if g.buddaSpawnCounter > 10 {
			g.scene = 1 // from scene 0 to 1
			g.workers[9].active = true
			for _, house := range g.house {
				house.active = false
				if house.variety == "new_house" ||
					house.variety == "new_house_small" ||
					house.variety == "chicken_house" {
					house.active = true
				}
			}
		}
	}
	g.workers[0].active = true // activate 2 workers at start
	g.workers[1].active = true

	// dopp all item if to greedy
	if g.Player.tomatoBasket == 5 || g.Player.wheatBasket == 5 || g.Player.coin == 5 {
		g.Player.tomatoBasket = 0
		g.Player.wheatBasket = 0
		g.Player.coin = 0
	}
}

// Move Workers to dest pos
func (g *Game) moveCharacters(c *Characters) {
	if c.pos != c.dest {
		c.img = g.workerIdleImg
		if c.pos.x < c.dest.x {
			c.pos.x++
		}
		if c.pos.x > c.dest.x {
			c.pos.x--
		}
		if c.pos.y < c.dest.y {
			c.pos.y++
		}
		if c.pos.y > c.dest.y {
			c.pos.y--
		}
	} else {
		c.img = g.workImg
	}
}

// check before moving chicken
func (g *Game) checkChickenMovment(c *Objects) {
	// chicken is running free and Player can pick it up
	if c.pickable {
		g.moveChickenToDest(c)
	}
	// if chicken is in the chicken house. Chicken is picked
	if c.pickable == false && c.picked {
		g.moveChickenToDest(c)
		// set c.dest to chicken_house.pos
	}
	//else  Player is carring chicken. Don't move chicken
}

// Move chicken to dest pos
func (g *Game) moveChickenToDest(c *Objects) {
	speed := 0.5
	if c.pos != c.dest {
		c.img = g.workerIdleImg
		if c.pos.x < c.dest.x {
			c.pos.x += speed
		}
		if c.pos.x > c.dest.x {
			c.pos.x -= speed
		}
		if c.pos.y < c.dest.y {
			c.pos.y += speed
		}
		if c.pos.y > c.dest.y {
			c.pos.y -= speed
		}
	} else {
		//c.img = g.workImg
	}
}

// check Animation tick every 60 FPS
func (g *Game) plantFrameAnim(plant *Objects) {
	var speed = 120 // wait two sec to next interation
	if plant.frameCounter < speed*5 {
		plant.frameCounter++
	}
	if plant.frameCounter < speed {
		plant.frame = 1
	} else if plant.frameCounter < speed*2 {
		plant.frame = 2
	} else if plant.frameCounter < speed*3 {
		plant.frame = 3
	} else if plant.frameCounter < speed*4 {
		plant.frame = 4
	} else if plant.frameCounter < speed*5 {
		plant.frame = 5
		plant.pickable = true
	}
}

// animation run once, when it's dune you can pick it.
func (g *Game) updateFourFrameAnimOnce(obj *Objects) {
	var speed = 120 // wait two sec to next interation
	if obj.frameCounter < speed*5 {
		obj.frameCounter++
	}
	if obj.frameCounter < speed {
		obj.frame = 1
	} else if obj.frameCounter < speed*2 {
		obj.frame = 2
	} else if obj.frameCounter < speed*3 {
		obj.frame = 3
	} else if obj.frameCounter < speed*4 {
		obj.frame = 4
		obj.pickable = true
	}
}

// ////////// Update:  Collision, Movement, Anim_frame, Anim_tick. ////////// //
func (g *Game) Update() error {
	// Exit game with "q" key
	if g.exitGame {
		return ebiten.Termination
	}

	g.Player.prePos = g.Player.pos // save old position before readKeys()
	g.readKeys()                   // read keys and move player
	g.coin_animation()

	////////////////////////////////////r
	// check Animation tick every 60 FPS. 2 values On or Off
	g.animTick()

	// pause all Update()
	if g.gamePause {
		return nil
	}

	// Chicken walk animation. And move chicken to random destination, Collision
	for _, chicken := range g.chickens {
		chicken.frame = g.fourTickAnim(chicken.frame)
		g.checkChickenMovment(chicken)
		// if chicken reached dest, set new dest
		if g.checkCollision(chicken.pos, chicken.dest) && chicken.pickable {
			chicken.dest = randomPoint()
		}
		// chicken in the chickenhouse. Set new dest
		if g.checkCollision(chicken.pos, chicken.dest) && !chicken.pickable && chicken.picked {
			// set random pos around the chickenhouse
			pos := randomPoint()
			if pos.x <= 500 {
				pos.x = 500
			} else if pos.x >= screenWidth-20 {
				pos.x = screenWidth - 20
			}
			if pos.y <= screenHeight/2-50 {
				pos.y = screenHeight/2 - 50
			} else if pos.y >= screenHeight-100 {
				pos.y = screenHeight - 100
			}
			chicken.dest = pos
		}
	}

	// plants animation
	for _, plant := range g.plants {
		if plant.active && plant.frame < 5 {
			g.plantFrameAnim(plant)
		}
	}
	// TEST Move workers to new dest pos for every new scene
	for i, w := range g.workers { // Idle animation for all workers
		g.idleWorkers(i)
		g.moveCharacters(g.workers[i])

		if w.coin > 0 { // TEST
			w.dest = g.plants[i].pos
			w.img = g.workImg
			g.plants[i].picked = false
			g.plants[i].active = true
		}
		if g.plants[i].picked {
			w.img = g.workerIdleImg
			g.workers[i].dest = Point{200 + (float64(i) * 30), 90}
		}

		//		// TEST
		//		if g.scene == 2 {
		//			g.workers[i].dest = Point{180 + (float64(i) * 40), 300}
		//		} else if g.scene == 3 {
		//			g.workers[i].dest = Point{30 + (float64(i) * 30), 20}
		//		} else if g.scene == 1 {
		//			g.workers[i].dest = Point{50, 10 + (float64(i) * 30)}
		//		} else if g.scene == 0 {
		//			g.workers[i].dest = Point{200 + (float64(i) * 20), 90}
		//		}

	}

	// Player border collision - Go to next sceen
	if g.Player.pos.x < 0-imgSize/2 {
		g.Player.pos.x = screenWidth - imgSize/2
		if g.scene > 0 {
			g.scene--
		}
	} else if g.Player.pos.x > screenWidth-imgSize/2 {
		g.Player.pos.x = 0 - imgSize/2
		if g.scene < 3 {
			g.scene++
		}
	} else if g.Player.pos.y < 0-imgSize/2 {
		g.Player.pos.y = screenHeight - imgSize/2
	} else if g.Player.pos.y > screenHeight {
		g.Player.pos.y = 0 - imgSize/2
	}
	//Player collide with []workers
	for i := range g.workers {
		if g.Collision_Character_Character(*g.workers[i], *g.Player) {
			if g.Player.coin > 0 && g.workers[i].active { // have coin and worker is active
				if g.workers[i].coin < 1 { // take only one coin
					g.workers[i].coin++
					g.Player.coin--
					playSound(audioCoin)
				}
				g.smokeSprite.active = true
				// move workers to new dest
				g.workers[i].dest = g.plants[i].pos // set worker.dest to plant.pos
				g.moveCharacters(g.workers[i])
			}
		}
	}

	// TEST set animation length for budda
	if g.buddaAnimCounter < 1 {
		g.buddaAnimCounter++
	} else {
		g.buddaAnimCounter = 0
		g.house[5].active = false
		g.house[6].active = false
		g.house[7].active = false
		g.house[8].active = false
		g.house[9].active = false
	}
	if g.scene == 0 {
		g.house[5].active = true // old_budda_image
	}
	if g.scene == 1 {
		g.house[9].active = true // gold_budda_image/
	}
	if g.buddaAnimCounter < 0 {
		g.budda_animation()
	}
	//Player collide with []house or budda_house or chicken_house
	for _, house := range g.house {
		if g.Collision_Object_Caracter(*house, *g.Player) {
			g.Player.pos = g.Player.prePos
			g.smokeSprite.active = true
			if house.variety == "budda" {
				g.buddaCollision()
				g.buddaAnimCounter = -60
			}
			if house.variety == "chicken_house" && g.Player.chicken > 0 {
				g.Player.chicken_count++
				g.Player.chicken--
				playSound(audioFx)
				if g.Player.chicken_count > 9 { // 10 chicken in the chicken_house
					g.eggs[1].active = true
					g.eggs[1].pickable = true
					g.Player.chicken_count = 0 // reset counter
					playSound(audioSecret)
					// set all chicken free
					for _, c := range g.chickens {
						c.active = true
						c.pickable = true
						c.picked = false
					}
				}
			}
		}
	}
	//Player collide with []plants if active
	for i := range g.plants {
		if g.Collision_Object_Caracter(*g.plants[i], *g.Player) {
			if g.plants[i].pickable && g.Player.tomatoBasket+g.Player.wheatBasket <= g.Player.basketSize {
				// pick plant
				playSound(audioFx)
				g.smokeSprite.active = true
				g.workers[i].coin = 0        // drop coint when plant are picked
				g.plants[i].active = false   // active animation
				g.plants[i].pickable = false // can be picked
				g.plants[i].picked = true    // Is picked
				g.plants[i].frame = 1        // set back to first anim-frame
				g.plants[i].frameCounter = 0 // counter back to zero
				if g.plants[i].variety == tomato {
					g.Player.tomatoBasket++
				} else if g.plants[i].variety == wheat {
					g.Player.wheatBasket++
				}
			}
		}
	}
	// Player collide with []coin
	for i := range g.coins {
		if g.Collision_Object_Caracter(*g.coins[i], *g.Player) && g.coins[i].picked == false {
			if g.Player.coin < g.Player.wallet { // add coins to your wallet
				g.Player.coin++
				playSound(audioCoin)
				g.coins[i].picked = true
				//				g.coins[i].pos = Point{
				//					x: -100,
				//					y: -100,
				//				}
			}
		}
	}
	// Player collide with chicken
	for _, chicken := range g.chickens {
		if g.Collision_Object_Caracter(*chicken, *g.Player) && chicken.pickable {
			g.smokeSprite.active = true
			if g.Player.chicken < 1 { // pick one at a time
				g.Player.chicken++
				chicken.pickable = false
				chicken.picked = true
				chicken.pos = Point{550, 150}
				chicken.dest = Point{570, 250}
			}
		}
	}
	// Player collide with Eggs
	for _, egg := range g.eggs {
		if g.Collision_Object_Caracter(*egg, *g.Player) && egg.pickable && egg.active {
			g.smokeSprite.active = true
			if g.Player.egg < 1 { // pick one at a time
				g.Player.egg++
				egg.pickable = false
				egg.picked = true
				egg.active = false
				playSound(audioSecret)
			}
		}
	}

	// Player collide with Chest
	for _, c := range g.buddaSpawnItems {
		if g.Collision_Object_Caracter(*c, *g.Player) && c.pickable && c.active {
			g.smokeSprite.active = true
			c.pickable = false
			c.picked = true
			c.active = false
			playSound(audioChest)
			if g.Player.wallet < 5 { // max 6 item at a time
				g.Player.wallet++
			}
			if g.Player.basketSize < 5 { // max 6 item at a time
				g.Player.basketSize++
			}
		}
	}

	// last in Update()
	return nil
}

// ////////// Draw function Draw all item at 60 fps ////////// //
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(dark_green) // background collor

	// 4 different sceens. Sceen 0 only a background img. 1-3 tilemaps
	op := &ebiten.DrawImageOptions{}
	if g.scene == 0 {
		///////// draw background ///////////
		op.GeoM.Translate(20, 20)

		screen.DrawImage(
			g.bgImg.SubImage(
				image.Rect(0, 0, 600, 370),
			).(*ebiten.Image),
			op,
		)
		op.GeoM.Reset()
	}
	if g.scene == 1 {
		/////////// draw bg tile layers ////////////
		for _, layer := range g.tilemapJSON1.Layers {
			for index, id := range layer.Data {
				x := index % layer.Width
				y := index / layer.Width
				x *= tileSize
				y *= tileSize

				srcX := (id - 1) % 22
				srcY := (id - 1) / 22
				srcX *= tileSize
				srcY *= tileSize

				op.GeoM.Translate(float64(x), float64(y))
				screen.DrawImage(
					g.tilemapImg.SubImage(image.Rect(srcX, srcY, srcX+tileSize, srcY+tileSize)).(*ebiten.Image),
					op,
				)
				op.GeoM.Reset()
			}
		}
	}
	if g.scene == 2 {
		/////////// draw bg tile layers ////////////
		for _, layer := range g.tilemapJSON2.Layers {
			for index, id := range layer.Data {
				x := index % layer.Width
				y := index / layer.Width
				x *= tileSize
				y *= tileSize

				srcX := (id - 1) % 22
				srcY := (id - 1) / 22
				srcX *= tileSize
				srcY *= tileSize

				op.GeoM.Translate(float64(x), float64(y))
				screen.DrawImage(
					g.tilemapImg.SubImage(image.Rect(srcX, srcY, srcX+tileSize, srcY+tileSize)).(*ebiten.Image),
					op,
				)
				op.GeoM.Reset()
			}
		}
	}
	if g.scene == 3 {
		/////////// draw bg tile layers from waterTileImg ////////////
		for _, layer := range g.tilemapJSON3.Layers {
			for index, id := range layer.Data {
				x := index % layer.Width
				y := index / layer.Width
				x *= tileSize
				y *= tileSize

				srcX := (id - 1) % 28
				srcY := (id - 1) / 28
				srcX *= tileSize
				srcY *= tileSize

				op.GeoM.Translate(float64(x), float64(y))
				screen.DrawImage(
					g.tilemapImgWater.SubImage(image.Rect(srcX, srcY, srcX+tileSize, srcY+tileSize)).(*ebiten.Image),
					op,
				)
				op.GeoM.Reset()
			}
		}
	}

	//// draw chickens ////
	for i := range g.chickens {
		g.drawChicken(screen, g.chickens[i].pos, g.chickens[i].frame)
	}
	//// draw eggs ////
	for _, egg := range g.eggs {
		if egg.active {
			g.drawItem(screen, egg.pos, egg.img, Point{16, 16})
		}
	}
	// draw budda_spawn_item
	for _, buddaItem := range g.buddaSpawnItems {
		if buddaItem.active == true {
			tileSize = 16
			g.drawChestAnim(screen, buddaItem.pos, buddaItem.img, tileSize)
		}
	}

	/////////// draw all HOUSES big and small  ////////////
	for _, house := range g.house {
		if house.active {
			opt := &ebiten.DrawImageOptions{}
			opt.GeoM.Translate(house.pos.x, house.pos.y) // house position x, y
			screen.DrawImage(
				house.img.SubImage(
					house.rectPos,
				).(*ebiten.Image),
				opt,
			)
			opt.GeoM.Reset()
		}
	}

	/// Draw COIN at same pos as Game constructor g.coins.pos in main() ///
	for i := 0; i < 10; i++ {
		if i >= 2 {
			g.coins[i].picked = true
		}
		if i < 2 {
			g.drawCoin(screen, g.coins[i].pos.x, g.coins[i].pos.y, *g.coins[i], i)
		}
	}

	/// Draw WORKERS /// if active. buddaSpawnLevel diside if active
	for i := range g.workers {
		// draw coin carring on workers head
		g.carry_objects(screen, g.workers[i].pos.x, g.workers[i].pos.y, g.workers[i].coin, g.coinImg, Point{10, 10})
		// draw all workers
		if g.workers[i].active {
			g.drawWorker(screen, g.workers[i].pos.x, g.workers[i].pos.y, i)
		}
	}

	///////// draw COINS, CHICKENS and PLANTS player caring on the head. SubImg 0,0,10,10 /////////
	g.carry_objects(screen, g.Player.pos.x, g.Player.pos.y, g.Player.egg, g.eggImg, Point{16, 16})
	g.carry_objects(screen, g.Player.pos.x, g.Player.pos.y, g.Player.coin, g.coinImg, Point{10, 10})
	g.carry_objects(screen, g.Player.pos.x, g.Player.pos.y, g.Player.chicken, g.chickenImg, Point{16, 16})
	// SubImg 0,0,16,16
	g.carry_plant(screen, g.Player.pos.x, g.Player.pos.y, g.Player.tomatoBasket, g.plantImg, tomato)
	g.carry_plant(screen, g.Player.pos.x, g.Player.pos.y, g.Player.wheatBasket, g.plantImg, wheat)

	///// Draw all plants  if active ///
	for i := range g.plants {
		if g.plants[i].active {
			g.drawPlants(screen, g.plants[i].pos.x, g.plants[i].pos.y, g.plants[i].variety, g.plants[i].frame) // wheat and tomato
		}
	}

	// draw infoBox. Active with key: a
	g.drawinfoBox(screen, g.infoBoxSpite.img, g.infoBoxSpite.pos.x, g.infoBoxSpite.pos.y)
	g.menuText(screen) // add text to infoBoxSprite

	// if active
	g.drawSmoke(screen, g.Player.pos.x, g.Player.pos.y)

	///////// draw Player ///////////
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(g.Player.pos.x, g.Player.pos.y)
	// amination position to Player.img.SubImage(image.Rect(0, 0, imgSize, imgSize))
	screen.DrawImage(
		g.Player.img.SubImage(
			image.Rect(int(g.Player.rectTop.x), int(g.Player.rectTop.y), int(g.Player.rectBot.x), int(g.Player.rectBot.y)),
		).(*ebiten.Image),
		opts,
	)
	/////// TEST Draw player and house collision rect
	// vector.StrokeRect(screen, float32(g.Player.pos.x+imgSize/4),float32(g.Player.pos.y+imgSize/4),imgSize/2,imgSize/2,3.0,color.RGBA{122, 222, 0, 100},false)
	// vector.StrokeRect(screen,float32(g.housePos.x)+float32(g.house[0].rectPos.Min.X),float32(g.housePos.y)+float32(g.house[0].rectPos.Min.Y),houseTileSize,imgSize,3.0,color.RGBA{222, 122, 0, 100},false)

	// play pause sceen
	if g.gamePause {
		g.pause(screen)
		return
	}
}

// draw images caring on the head //
func (g *Game) carry_objects(screen *ebiten.Image, x, y float64, amount int, img *ebiten.Image, tile Point) {
	optst := &ebiten.DrawImageOptions{}
	for i := 3; i < 3+amount; i++ { // i=3 3 pix apart
		optst.GeoM.Translate(x+imgSize/2-3, y+float64(2.0*i)-10.0)

		screen.DrawImage(
			img.SubImage(
				image.Rect(0, 0, int(tile.x), int(tile.y)),
			).(*ebiten.Image),
			optst,
		)
		optst.GeoM.Reset()
	}
}

// draw images caring on the head //
func (g *Game) carry_plant(screen *ebiten.Image, x, y float64, amount int, img *ebiten.Image, varity string) {
	opt := &ebiten.DrawImageOptions{}
	for i := 5; i < 5+amount; i++ { // i=5 5 pix apart
		if varity == wheat {
			opt.GeoM.Translate(x+imgSize/2, y+float64(2.0*i)-10.0)
			screen.DrawImage(
				img.SubImage(
					image.Rect(16*5, 0, 16*6, 16),
				).(*ebiten.Image),
				opt,
			)
		}
		if varity == tomato {
			opt.GeoM.Translate(x+imgSize/2-16, y+float64(2.0*i)-10.0)
			screen.DrawImage(
				img.SubImage(
					image.Rect(16*5, 16, 16*6, 16*2),
				).(*ebiten.Image),
				opt,
			)
		}
		opt.GeoM.Reset()
	}
}

// Main Animation Tick. Check every 60 FPS. 2 values On or Off
func (g *Game) animTick() error {
	if time.Since(g.lastUpdate) < gameSpeed {
		return nil
	}
	if g.tick {
		g.tick = false
	} else {
		g.tick = true
	}
	g.lastUpdate = time.Now() // update lastUpdate
	return nil
}

func (g *Game) smoke_animation() {
	if g.tick {
		plant_anim = 32
		if time.Since(g.lastUpdate) < gameSpeed/2 {
			plant_anim = 32 * 2
		}
	} else {
		plant_anim = 32 * 3
		if time.Since(g.lastUpdate) < gameSpeed/2 {
			plant_anim = 32 * 4
		}
	}
}

// set new sprite.frame 4 times every tick
func (g *Game) fourTickAnim(spriteFrame int) int {
	if g.tick {
		spriteFrame = 16 * 0
		if time.Since(g.lastUpdate) < gameSpeed/2 {
			spriteFrame = 16 * 1
		}
	} else {
		spriteFrame = 16 * 2
		if time.Since(g.lastUpdate) < gameSpeed/2 {
			spriteFrame = 16 * 3
		}
	}
	return spriteFrame
}

// set new sprite.frame 4 times every tick
// frame = sprite.frame
// tileSize = image size 64*64 or 16*16 ...
func (g *Game) animation(frame, tileSize int) int {
	if g.tick {
		frame = tileSize
		if time.Since(g.lastUpdate) < gameSpeed/2 {
			frame = tileSize * 2
		}
	} else {
		frame = tileSize * 3
		if time.Since(g.lastUpdate) < gameSpeed/2 {
			frame = tileSize * 4
		}
	}
	return frame
}

func (g *Game) drawItem(screen *ebiten.Image, pos Point, img *ebiten.Image, botPos Point) {
	//g.animation(0, 64)
	option := &ebiten.DrawImageOptions{}
	option.GeoM.Translate(pos.x, pos.y) // position x, y on the screen
	screen.DrawImage(
		img.SubImage(
			image.Rect(0, 0, int(botPos.x), int(botPos.y)), // top and bottom position of the image
		).(*ebiten.Image),
		option,
	)
	option.GeoM.Reset()
}

func (g *Game) drawChestAnim(screen *ebiten.Image, pos Point, img *ebiten.Image, tileSize int) {
	topx, topy := tileSize, tileSize     // top start pos
	botx, boty := tileSize*2, tileSize*2 // botstart pos
	if g.tick {                          // img 2. the animation
		topx = tileSize * 4 // pos 4 on Chest.png (tileSize space beteen chest)
		botx = tileSize * 5
	}
	g.animation(0, 64)
	option := &ebiten.DrawImageOptions{}
	option.GeoM.Translate(pos.x, pos.y) // position x, y on the screen
	screen.DrawImage(
		img.SubImage(
			image.Rect(int(topx), int(topy), int(botx), int(boty)), // top and bottom position of the image
		).(*ebiten.Image),
		option,
	)
	option.GeoM.Reset()
}
func (g *Game) drawChicken(screen *ebiten.Image, pos Point, frame int) {
	g.animation(0, 64)
	option := &ebiten.DrawImageOptions{}
	option.GeoM.Translate(pos.x, pos.y) // position x, y
	screen.DrawImage(
		g.chickenImg.SubImage(
			image.Rect(frame, 16, frame+16, 32), //row2, first interation: x=0,16 y=16,32
		).(*ebiten.Image),
		option,
	)
	option.GeoM.Reset()
}
func (g *Game) budda_animation() {
	g.house[7].active = false
	if g.tick {
		g.house[6].active = true
		if time.Since(g.lastUpdate) < gameSpeed/2 {
			g.house[6].active = false
			g.house[9].active = true
		}
	} else {
		g.house[9].active = false
		g.house[8].active = true
		if time.Since(g.lastUpdate) < gameSpeed/2 {
			g.house[8].active = false
			g.house[7].active = true
		}
	}
}
func (g *Game) drawSmoke(screen *ebiten.Image, x, y float64) {
	if g.smokeSprite.active {
		g.smoke_animation()
		option := &ebiten.DrawImageOptions{}
		option.GeoM.Translate(x, y) // position x, y
		screen.DrawImage(
			g.smokeSprite.img.SubImage(
				image.Rect(plant_anim, 0, plant_anim+32, 32),
			).(*ebiten.Image),
			option,
		)
		option.GeoM.Reset()
		g.smokeSprite.active = false
	}
}

// TEST plants animation
func (g *Game) plant_animation(frame int) {
	plant_anim = 16 * frame
}
func (g *Game) drawPlants(screen *ebiten.Image, x, y float64, variety string, frame int) {
	g.plant_animation(frame) // activate animation
	option := &ebiten.DrawImageOptions{}
	option.GeoM.Translate(x, y) // position x, y
	if variety == "wheat" {
		screen.DrawImage(
			g.plantImg.SubImage(
				//image.Rect(plant_anim, 0, plant_anim+imgSize, imgSize),
				image.Rect(plant_anim, 0, plant_anim+16, 16),
			).(*ebiten.Image),
			option,
		)
		option.GeoM.Reset()
	} else if variety == "tomato" {
		screen.DrawImage(
			g.plantImg.SubImage(
				//image.Rect(plant_anim, 0, plant_anim+imgSize, imgSize),
				image.Rect(plant_anim, 16, plant_anim+16, 16+16),
			).(*ebiten.Image),
			option,
		)
		option.GeoM.Reset()

	}
}

func (g *Game) drawWorker(screen *ebiten.Image, x, y float64, i int) {
	option := &ebiten.DrawImageOptions{}
	option.GeoM.Translate(x, y) // worker position x, y
	screen.DrawImage(
		g.workers[i].img.SubImage(
			image.Rect(int(g.workers[i].rectTop.x), int(g.workers[i].rectTop.y), int(g.workers[i].rectBot.x), int(g.workers[i].rectBot.y)),
		).(*ebiten.Image),
		option,
	)
	option.GeoM.Reset()
}

func (g *Game) coin_animation() {
	if g.tick {
		coin_anim = 0
		if time.Since(g.lastUpdate) < gameSpeed/2 {
			coin_anim = 10
		}
	} else {
		coin_anim = 20
		if time.Since(g.lastUpdate) < gameSpeed/2 {
			coin_anim = 30
		}
	}
}
func (g *Game) drawCoin(screen *ebiten.Image, x, y float64, coin Objects, index int) {
	if coin.picked {
		g.coins[index].pos = Point{-100, -100} // outside of screen
		return
	}
	option := &ebiten.DrawImageOptions{}
	option.GeoM.Translate(x, y) // coin position x, y
	screen.DrawImage(
		g.coins[index].img.SubImage(
			image.Rect(coin_anim, 0, coin_anim+10, 10),
		).(*ebiten.Image),
		option,
	)
	option.GeoM.Reset()
}

// Arrowkeys to move or vim-keys "hjkl"
func (g *Game) readKeys() {
	if ebiten.IsKeyPressed(ebiten.KeyJ) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.dirDown()
		g.Player.Dir.down = true
	} else {
		g.idle()
	}
	if ebiten.IsKeyPressed(ebiten.KeyK) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.dirUp()
		g.Player.Dir.up = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyH) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.dirLeft()
		g.Player.Dir.left = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyL) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.dirRight()
		g.Player.Dir.right = true
	} else if inpututil.IsKeyJustPressed(ebiten.KeyF) { // Full screen
		g.fullScreen()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyQ) { // Quit the game
		g.quitGame()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyEscape) { // Pause the game
		g.pauseGame()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyA) { // Action key
		g.actionKey()
	} else if inpututil.IsKeyJustPressed(ebiten.Key0) { // scene 0
		g.scene = 0
	} else if inpututil.IsKeyJustPressed(ebiten.Key1) { //  scene 1
		g.scene = 1
	} else if inpututil.IsKeyJustPressed(ebiten.Key2) { //  scene 2
		g.scene = 2
	} else if inpututil.IsKeyJustPressed(ebiten.Key3) { //  scene 3
		g.scene = 3
	}
}

// F key for full screen
func (g *Game) fullScreen() {
	if !g.fullWindow {
		ebiten.SetFullscreen(true)
		g.fullWindow = true
	} else {
		g.fullWindow = false
		ebiten.SetFullscreen(false)
	}
}

// Q key for quit
func (g *Game) quitGame() {
	g.exitGame = true
}

// Escape-key to Pause the game
func (g *Game) pauseGame() {
	if !g.gamePause {
		g.gamePause = true
	} else {
		g.gamePause = false
	}
}
func (g *Game) pause(screen *ebiten.Image) {
	// Pause the game
	vector.DrawFilledRect(
		screen,
		float32(20),     // x position
		float32(20),     // y position
		screenWidth-40,  // width size
		screenHeight-40, // Height size
		blue_transp,
		true,
	)
	addText(screen, 32, "Pause", black, screenWidth+5, screenHeight/3+4)
	addText(screen, 32, "Pause", yellow, screenWidth, screenHeight/3)
	addText(screen, 16, "Pause the Game - Esc", yellow, screenWidth, screenHeight/3+100)
	addText(screen, 16, "Quit the game - q", yellow, screenWidth, screenHeight/3+150)
	addText(screen, 16, "Full screen - f", yellow, screenWidth, screenHeight/3+200)
	addText(screen, 16, "Action key - a", yellow, screenWidth, screenHeight/3+250)
	addText(screen, 16, "Change scene key: 0-3", purple, screenWidth, screenHeight/3+300)
	addText(screen, 20, "*********************", green, screenWidth, screenHeight/3+500)
}

func (g Game) menuText(screen *ebiten.Image) {
	if g.infoBoxSpite.active {
		addText(screen, 10, "Pause the Game - Esc", black, 780, 630)
		addText(screen, 10, "Quit the game - q", blue, 780, 650)
		addText(screen, 10, "Full screen - f", purple, 780, 670)
		addText(screen, 10, "Move - arrowkey", red, 780, 690)
	}
}

// Action-key "a"
func (g *Game) actionKey() {
	if !g.infoBoxSpite.active {
		g.infoBoxSpite.active = true
	} else {
		g.infoBoxSpite.active = false
	}
}

// draw infoBox background at x,y pos
func (g *Game) drawinfoBox(screen, img *ebiten.Image, x, y float64) {
	if g.infoBoxSpite.active {
		option := &ebiten.DrawImageOptions{}
		option.GeoM.Translate(x, y) // position x, y
		screen.DrawImage(
			img.SubImage(
				image.Rect(0, 0, 400, 64),
			).(*ebiten.Image),
			option,
		)
		option.GeoM.Reset()
	}
}

// TEST add button
func (g *Game) menu() {

	s, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		log.Fatal(err)
	}

	fontFace := &text.GoTextFace{
		Source: s,
		Size:   32,
	}
	button := widget.NewButton(
		// specify the images to use
		//widget.ButtonOpts.Image(g.addBottonImg),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Hello, World!", fontFace, &widget.ButtonTextColor{
			Idle: color.RGBA{0xdf, 0xf4, 0xff, 0xff},
		}),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:  30,
			Right: 30,
		}),
		// ... click handler, etc. ...
	)
	button.Update()
}

func addText(screen *ebiten.Image, textSize int, t string, color color.Color, width, height float64) {
	face := &text.GoTextFace{
		Source: mplusFaceSource,
		Size:   float64(textSize),
	}
	// t := "YOURE TEXT"
	w, h := text.Measure(
		t,
		face,
		face.Size,
	)
	op := &text.DrawOptions{}
	op.GeoM.Translate(
		width/2-w/2, height/2-h/2,
	)
	op.ColorScale.ScaleWithColor(color)
	text.Draw(
		screen,
		t,
		face,
		op,
	)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// check for errors
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Window properties
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Gopher Land")

	// Text, font
	textsource, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	checkErr(err)
	mplusFaceSource = textsource

	// TilemapJSON1
	tilemapJSON1, err := tilemaps.NewTilemapJSON("assets/map/level1_bg.json")
	checkErr(err)

	// TilemapJSON2
	tilemapJSON2, err := tilemaps.NewTilemapJSON("assets/map/level2_bg.json")
	checkErr(err)

	// TilemapJSON2 Water
	tilemapJSON3, err := tilemaps.NewTilemapJSON("assets/map/water_bg.json")
	checkErr(err)

	// load tilemapImg image to tilemap 1 - 2
	tilemapImg, _, err := ebitenutil.NewImageFromFile("assets/map/tileset_floor.png")
	checkErr(err)

	// load tilemapImg image to tilemap water
	tilemapImgWater, _, err := ebitenutil.NewImageFromFile("assets/map/TilesetWater.png")
	checkErr(err)

	// load village image
	old_village, _, err := ebitenutil.NewImageFromFile("assets/images/village_old.png")
	checkErr(err)

	// load village house image
	new_village, _, err := ebitenutil.NewImageFromFile("assets/images/TilesetHouse.png")
	checkErr(err)

	// load background image
	bgImg, _, err := ebitenutil.NewImageFromFile("assets/images/grass.png")
	checkErr(err)

	// load Player image
	playerImg, _, err := ebitenutil.NewImageFromFile("assets/images/playerBlue.png")
	checkErr(err)

	// load Worker image
	workerImg, _, err := ebitenutil.NewImageFromFile("assets/images/player.png")
	checkErr(err)

	// load Work image
	workImg, _, err := ebitenutil.NewImageFromFile("assets/images/workers.png")
	checkErr(err)

	// load coin image
	coinImg, _, err := ebitenutil.NewImageFromFile("assets/images/coin2.png")
	checkErr(err)

	// load chicken image
	chickenImg, _, err := ebitenutil.NewImageFromFile("assets/images/chicken.png")
	checkErr(err)

	// load chicken image
	eggImg, _, err := ebitenutil.NewImageFromFile("assets/images/Egg.png")
	checkErr(err)

	// load chicken_house image
	chicken_houseImg, _, err := ebitenutil.NewImageFromFile("assets/images/Chicken_House.png")
	checkErr(err)

	// load info box background image
	infoBoxImg, _, err := ebitenutil.NewImageFromFile("assets/images/InfoBox.png")
	checkErr(err)

	// load plants image
	plantImg, _, err := ebitenutil.NewImageFromFile("assets/images/plants.png")
	checkErr(err)

	// load plants image
	chestImg, _, err := ebitenutil.NewImageFromFile("assets/images/Chest.png")
	checkErr(err)

	// 	// load add-button image
	// 	addButton, _, err := ebitenutil.NewImageFromFile("assets/images/add-button64.png")
	// 	checkErr(err)

	// load smoke image
	smokeImg, _, err := ebitenutil.NewImageFromFile("assets/images/smoke.png")
	checkErr(err)

	// Game constructor. add Player
	g := &Game{
		Player: &Characters{
			Sprite: &Sprite{
				img: playerImg,
				pos: Point{305, 305},
				//pos: Point{screenWidth/2 - (imgSize / 2), screenHeight/2 - (imgSize / 2)},
			},
			speed:      PlayerSpeed,
			coin:       0,
			wallet:     2,
			basketSize: 2,
		},
	}
	g.Player.rectTop = Point{g.Player.pos.y + imgSize/4, g.Player.pos.y + imgSize/4}
	g.Player.rectBot = Point{imgSize / 2, imgSize / 2}

	// add 10 workers
	for i := 0; i < 10; i++ {
		g.workers = append(g.workers, &Characters{
			Sprite: &Sprite{
				img:     workerImg,
				pos:     Point{40, 20*float64(i) + 60},
				rectPos: image.Rect(0, 0, imgSize/2, imgSize/2),
			},
			speed: 1.5,
			dest:  Point{screenWidth - imgSize - (float64(i * imgSize)), screenHeight/2 - imgSize - (float64(i * imgSize))},
		})
	}
	for i := range g.workers { //set rectTop and rectBot for animation
		g.workers[i].rectTop = Point{0, 0}
		g.workers[i].rectBot = Point{imgSize, imgSize}
		g.workers[i].dest = Point{200 + (float64(i) * 30), 90}
		g.workers[i].active = false // start with inactive workers. buddaSpawn activate workers
	}

	// add 10 coins
	for i := 1; i < 11; i++ {
		g.coins = append(g.coins, &Objects{
			Sprite: &Sprite{
				img:     coinImg,
				pos:     Point{screenWidth/2 + 30 + float64(i)*10, screenHeight/2 + houseTileSize - 30.0},
				rectPos: image.Rect(0, 0, imgSize/2, imgSize/2),
			},
			variety: "coin",
		})
	}
	// add 5 wheat and 5 tomato. Even plant[0,2,4,...]=wheat, odd plant[1,...]=tomato
	for i := 0; i < 5; i++ {
		g.plants = append(g.plants, &Objects{
			Sprite: &Sprite{
				img:     plantImg,
				pos:     Point{178 + float64(i)*40, 300},
				rectPos: image.Rect(0, 0, imgSize/2, imgSize/2),
				active:  false,
			},
			variety: "wheat",
		})

		g.plants = append(g.plants, &Objects{
			Sprite: &Sprite{
				img:     plantImg,
				pos:     Point{60 + float64(i)*40, 40},
				rectPos: image.Rect(0, 0, imgSize/2, imgSize/2),
				active:  false,
			},
			variety: "tomato",
		})
	}
	// add 10 chickens
	for i := 1; i < 11; i++ {
		g.chickens = append(g.chickens, &Objects{
			Sprite: &Sprite{
				img:     chickenImg,
				pos:     randomPoint(), // start at random point
				rectPos: image.Rect(0, 0, imgSize/2, imgSize/2),
			},
			variety:  "chicken",
			pickable: true,
		})
	}
	// add 10 eggs
	for i := 1; i < 11; i++ {
		g.eggs = append(g.eggs, &Objects{
			Sprite: &Sprite{
				active:  false,
				img:     eggImg,
				pos:     Point{560, screenHeight/2 + float64(i)*10}, // start point
				rectPos: image.Rect(0, 0, imgSize/2, imgSize/2),
			},
			variety:  "egg",
			pickable: true,
		})
	}
	//	add house objects 0
	g.house = append(g.house, &Objects{ // house[0]old house with roof
		Sprite: &Sprite{
			img:     old_village,
			pos:     Point{250, houseTileSize},
			rectPos: image.Rect(0, 0, houseTileSize, imgSize),
			active:  true,
		},
		variety: "house",
	})
	g.house = append(g.house, &Objects{ // house[1]old house without roof
		Sprite: &Sprite{
			img:     old_village,
			pos:     Point{100, 100},
			rectPos: image.Rect(houseTileSize, 0, houseTileSize*2, imgSize),
			active:  true,
		},
		variety: "house",
	})
	g.house = append(g.house, &Objects{ // house[2]
		Sprite: &Sprite{
			img:     old_village,
			pos:     Point{400, imgSize},
			rectPos: image.Rect(houseTileSize*2+imgSize, 0, houseTileSize*2+imgSize*2, imgSize),
			active:  true,
		},
		variety: "small_house",
	})
	g.house = append(g.house, &Objects{ // house[3]
		Sprite: &Sprite{
			img:     old_village,
			pos:     Point{500, imgSize},
			rectPos: image.Rect(houseTileSize*2+imgSize, imgSize, houseTileSize*2+imgSize*2, imgSize*2),
			active:  true,
		},
		variety: "small_house",
	})
	////// NEW house //// 4
	g.house = append(g.house, &Objects{ // house[4]New house with roof
		Sprite: &Sprite{
			img:     new_village,
			pos:     Point{250, houseTileSize},
			rectPos: image.Rect(0, 0, houseTileSize, imgSize),
			active:  false,
		},
		variety: "new_house",
	})

	g.house = append(g.house, &Objects{ // house[5] old_budda //5
		Sprite: &Sprite{
			img:     old_village,
			pos:     Point{screenWidth/2 + houseTileSize, screenHeight/2 + houseTileSize},
			rectPos: image.Rect(0, imgSize, imgSize-16, imgSize*2-16),
			active:  true,
		},
		variety: "budda",
	})
	g.house = append(g.house, &Objects{ // house[6] gray
		Sprite: &Sprite{
			img:     new_village,
			pos:     Point{screenWidth/2 + houseTileSize, screenHeight/2 + houseTileSize},
			rectPos: image.Rect(imgSize*2-15, imgSize*6+16, imgSize*3-32, imgSize*7),
			active:  false,
		},
		variety: "budda",
	})
	g.house = append(g.house, &Objects{ // house[7] gray with pearl
		Sprite: &Sprite{
			img:     new_village,
			pos:     Point{screenWidth/2 + houseTileSize, screenHeight/2 + houseTileSize},
			rectPos: image.Rect(imgSize*1, imgSize*6+16, imgSize*2-16, imgSize*7),
			active:  false,
		},
		variety: "budda",
	})
	g.house = append(g.house, &Objects{ // house[8] orange
		Sprite: &Sprite{
			img:     new_village,
			pos:     Point{screenWidth/2 + houseTileSize, screenHeight/2 + houseTileSize},
			rectPos: image.Rect(imgSize*2-16, imgSize*5, imgSize*2+16, imgSize*6-16),
			active:  false,
		},
		variety: "budda",
	})
	g.house = append(g.house, &Objects{ // house[9] orange with pearl
		Sprite: &Sprite{
			img:     new_village,
			pos:     Point{screenWidth/2 + houseTileSize, screenHeight/2 + houseTileSize},
			rectPos: image.Rect(imgSize, imgSize*5, imgSize*2-16, imgSize*6-16),
			active:  false,
		},
		variety: "budda",
	})
	// NEW house, for Lever 2
	g.house = append(g.house, &Objects{ // New house with roof
		Sprite: &Sprite{
			img:     new_village,
			pos:     Point{100, 100},
			rectPos: image.Rect(houseTileSize, 0, houseTileSize*2, imgSize),
			active:  false,
		},
		variety: "new_house",
	})
	// NEW house, for Lever 2
	g.house = append(g.house, &Objects{ // New small_house with roof
		Sprite: &Sprite{
			img:     new_village,
			pos:     Point{400, imgSize},
			rectPos: image.Rect(houseTileSize*5-16, imgSize*6+16, houseTileSize*5+imgSize-16, imgSize*6+imgSize+16),
			active:  false,
		},
		variety: "new_house_small",
	})
	// NEW house, for Lever 2
	g.house = append(g.house, &Objects{ // New small_house with roof
		Sprite: &Sprite{
			img:     new_village,
			pos:     Point{500, imgSize},
			rectPos: image.Rect(houseTileSize*5-16, imgSize*6+16, houseTileSize*5+imgSize-16, imgSize*6+imgSize+16),
			active:  false,
		},
		variety: "new_house_small",
	})

	// chicken_house loads separat
	g.house = append(g.house, &Objects{
		Sprite: &Sprite{
			img:     chicken_houseImg,
			pos:     Point{550, screenHeight/2 - houseTileSize},
			rectPos: image.Rect(0, 0, imgSize, imgSize),
			active:  true,
		},
		variety: "chicken_house",
	})

	// add buddaSpawnItems objects
	g.buddaSpawnItems = append(g.buddaSpawnItems, &Objects{
		Sprite: &Sprite{
			img:     chestImg,
			pos:     Point{screenWidth / 2, screenHeight / 2},
			rectPos: image.Rect(0, 0, 32, 32),
			active:  false,
		},
		variety:  "chest",
		pickable: true,
	})

	// Add Images and tilemapJSON
	g.bgImg = bgImg
	g.village = old_village
	g.tilemapImg = tilemapImg
	g.tilemapImgWater = tilemapImgWater
	g.plantImg = plantImg
	g.workImg = workImg
	g.workerIdleImg = workerImg
	g.coinImg = coinImg
	g.chickenImg = chickenImg
	g.eggImg = eggImg

	// info box background
	g.infoBoxSpite = &Sprite{
		img:    infoBoxImg,
		pos:    Point{300, 300},
		active: true,
	}

	// smoke sprite
	g.smokeSprite = &Sprite{
		img:    smokeImg,
		pos:    Point{50, 50},
		active: false,
	}

	g.tilemapJSON1 = tilemapJSON1
	g.tilemapJSON2 = tilemapJSON2
	g.tilemapJSON3 = tilemapJSON3

	g.scene = 0 // scene or level, 4 different backgrounds

	////// play background music //////
	_ = audio.NewContext(SampleRate)
	stream, err := vorbis.DecodeWithSampleRate(SampleRate, bytes.NewReader(audioBG))
	checkErr(err)

	// infinite loop Bg music
	audioPlayer, err := audio.CurrentContext().NewPlayer(
		audio.NewInfiniteLoop(stream,
			int64(len(audioBG)*6*SampleRate)))
	checkErr(err)

	//audioPlayer, _ := audio.CurrentContext().NewPlayer(stream)
	// you pass the audio player to your game struct, and just call
	audioPlayer.SetVolume(0.2)
	audioPlayer.Play() //when you want your music to start, and
	// audioPlayer.Pause()

	//	// chose audio file to on scene 0-1
	//	if g.scene == 1 {
	//		playSound(audioVillage)
	//	}
	//	if g.scene == 0 {
	//		playSound(audioBG)
	//	}

	// Start game
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func playSound(sound []byte) {
	//_ = audio.NewContext(SampleRate)
	stream, err := vorbis.DecodeWithSampleRate(SampleRate, bytes.NewReader(sound))
	checkErr(err)
	audioPlayer, _ := audio.CurrentContext().NewPlayer(stream)
	// TEST add parameter to the func volume float 0.0 - 1.0
	audioPlayer.SetVolume(0.3)
	audioPlayer.Play()
}
