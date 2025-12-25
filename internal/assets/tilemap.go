package assets

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// TiledMap represents a Tiled map (.tmx exported as JSON).
type TiledMap struct {
	Width      int          `json:"width"`
	Height     int          `json:"height"`
	TileWidth  int          `json:"tilewidth"`
	TileHeight int          `json:"tileheight"`
	Layers     []TiledLayer `json:"layers"`
	Tilesets   []Tileset    `json:"tilesets"`

	// Loaded tileset images
	TileImages map[int]*ebiten.Image
}

// TiledLayer represents a layer in a Tiled map.
type TiledLayer struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Width   int      `json:"width"`
	Height  int      `json:"height"`
	Data    []int    `json:"data"`
	Visible bool     `json:"visible"`
	Objects []Object `json:"objects,omitempty"`
}

// Object represents an object in an object layer.
type Object struct {
	ID         int            `json:"id"`
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	X          float64        `json:"x"`
	Y          float64        `json:"y"`
	Width      float64        `json:"width"`
	Height     float64        `json:"height"`
	Properties map[string]any `json:"properties,omitempty"`
}

// Tileset represents a tileset definition.
type Tileset struct {
	FirstGID    int    `json:"firstgid"`
	Name        string `json:"name"`
	TileWidth   int    `json:"tilewidth"`
	TileHeight  int    `json:"tileheight"`
	TileCount   int    `json:"tilecount"`
	Columns     int    `json:"columns"`
	Image       string `json:"image"`
	ImageWidth  int    `json:"imagewidth"`
	ImageHeight int    `json:"imageheight"`
}

// LoadTiledMap loads a Tiled map from JSON format.
func LoadTiledMap(filesystem fs.FS, mapPath string) (*TiledMap, error) {
	data, err := fs.ReadFile(filesystem, mapPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read map file: %w", err)
	}

	var tiledMap TiledMap
	if err := json.Unmarshal(data, &tiledMap); err != nil {
		return nil, fmt.Errorf("failed to parse map JSON: %w", err)
	}

	// Load tileset images
	tiledMap.TileImages = make(map[int]*ebiten.Image)
	loader := NewLoader(filesystem)
	mapDir := filepath.Dir(mapPath)

	for _, ts := range tiledMap.Tilesets {
		imagePath := filepath.Join(mapDir, ts.Image)

		img, err := loader.LoadImage(imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load tileset image %s: %w", ts.Image, err)
		}

		// Extract individual tiles
		cols := ts.Columns

		rows := ts.TileCount / cols
		for row := range rows {
			for col := range cols {
				tileID := ts.FirstGID + row*cols + col
				x := col * ts.TileWidth
				y := row * ts.TileHeight
				tileImg := img.SubImage(
					ebiten.NewImage(ts.TileWidth, ts.TileHeight).Bounds().Add(
						struct{ X, Y int }{x, y},
					),
				).(*ebiten.Image)
				tiledMap.TileImages[tileID] = tileImg
			}
		}
	}

	return &tiledMap, nil
}

// GetLayer returns a layer by name.
func (m *TiledMap) GetLayer(name string) *TiledLayer {
	for i := range m.Layers {
		if m.Layers[i].Name == name {
			return &m.Layers[i]
		}
	}

	return nil
}

// GetTileAt returns the tile GID at a given tile position in a layer.
func (m *TiledMap) GetTileAt(layer *TiledLayer, x, y int) int {
	if x < 0 || x >= layer.Width || y < 0 || y >= layer.Height {
		return 0
	}

	return layer.Data[y*layer.Width+x]
}

// DrawLayer draws a tile layer to the screen.
func (m *TiledMap) DrawLayer(screen *ebiten.Image, layer *TiledLayer, offsetX, offsetY float64) {
	if layer.Type != "tilelayer" || !layer.Visible {
		return
	}

	for y := 0; y < layer.Height; y++ {
		for x := 0; x < layer.Width; x++ {
			gid := m.GetTileAt(layer, x, y)
			if gid == 0 {
				continue // Empty tile
			}

			tileImg := m.TileImages[gid]
			if tileImg == nil {
				continue
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(
				float64(x*m.TileWidth)+offsetX,
				float64(y*m.TileHeight)+offsetY,
			)
			screen.DrawImage(tileImg, op)
		}
	}
}

// PixelWidth returns the map width in pixels.
func (m *TiledMap) PixelWidth() int {
	return m.Width * m.TileWidth
}

// PixelHeight returns the map height in pixels.
func (m *TiledMap) PixelHeight() int {
	return m.Height * m.TileHeight
}

// TMX support for raw XML format

// TMXMap represents a Tiled map in XML format.
type TMXMap struct {
	XMLName    xml.Name     `xml:"map"`
	Width      int          `xml:"width,attr"`
	Height     int          `xml:"height,attr"`
	TileWidth  int          `xml:"tilewidth,attr"`
	TileHeight int          `xml:"tileheight,attr"`
	Layers     []TMXLayer   `xml:"layer"`
	Tilesets   []TMXTileset `xml:"tileset"`
}

type TMXLayer struct {
	ID     int     `xml:"id,attr"`
	Name   string  `xml:"name,attr"`
	Width  int     `xml:"width,attr"`
	Height int     `xml:"height,attr"`
	Data   TMXData `xml:"data"`
}

type TMXData struct {
	Encoding string `xml:"encoding,attr"`
	Content  string `xml:",chardata"`
}

type TMXTileset struct {
	FirstGID   int      `xml:"firstgid,attr"`
	Name       string   `xml:"name,attr"`
	TileWidth  int      `xml:"tilewidth,attr"`
	TileHeight int      `xml:"tileheight,attr"`
	TileCount  int      `xml:"tilecount,attr"`
	Columns    int      `xml:"columns,attr"`
	Image      TMXImage `xml:"image"`
}

type TMXImage struct {
	Source string `xml:"source,attr"`
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
}

// LoadTMXMap loads a Tiled map from TMX (XML) format.
func LoadTMXMap(filesystem fs.FS, mapPath string) (*TiledMap, error) {
	data, err := fs.ReadFile(filesystem, mapPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read TMX file: %w", err)
	}

	var tmx TMXMap
	if err := xml.Unmarshal(data, &tmx); err != nil {
		return nil, fmt.Errorf("failed to parse TMX: %w", err)
	}

	// Convert to TiledMap
	tiledMap := &TiledMap{
		Width:      tmx.Width,
		Height:     tmx.Height,
		TileWidth:  tmx.TileWidth,
		TileHeight: tmx.TileHeight,
		Layers:     make([]TiledLayer, 0, len(tmx.Layers)),
		Tilesets:   make([]Tileset, 0, len(tmx.Tilesets)),
		TileImages: make(map[int]*ebiten.Image),
	}

	// Convert tilesets
	for _, ts := range tmx.Tilesets {
		tiledMap.Tilesets = append(tiledMap.Tilesets, Tileset{
			FirstGID:    ts.FirstGID,
			Name:        ts.Name,
			TileWidth:   ts.TileWidth,
			TileHeight:  ts.TileHeight,
			TileCount:   ts.TileCount,
			Columns:     ts.Columns,
			Image:       ts.Image.Source,
			ImageWidth:  ts.Image.Width,
			ImageHeight: ts.Image.Height,
		})
	}

	// Convert layers
	for _, layer := range tmx.Layers {
		tl := TiledLayer{
			ID:      layer.ID,
			Name:    layer.Name,
			Type:    "tilelayer",
			Width:   layer.Width,
			Height:  layer.Height,
			Visible: true,
		}

		// Parse CSV data
		if layer.Data.Encoding == "csv" || layer.Data.Encoding == "" {
			lines := strings.SplitSeq(strings.TrimSpace(layer.Data.Content), "\n")
			for line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				parts := strings.SplitSeq(strings.TrimRight(line, ","), ",")
				for p := range parts {
					p = strings.TrimSpace(p)
					if p == "" {
						continue
					}

					gid, _ := strconv.Atoi(p)
					tl.Data = append(tl.Data, gid)
				}
			}
		}

		tiledMap.Layers = append(tiledMap.Layers, tl)
	}

	// Load tileset images
	loader := NewLoader(filesystem)
	mapDir := filepath.Dir(mapPath)

	for _, ts := range tiledMap.Tilesets {
		imagePath := filepath.Join(mapDir, ts.Image)

		img, err := loader.LoadImage(imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load tileset image %s: %w", ts.Image, err)
		}

		cols := ts.Columns
		if cols == 0 {
			cols = ts.ImageWidth / ts.TileWidth
		}

		rows := ts.TileCount / cols
		if rows == 0 {
			rows = ts.ImageHeight / ts.TileHeight
		}

		for row := 0; row < rows; row++ {
			for col := 0; col < cols; col++ {
				tileID := ts.FirstGID + row*cols + col
				x := col * ts.TileWidth
				y := row * ts.TileHeight
				tileImg := img.SubImage(
					ebiten.NewImage(ts.TileWidth, ts.TileHeight).Bounds().Add(
						struct{ X, Y int }{x, y},
					),
				).(*ebiten.Image)
				tiledMap.TileImages[tileID] = tileImg
			}
		}
	}

	return tiledMap, nil
}
