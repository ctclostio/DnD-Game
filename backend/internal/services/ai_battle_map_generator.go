package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
	"github.com/google/uuid"
)

type AIBattleMapGenerator struct {
	llmProvider LLMProvider
	config      *AIConfig
	logger      *logger.LoggerV2
}

func NewAIBattleMapGenerator(provider LLMProvider, config *AIConfig, log *logger.LoggerV2) *AIBattleMapGenerator {
	return &AIBattleMapGenerator{
		llmProvider: provider,
		config:      config,
		logger:      log,
	}
}

// GenerateBattleMap creates a tactical map based on location description
func (abmg *AIBattleMapGenerator) GenerateBattleMap(ctx context.Context, req models.GenerateBattleMapRequest) (*models.BattleMap, error) {
	if !abmg.config.Enabled {
		return abmg.generateDefaultBattleMap(req), nil
	}

	// Determine grid size based on desired size
	gridX, gridY := abmg.determineGridSize(req.DesiredSize)

	prompt := abmg.buildBattleMapPrompt(req, gridX, gridY)

	systemPrompt := "You are an expert D&D 5e battle map designer. Create tactical, balanced maps that enhance combat encounters with interesting terrain features, cover, and environmental elements. Your response must be valid JSON."

	response, err := abmg.llmProvider.GenerateCompletion(ctx, prompt, systemPrompt)
	if err != nil {
		abmg.logger.WithContext(ctx).
			Error().
			Err(err).
			Str("location", req.LocationDescription).
			Str("map_type", req.MapType).
			Msg("Error generating AI battle map")
		return abmg.generateDefaultBattleMap(req), nil
	}

	var generatedMap GeneratedBattleMap
	if err := json.Unmarshal([]byte(response), &generatedMap); err != nil {
		abmg.logger.WithContext(ctx).
			Error().
			Err(err).
			Str("response_length", fmt.Sprintf("%d", len(response))).
			Msg("Error parsing AI battle map response")
		return abmg.generateDefaultBattleMap(req), nil
	}

	// Convert to database model
	return abmg.convertToBattleMap(generatedMap, req, gridX, gridY), nil
}

func (abmg *AIBattleMapGenerator) buildBattleMapPrompt(req models.GenerateBattleMapRequest, gridX, gridY int) string {
	return fmt.Sprintf(`Generate a tactical battle map based on this location:

Location Description: %s
Map Type: %s
Grid Size: %d x %d
Include Hazards: %v
Terrain Complexity: %s

Create a battle map with:
1. Terrain features that match the location description
2. Strategic cover positions and obstacles
3. Interesting tactical elements
4. Clear spawn points for both parties and enemies
5. Environmental hazards if requested
6. Tactical advice for running combat

Format the response as JSON:
{
  "terrain_features": [
    {
      "type": "wall/pillar/tree/water/elevation/etc",
      "position": {"x": 0, "y": 0},
      "size": {"width": 1, "height": 1},
      "properties": ["blocks_movement", "blocks_sight", "difficult_terrain", "provides_cover"]
    }
  ],
  "obstacle_positions": [
    {
      "type": "boulder/crate/rubble/furniture",
      "position": {"x": 0, "y": 0},
      "provides_cover": "half/three_quarters/full"
    }
  ],
  "cover_positions": [
    {
      "position": {"x": 0, "y": 0},
      "cover_type": "half/three_quarters/full",
      "direction": "north/south/east/west/all"
    }
  ],
  "hazard_zones": [
    {
      "type": "fire/acid/spike_pit/magical",
      "area": [{"x": 0, "y": 0}],
      "damage_type": "fire/acid/piercing/etc",
      "damage_dice": "2d6",
      "save_dc": 13,
      "save_type": "dexterity/constitution"
    }
  ],
  "spawn_points": {
    "party": [{"x": 0, "y": 0, "note": "Safe starting position"}],
    "enemies": [{"x": 0, "y": 0, "note": "Ambush position with cover"}]
  },
  "tactical_notes": [
    {
      "position": {"x": 0, "y": 0},
      "note": "High ground provides advantage on ranged attacks",
      "importance": "high"
    }
  ],
  "visual_theme": "dungeon_stone/forest_glade/urban_street/etc",
  "lighting_conditions": "bright/dim/darkness",
  "environmental_effects": ["fog", "rain", "wind"]
}`,
		req.LocationDescription,
		req.MapType,
		gridX, gridY,
		req.IncludeHazards,
		req.TerrainComplexity)
}

func (abmg *AIBattleMapGenerator) determineGridSize(desiredSize string) (int, int) {
	switch desiredSize {
	case "small":
		return 15, 15
	case "large":
		return 30, 30
	case "huge":
		return 40, 40
	default: // medium
		return 20, 20
	}
}

func (abmg *AIBattleMapGenerator) convertToBattleMap(generated GeneratedBattleMap, req models.GenerateBattleMapRequest, gridX, gridY int) *models.BattleMap {
	terrainJSON, _ := json.Marshal(generated.TerrainFeatures)
	obstaclesJSON, _ := json.Marshal(generated.ObstaclePositions)
	coverJSON, _ := json.Marshal(generated.CoverPositions)
	hazardsJSON, _ := json.Marshal(generated.HazardZones)
	spawnJSON, _ := json.Marshal(generated.SpawnPoints)
	notesJSON, _ := json.Marshal(generated.TacticalNotes)

	mapType := req.MapType
	if mapType == "" {
		mapType = abmg.inferMapType(req.LocationDescription)
	}

	return &models.BattleMap{
		ID:                  uuid.New(),
		LocationDescription: req.LocationDescription,
		MapType:             mapType,
		GridSizeX:           gridX,
		GridSizeY:           gridY,
		TerrainFeatures:     models.JSONB(terrainJSON),
		ObstaclePositions:   models.JSONB(obstaclesJSON),
		CoverPositions:      models.JSONB(coverJSON),
		HazardZones:         models.JSONB(hazardsJSON),
		SpawnPoints:         models.JSONB(spawnJSON),
		TacticalNotes:       models.JSONB(notesJSON),
		VisualTheme:         generated.VisualTheme,
	}
}

func (abmg *AIBattleMapGenerator) inferMapType(description string) string {
	desc := strings.ToLower(description)

	if strings.Contains(desc, "dungeon") || strings.Contains(desc, "cave") || strings.Contains(desc, "underground") {
		return "dungeon"
	} else if strings.Contains(desc, "forest") || strings.Contains(desc, "outdoor") || strings.Contains(desc, "field") {
		return "outdoor"
	} else if strings.Contains(desc, "city") || strings.Contains(desc, "street") || strings.Contains(desc, "tavern") {
		return "urban"
	}

	return "special"
}

func (abmg *AIBattleMapGenerator) generateDefaultBattleMap(req models.GenerateBattleMapRequest) *models.BattleMap {
	gridX, gridY := abmg.determineGridSize(req.DesiredSize)

	// Generate some basic terrain features
	var terrainFeatures []models.BattleMapTerrainFeature

	// Add some walls or trees based on map type
	switch req.MapType {
	case "dungeon":
		// Add walls around the edges
		for x := 0; x < gridX; x++ {
			terrainFeatures = append(terrainFeatures,
				models.BattleMapTerrainFeature{
					Type:       "wall",
					Position:   models.Position{X: x, Y: 0},
					Size:       models.Size{Width: 1, Height: 1},
					Properties: []string{"blocks_movement", "blocks_sight"},
				},
				models.BattleMapTerrainFeature{
					Type:       "wall",
					Position:   models.Position{X: x, Y: gridY - 1},
					Size:       models.Size{Width: 1, Height: 1},
					Properties: []string{"blocks_movement", "blocks_sight"},
				},
			)
		}
	case "outdoor":
		// Add some trees
		for i := 0; i < 5; i++ {
			terrainFeatures = append(terrainFeatures, models.BattleMapTerrainFeature{
				Type:       "tree",
				Position:   models.Position{X: rand.Intn(gridX), Y: rand.Intn(gridY)},
				Size:       models.Size{Width: 1, Height: 1},
				Properties: []string{"blocks_movement", "provides_cover"},
			})
		}
	}

	// Add some cover positions
	var coverPositions []map[string]interface{}
	for i := 0; i < 3; i++ {
		coverPositions = append(coverPositions, map[string]interface{}{
			"position":   models.Position{X: rand.Intn(gridX), Y: rand.Intn(gridY)},
			"cover_type": "half",
			"direction":  "all",
		})
	}

	// Basic spawn points
	spawnPoints := map[string][]map[string]interface{}{
		"party": {
			{
				"x":    2,
				"y":    gridY / 2,
				"note": "Starting position",
			},
		},
		"enemies": {
			{
				"x":    gridX - 3,
				"y":    gridY / 2,
				"note": "Enemy starting position",
			},
		},
	}

	terrainJSON, _ := json.Marshal(terrainFeatures)
	coverJSON, _ := json.Marshal(coverPositions)
	spawnJSON, _ := json.Marshal(spawnPoints)

	return &models.BattleMap{
		ID:                  uuid.New(),
		LocationDescription: req.LocationDescription,
		MapType:             req.MapType,
		GridSizeX:           gridX,
		GridSizeY:           gridY,
		TerrainFeatures:     models.JSONB(terrainJSON),
		ObstaclePositions:   models.JSONB(`[]`),
		CoverPositions:      models.JSONB(coverJSON),
		HazardZones:         models.JSONB(`[]`),
		SpawnPoints:         models.JSONB(spawnJSON),
		TacticalNotes:       models.JSONB(`[]`),
		VisualTheme:         "default",
	}
}

// GeneratedBattleMap represents the AI-generated battle map structure
type GeneratedBattleMap struct {
	TerrainFeatures      []models.BattleMapTerrainFeature `json:"terrain_features"`
	ObstaclePositions    []ObstaclePosition               `json:"obstacle_positions"`
	CoverPositions       []CoverPosition                  `json:"cover_positions"`
	HazardZones          []models.HazardZone              `json:"hazard_zones"`
	SpawnPoints          map[string][]SpawnPoint          `json:"spawn_points"`
	TacticalNotes        []models.TacticalNote            `json:"tactical_notes"`
	VisualTheme          string                           `json:"visual_theme"`
	LightingConditions   string                           `json:"lighting_conditions"`
	EnvironmentalEffects []string                         `json:"environmental_effects"`
}

type ObstaclePosition struct {
	Type          string          `json:"type"`
	Position      models.Position `json:"position"`
	ProvidesCover string          `json:"provides_cover"`
}

type CoverPosition struct {
	Position  models.Position `json:"position"`
	CoverType string          `json:"cover_type"`
	Direction string          `json:"direction"`
}

type SpawnPoint struct {
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Note string `json:"note"`
}
