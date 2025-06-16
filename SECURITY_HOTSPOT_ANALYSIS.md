# SonarCloud Security Hotspot Analysis

## Summary
Found 102 security hotspots in the project:
- **3 HIGH priority** - Write permissions on copied Docker resources
- **2 MEDIUM priority** - Regex vulnerability to ReDoS attacks
- **97 MEDIUM priority** - Pseudorandom number generator usage

## HIGH Priority Issues

### 1. Docker File Permissions (3 issues)
**Risk**: Copied resources have write permissions when they should be read-only in production.

#### Affected Files:
1. `backend/Dockerfile:40` - Data directory copied with write permissions
2. `frontend/Dockerfile:46` - Static build files copied with write permissions  
3. `frontend/Dockerfile:49` - Nginx config copied with write permissions

**Security Impact**: 
- Allows potential runtime modification of application files
- Could enable privilege escalation or code injection
- Violates principle of least privilege

**Fix Strategy**:
Set read-only permissions (chmod) after copying files to ensure immutability in production.

## MEDIUM Priority Issues

### 2. ReDoS Vulnerability (2 issues)
**Risk**: Regex patterns vulnerable to catastrophic backtracking causing denial of service.

#### Affected Files:
1. `e2e/pages/CombatPage.ts:180`
2. `e2e/tests/combat-encounter.spec.ts:465`

**Security Impact**:
- Can cause exponential time complexity with crafted input
- Potential DoS attack vector
- Only affects test files (lower risk)

### 3. Pseudorandom Number Generator (97 issues)
**Risk**: Using `math/rand` instead of `crypto/rand` for potentially security-sensitive operations.

#### Affected Services:
- Game mechanics (dice rolls, combat) - 40 issues
- World generation systems - 57 issues

**Security Impact**:
- Predictable randomness in game mechanics
- Not cryptographically secure
- Acceptable for game logic, problematic for security features

## Recommendations

1. **Immediate Action**: Fix HIGH priority Docker permission issues
2. **Review**: Assess if any random number usage is security-critical
3. **Consider**: Replacing ReDoS-vulnerable regex patterns in tests
4. **Document**: Which random operations need crypto-secure randomness