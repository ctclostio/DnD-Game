#!/bin/bash

# Fix fighter lookups
sed -i 's/fighter := combat\.Combatants\["char-1"\]/\/\/ Find fighter in combatants slice\n\t\t\t\t\tvar fighter *models.Combatant\n\t\t\t\t\tfor i := range combat.Combatants {\n\t\t\t\t\t\tif combat.Combatants[i].ID == "char-1" {\n\t\t\t\t\t\t\tfighter = \&combat.Combatants[i]\n\t\t\t\t\t\t\tbreak\n\t\t\t\t\t\t}\n\t\t\t\t\t}\n\t\t\t\t\trequire.NotNil(t, fighter)/g' internal/services/combat_test.go

# Fix tiefling lookups
sed -i 's/tiefling := combat\.Combatants\["char-1"\]/\/\/ Find tiefling in combatants slice\n\t\t\t\t\tvar tiefling *models.Combatant\n\t\t\t\t\tfor i := range combat.Combatants {\n\t\t\t\t\t\tif combat.Combatants[i].ID == "char-1" {\n\t\t\t\t\t\t\ttiefling = \&combat.Combatants[i]\n\t\t\t\t\t\t\tbreak\n\t\t\t\t\t\t}\n\t\t\t\t\t}\n\t\t\t\t\trequire.NotNil(t, tiefling)/g' internal/services/combat_test.go

# Fix construct lookups
sed -i 's/construct := combat\.Combatants\["char-1"\]/\/\/ Find construct in combatants slice\n\t\t\t\t\tvar construct *models.Combatant\n\t\t\t\t\tfor i := range combat.Combatants {\n\t\t\t\t\t\tif combat.Combatants[i].ID == "char-1" {\n\t\t\t\t\t\t\tconstruct = \&combat.Combatants[i]\n\t\t\t\t\t\t\tbreak\n\t\t\t\t\t\t}\n\t\t\t\t\t}\n\t\t\t\t\trequire.NotNil(t, construct)/g' internal/services/combat_test.go

# Fix shadow lookups
sed -i 's/shadow := combat\.Combatants\["char-1"\]/\/\/ Find shadow in combatants slice\n\t\t\t\t\tvar shadow *models.Combatant\n\t\t\t\t\tfor i := range combat.Combatants {\n\t\t\t\t\t\tif combat.Combatants[i].ID == "char-1" {\n\t\t\t\t\t\t\tshadow = \&combat.Combatants[i]\n\t\t\t\t\t\t\tbreak\n\t\t\t\t\t\t}\n\t\t\t\t\t}\n\t\t\t\t\trequire.NotNil(t, shadow)/g' internal/services/combat_test.go

# Add missing newline at end of file
echo "" >> internal/services/combat_test.go