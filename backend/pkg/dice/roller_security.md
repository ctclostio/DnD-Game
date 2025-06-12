# Security Note: Dice Roller

The dice roller uses math/rand for game mechanics randomness.
This is intentional as:

1. It's used for game dice rolls, not security
2. Predictable random is acceptable for game mechanics
3. crypto/rand would be overkill and slower for dice rolls

If you need secure random for security purposes (tokens, passwords, etc),
use crypto/rand instead.
