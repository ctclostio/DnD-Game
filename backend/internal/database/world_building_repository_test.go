package database

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

func TestUpdateSettlementProsperity(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewWorldBuildingRepository(db)
	
	// Create a test settlement
	settlement := &models.Settlement{
		GameSessionID:   uuid.New(),
		Name:           "Test Village",
		Type:           models.SettlementVillage,
		Population:     100,
		WealthLevel:    5,
		Region:         "Test Region",
		DangerLevel:    2,
		CorruptionLevel: 1,
		AgeCategory:    models.SettlementAgeAncient,
		TerrainType:    models.TerrainTypeForest,
		Climate:        models.ClimateTemperate,
		GovernmentType: models.GovernmentMonarchy,
		Alignment:      models.AlignmentLawfulGood,
	}
	
	// Create the settlement
	err := repo.CreateSettlement(settlement)
	require.NoError(t, err)
	assert.NotEmpty(t, settlement.ID)
	
	// Update prosperity
	newWealthLevel := 8
	err = repo.UpdateSettlementProsperity(settlement.ID, newWealthLevel)
	require.NoError(t, err)
	
	// Retrieve and verify
	updated, err := repo.GetSettlement(settlement.ID)
	require.NoError(t, err)
	assert.Equal(t, newWealthLevel, updated.WealthLevel)
	assert.True(t, updated.UpdatedAt.After(settlement.UpdatedAt))
}

func TestUpdateSettlement(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewWorldBuildingRepository(db)
	
	// Create a test settlement
	settlement := &models.Settlement{
		GameSessionID:   uuid.New(),
		Name:           "Test City",
		Type:           models.SettlementCity,
		Population:     1000,
		WealthLevel:    7,
		Region:         "Test Region",
		DangerLevel:    3,
		CorruptionLevel: 2,
		Description:    "A bustling test city",
		AgeCategory:    models.SettlementAgeAncient,
		TerrainType:    models.TerrainTypePlains,
		Climate:        models.ClimateTemperate,
		GovernmentType: models.GovernmentRepublic,
		Alignment:      models.AlignmentLawfulNeutral,
	}
	
	// Create the settlement
	err := repo.CreateSettlement(settlement)
	require.NoError(t, err)
	
	// Update multiple fields
	settlement.Name = "Updated City"
	settlement.Population = 1500
	settlement.WealthLevel = 9
	settlement.Description = "An even more bustling city"
	settlement.DangerLevel = 2
	
	err = repo.UpdateSettlement(settlement)
	require.NoError(t, err)
	
	// Retrieve and verify
	updated, err := repo.GetSettlement(settlement.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated City", updated.Name)
	assert.Equal(t, 1500, updated.Population)
	assert.Equal(t, 9, updated.WealthLevel)
	assert.Equal(t, "An even more bustling city", updated.Description)
	assert.Equal(t, 2, updated.DangerLevel)
	assert.True(t, updated.UpdatedAt.After(settlement.CreatedAt))
}