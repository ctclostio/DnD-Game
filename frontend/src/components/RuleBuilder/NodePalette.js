import React, { useState } from 'react';
import { useDrag } from 'react-dnd';
import { FaSearch, FaChevronDown, FaChevronRight } from 'react-icons/fa';

const NodeTemplateItem = ({ template, onNodeAdd }) => {
  const [{ isDragging }, drag] = useDrag({
    type: 'node-template',
    item: {
      ...template,
      onDrop: (position) => {
        const newNode = {
          type: template.node_type,
          subtype: template.subtype,
          position,
          properties: { ...template.default_properties },
          inputs: template.input_ports,
          outputs: template.output_ports
        };
        onNodeAdd(template);
      }
    },
    collect: (monitor) => ({
      isDragging: monitor.isDragging()
    })
  });

  return (
    <div
      ref={drag}
      className={`node-template ${isDragging ? 'dragging' : ''}`}
      style={{
        opacity: isDragging ? 0.5 : 1,
        borderColor: template.color || '#ddd'
      }}
    >
      <span className="node-template-icon" style={{ color: template.color }}>
        {getNodeIcon(template.icon)}
      </span>
      <div className="node-template-info">
        <div className="node-template-name">{template.name}</div>
        <div className="node-template-desc">{template.description}</div>
      </div>
    </div>
  );
};

const NodePalette = ({ nodeTemplates, onNodeAdd }) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [collapsedCategories, setCollapsedCategories] = useState({});

  // Group templates by category
  const groupedTemplates = nodeTemplates.reduce((acc, template) => {
    const category = template.category || 'uncategorized';
    if (!acc[category]) {
      acc[category] = [];
    }
    acc[category].push(template);
    return acc;
  }, {});

  // Filter templates by search term
  const filterTemplates = (templates) => {
    if (!searchTerm) return templates;
    
    return templates.filter(template => 
      template.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      template.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
      template.node_type.toLowerCase().includes(searchTerm.toLowerCase())
    );
  };

  const toggleCategory = (category) => {
    setCollapsedCategories(prev => ({
      ...prev,
      [category]: !prev[category]
    }));
  };

  const getCategoryInfo = (category) => {
    const categoryData = {
      triggers: { name: 'Triggers', icon: '⚡', description: 'Start rule execution' },
      conditions: { name: 'Conditions', icon: '❓', description: 'Make decisions' },
      actions: { name: 'Actions', icon: '⚔️', description: 'Perform effects' },
      calculations: { name: 'Calculations', icon: '🧮', description: 'Process data' },
      flow: { name: 'Flow Control', icon: '🔀', description: 'Control execution' }
    };
    
    return categoryData[category] || { 
      name: category.charAt(0).toUpperCase() + category.slice(1), 
      icon: '📦',
      description: '' 
    };
  };

  return (
    <div className="node-palette">
      <h3>Node Library</h3>
      
      {/* Search bar */}
      <div className="node-search">
        <FaSearch />
        <input
          type="text"
          placeholder="Search nodes..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
        />
      </div>

      {/* Node categories */}
      {Object.entries(groupedTemplates).map(([category, templates]) => {
        const filteredTemplates = filterTemplates(templates);
        if (filteredTemplates.length === 0 && searchTerm) return null;
        
        const categoryInfo = getCategoryInfo(category);
        const isCollapsed = collapsedCategories[category];

        return (
          <div key={category} className="node-category">
            <div 
              className="category-header"
              onClick={() => toggleCategory(category)}
            >
              <span className="category-toggle">
                {isCollapsed ? <FaChevronRight /> : <FaChevronDown />}
              </span>
              <span className="category-icon">{categoryInfo.icon}</span>
              <h4>{categoryInfo.name}</h4>
              <span className="category-count">({filteredTemplates.length})</span>
            </div>
            
            {!isCollapsed && (
              <>
                {categoryInfo.description && (
                  <p className="category-description">{categoryInfo.description}</p>
                )}
                <div className="node-templates">
                  {filteredTemplates.map((template, index) => (
                    <NodeTemplateItem
                      key={`${template.node_type}-${index}`}
                      template={template}
                      onNodeAdd={onNodeAdd}
                    />
                  ))}
                </div>
              </>
            )}
          </div>
        );
      })}

      {/* Empty state */}
      {Object.keys(groupedTemplates).length === 0 && (
        <div className="empty-palette">
          <p>No node templates available</p>
        </div>
      )}

      {/* Instructions */}
      <div className="palette-instructions">
        <p>Drag nodes to the canvas to build your rule</p>
      </div>
    </div>
  );
};

// Helper function to get icon based on icon name
const getNodeIcon = (iconName) => {
  const icons = {
    bolt: '⚡',
    'heart-broken': '💔',
    clock: '⏰',
    'code-branch': '🔀',
    'balance-scale': '⚖️',
    'dice-d20': '🎲',
    sword: '⚔️',
    magic: '✨',
    database: '💾',
    calculator: '🧮',
    dice: '🎲',
    tag: '🏷️'
  };
  
  return icons[iconName] || '📦';
};

export default NodePalette;