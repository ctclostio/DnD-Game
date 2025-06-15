package services

import "github.com/ctclostio/DnD-Game/backend/internal/models"

// AttributeProvider interface for types that provide attributes
type AttributeProvider interface {
	GetAttributes() models.Attributes
}

// CharacterAttributeProvider wraps Character to implement AttributeProvider
type CharacterAttributeProvider struct {
	*models.Character
}

func (c CharacterAttributeProvider) GetAttributes() models.Attributes {
	return c.Attributes
}

// NPCAttributeProvider wraps NPC to implement AttributeProvider
type NPCAttributeProvider struct {
	*models.NPC
}

func (n NPCAttributeProvider) GetAttributes() models.Attributes {
	return n.Attributes
}

// CalculateAbilityModifier computes the ability modifier from a score
func CalculateAbilityModifier(score int) int {
	return (score - 10) / 2
}

// CalculateSavingThrows generates saving throws for any entity with attributes
func CalculateSavingThrows(provider AttributeProvider) models.SavingThrows {
	attrs := provider.GetAttributes()
	
	return models.SavingThrows{
		Strength: models.SavingThrow{
			Modifier:    CalculateAbilityModifier(attrs.Strength),
			Proficiency: false,
		},
		Dexterity: models.SavingThrow{
			Modifier:    CalculateAbilityModifier(attrs.Dexterity),
			Proficiency: false,
		},
		Constitution: models.SavingThrow{
			Modifier:    CalculateAbilityModifier(attrs.Constitution),
			Proficiency: false,
		},
		Intelligence: models.SavingThrow{
			Modifier:    CalculateAbilityModifier(attrs.Intelligence),
			Proficiency: false,
		},
		Wisdom: models.SavingThrow{
			Modifier:    CalculateAbilityModifier(attrs.Wisdom),
			Proficiency: false,
		},
		Charisma: models.SavingThrow{
			Modifier:    CalculateAbilityModifier(attrs.Charisma),
			Proficiency: false,
		},
	}
}