#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const glob = require('glob');

// Pattern to find divs with onClick but no onKeyDown
const ONCLICK_PATTERN = /<div[^>]*onClick\s*=\s*{[^}]+}[^>]*>/g;
const HAS_ONKEYDOWN_PATTERN = /onKeyDown/;

// Components reported by SonarCloud with accessibility issues
const componentsWithIssues = [
  'NPCAutonomy.js',
  'GeneratedContent.js', 
  'EconomicDashboard.js',
  'CultureExplorer.js',
  'FactionPersonalities.js',
  'WorldEventsFeed.js',
  'SettlementGenerator.js',
  'FactionCreator.js',
  'WorldTimeline.js',
  'WorldEventViewer.js',
  'StoryThreads.js',
  'StoryElements.js',
  'SettlementManager.js',
  'PerspectiveViewer.js',
  'NPCDialogue.js',
  'NodePalette.js',
  'NarrativeEngine.js',
  'LogicNode.js',
  'LocationGenerator.js',
  'FactionManager.js',
  'CombatAutomation.js',
  'BattleMapViewer.js'
];

// Find all files that need fixing
const filesToFix = [];
componentsWithIssues.forEach(component => {
  const files = glob.sync(`frontend/src/**/${component}`);
  filesToFix.push(...files);
});

console.log(`Found ${filesToFix.length} files to check for accessibility issues`);

let totalFixed = 0;

filesToFix.forEach(filePath => {
  let content = fs.readFileSync(filePath, 'utf8');
  const originalContent = content;
  
  // Check if file already imports accessibility utils
  const hasAccessibilityImport = content.includes('accessibility');
  
  // Add import if needed
  if (!hasAccessibilityImport && content.includes('onClick')) {
    const importRegex = /^import[^;]+;$/gm;
    let lastImport = null;
    let match;
    
    while ((match = importRegex.exec(content)) !== null) {
      lastImport = match;
    }
    
    if (lastImport) {
      const insertPosition = lastImport.index + lastImport[0].length;
      const relativePathDepth = filePath.split('/').length - 4; // Adjust based on src location
      const importPath = '../'.repeat(relativePathDepth) + 'utils/accessibility';
      content = content.slice(0, insertPosition) + 
                `\nimport { getClickableProps, getSelectableProps } from '${importPath}';` +
                content.slice(insertPosition);
    }
  }
  
  // Find and count clickable divs without keyboard support
  const matches = content.match(ONCLICK_PATTERN) || [];
  let fixedCount = 0;
  
  matches.forEach(match => {
    if (!HAS_ONKEYDOWN_PATTERN.test(match)) {
      console.log(`  Found accessibility issue in ${path.basename(filePath)}`);
      fixedCount++;
    }
  });
  
  if (fixedCount > 0) {
    totalFixed += fixedCount;
    console.log(`  Will fix ${fixedCount} issues in ${path.basename(filePath)}`);
  }
});

console.log(`\nTotal accessibility issues found: ${totalFixed}`);
console.log('\nNote: This is a dry run. Actual fixes would need to be applied manually or with more sophisticated AST parsing.');