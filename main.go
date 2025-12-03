package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/lafriks/go-tiled"
	"github.com/setanarut/tilecollider"
)

const mapPath = "demo.tmx" // Path to your Tiled Map.

type mapGame struct {
	Level         *tiled.Map
	tileHash      map[uint32]*ebiten.Image
	drawableLevel *ebiten.Image
	collider      *tilecollider.Collider[int]
	demoPlayer    player
}

type player struct {
	x, y float64
	pict *ebiten.Image
}

func (m *mapGame) Update() error {
	Player_dx, Player_dy := getPlayerInput()
	final_dx, final_dy := m.collider.Collide(m.demoPlayer.x, m.demoPlayer.y,
		float64(m.demoPlayer.pict.Bounds().Dx()), float64(m.demoPlayer.pict.Bounds().Dy()),
		Player_dx, Player_dy, nil) //the final nil is a callback that we could have called if a collision occurs
	m.demoPlayer.x += final_dx
	m.demoPlayer.y += final_dy
	return nil
}

func (m mapGame) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func main() {
	// Parse .tmx file.
	gameMap, err := tiled.LoadFile(mapPath)
	windowWidth := gameMap.Width * gameMap.TileWidth
	windowHeight := gameMap.Height * gameMap.TileHeight
	ebiten.SetWindowSize(windowWidth, windowHeight)
	if err != nil {
		fmt.Printf("error parsing map: %s", err.Error())
		os.Exit(2)
	}
	ebitenImageMap := makeEbiteImagesFromMap(*gameMap)
	playerPict, _, err := ebitenutil.NewImageFromFile("boy2.png")
	if err != nil {
		fmt.Println("Error loading player image:", err)
	}
	var mapAsIntSlice [][]int = makeCollideMap(gameMap)
	oneLevelGame := mapGame{
		Level:         gameMap,
		tileHash:      ebitenImageMap,
		drawableLevel: ebiten.NewImage(windowWidth, windowHeight),
		collider:      tilecollider.NewCollider(mapAsIntSlice, gameMap.TileWidth, gameMap.TileHeight),
		demoPlayer:    player{x: 200, y: 150, pict: playerPict},
	}
	buildDrawableLevel(&oneLevelGame)
	err = ebiten.RunGame(&oneLevelGame)
	if err != nil {
		fmt.Println("Couldn't run game:", err)
	}
}

func makeCollideMap(gameMap *tiled.Map) [][]int {
	//here my map has one layer, in a more realistic example I might have the
	//0 layer be the ground and then the next layer have only the obstacles
	//the tilemap is unrolled, with the 2d array unrolled into a 1d array
	//we have to convert it to a 2d array for the tilecollider to work
	var mapAsIntSlice [][]int = make([][]int, gameMap.Height)
	for tileY := 0; tileY < gameMap.Height; tileY += 1 {
		//get each row of tiles
		mapAsIntSlice[tileY] = make([]int, gameMap.Width)
		for tileX := 0; tileX < gameMap.Width; tileX += 1 {
			mapTileID := int(gameMap.Layers[0].Tiles[tileY*gameMap.Width+tileX].ID)
			//the tile collider wants 0 for all open tiles and non-zero for all obstacles
			//I want the brown tiles to be the obstacles, in the map they are 1 and 4, because of the
			//way the tiled library adjusts by one, they will be 0 and 3 in the array
			//those will be zero, the rest will be 1
			if mapTileID == 0 || mapTileID == 3 {
				mapTileID = 1
			} else {
				mapTileID = 0
			}
			mapAsIntSlice[tileY][tileX] = mapTileID
		}
	}
	return mapAsIntSlice
}

func buildDrawableLevel(game *mapGame) {
	screen := game.drawableLevel
	drawOptions := ebiten.DrawImageOptions{}
	for tileY := 0; tileY < game.Level.Height; tileY += 1 {
		for tileX := 0; tileX < game.Level.Width; tileX += 1 {
			drawOptions.GeoM.Reset()
			TileXpos := float64(game.Level.TileWidth * tileX)
			TileYpos := float64(game.Level.TileHeight * tileY)
			drawOptions.GeoM.Translate(TileXpos, TileYpos)
			tileToDraw := game.Level.Layers[0].Tiles[tileY*game.Level.Width+tileX]
			ebitenTileToDraw := game.tileHash[tileToDraw.ID]
			screen.DrawImage(ebitenTileToDraw, &drawOptions)
		}
	}
}

func makeEbiteImagesFromMap(tiledMap tiled.Map) map[uint32]*ebiten.Image {
	idToImage := make(map[uint32]*ebiten.Image)
	for _, tile := range tiledMap.Tilesets[0].Tiles {
		ebitenImageTile, _, err := ebitenutil.NewImageFromFile(tile.Image.Source)
		if err != nil {
			fmt.Println("Error loading tile image:", tile.Image.Source, err)
		}
		idToImage[tile.ID] = ebitenImageTile
	}
	return idToImage
}

func (game mapGame) Draw(screen *ebiten.Image) {
	drawOptions := ebiten.DrawImageOptions{}
	screen.DrawImage(game.drawableLevel, &drawOptions)
	drawOptions.GeoM.Reset()
	drawOptions.GeoM.Translate(float64(game.demoPlayer.x), float64(game.demoPlayer.y))
	screen.DrawImage(game.demoPlayer.pict, &drawOptions)
}

func getPlayerInput() (dX, dY float64) {
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		dY = -3
	} else if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		dY = 3
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		dX = -3
	} else if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		dX = 3
	}
	return dX, dY

}
