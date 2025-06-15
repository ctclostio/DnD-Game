import { createSelector } from 'reselect';
import { RootState } from '../index';
import { CombatParticipant } from '../../types/game';

// Base selectors
const selectCombatState = (state: RootState) => state.combat;
const selectParticipants = (state: RootState) => state.combat.participants;
const selectCurrentParticipantId = (state: RootState) => state.combat.currentParticipantId;
const selectInitiativeOrder = (state: RootState) => state.combat.initiativeOrder;

// Get all participants as array
export const selectParticipantsArray = createSelector(
  [selectParticipants],
  (participants) => participants.ids.map((id: string) => participants.entities[id])
);

// Get current participant
export const selectCurrentParticipant = createSelector(
  [selectParticipants, selectCurrentParticipantId],
  (participants, currentId) => currentId ? participants.entities[currentId] : null
);

// Get participants sorted by initiative
export const selectParticipantsByInitiative = createSelector(
  [selectParticipants, selectInitiativeOrder],
  (participants, order) => order.map((id: string) => participants.entities[id])
);

// Get active participants (not unconscious)
export const selectActiveParticipants = createSelector(
  [selectParticipantsArray],
  (participants) => participants.filter((p: CombatParticipant) => !p.conditions.includes('unconscious'))
);

// Get participants by type
export const selectPlayerParticipants = createSelector(
  [selectParticipantsArray],
  (participants) => participants.filter((p: CombatParticipant) => p.isPlayer)
);

export const selectNPCParticipants = createSelector(
  [selectParticipantsArray],
  (participants) => participants.filter((p: CombatParticipant) => !p.isPlayer)
);

// Combat status selectors
export const selectCombatRound = (state: RootState) => state.combat.round;
export const selectCombatTurn = (state: RootState) => state.combat.turn;
export const selectIsCombatActive = (state: RootState) => state.combat.active;

// Get participants with conditions
export const selectParticipantsWithConditions = createSelector(
  [selectParticipantsArray],
  (participants) => participants.filter((p: CombatParticipant) => p.conditions.length > 0)
);

// Get concentrating participants
export const selectConcentratingParticipants = createSelector(
  [selectParticipantsArray],
  (participants) => participants.filter((p: CombatParticipant) => p.concentrating)
);

// Combat statistics
export const selectCombatStatistics = createSelector(
  [selectParticipantsArray, selectCombatRound],
  (participants, round) => {
    const totalParticipants = participants.length;
    const activeParticipants = participants.filter((p: CombatParticipant) => !p.conditions.includes('unconscious')).length;
    const downedParticipants = totalParticipants - activeParticipants;
    
    const totalDamageDealt = participants.reduce((sum: number, p: CombatParticipant) => 
      sum + (p.hitPointsMax - p.hitPointsCurrent), 0
    );
    
    return {
      round,
      totalParticipants,
      activeParticipants,
      downedParticipants,
      totalDamageDealt,
    };
  }
);

// Get next participant in initiative
export const selectNextParticipant = createSelector(
  [selectParticipants, selectInitiativeOrder, selectCurrentParticipantId],
  (participants, order, currentId) => {
    if (!currentId || order.length === 0) return null;
    
    const currentIndex = order.indexOf(currentId);
    const nextIndex = (currentIndex + 1) % order.length;
    return participants.entities[order[nextIndex]];
  }
);

// Check if current participant can act
export const selectCanCurrentParticipantAct = createSelector(
  [selectCurrentParticipant],
  (participant) => {
    if (!participant) return false;
    return !participant.hasActed || !participant.hasBonusActed || participant.movementUsed < participant.movementMax;
  }
);
