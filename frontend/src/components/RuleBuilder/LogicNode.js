import React, { useState, useRef } from 'react';
import { useDrag } from 'react-dnd';
import { FaTrash, FaStar, FaGripVertical } from 'react-icons/fa';
import { getClickableProps, getSelectableProps } from '../../utils/accessibility';

const LogicNode = ({
  node,
  isSelected,
  isStartNode,
  onSelect,
  onUpdate,
  onDelete,
  onSetAsStart,
  onConnectionStart,
  onConnectionEnd,
  onPortHover
}) => {
  const nodeRef = useRef(null);
  const [isDraggingNode, setIsDraggingNode] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });

  // Node dragging
  const handleNodeDragStart = (e) => {
    e.stopPropagation();
    setIsDraggingNode(true);
    setDragStart({
      x: e.clientX - node.position.x,
      y: e.clientY - node.position.y
    });
    onSelect();
  };

  const handleNodeDrag = (e) => {
    if (isDraggingNode) {
      const newX = e.clientX - dragStart.x;
      const newY = e.clientY - dragStart.y;
      onUpdate({ position: { x: newX, y: newY } });
    }
  };

  const handleNodeDragEnd = () => {
    setIsDraggingNode(false);
  };

  React.useEffect(() => {
    const handleMouseMove = (e) => {
      handleNodeDrag(e);
    };

    const handleMouseUp = () => {
      handleNodeDragEnd();
    };

    if (isDraggingNode) {
      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);
      
      return () => {
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
      };
    }
  }, [isDraggingNode, dragStart]);

  // Port connection handling
  const handlePortMouseDown = (e, portId, portType) => {
    e.stopPropagation();
    const rect = e.target.getBoundingClientRect();
    const position = {
      x: rect.left + rect.width / 2,
      y: rect.top + rect.height / 2
    };
    onConnectionStart(node.id, portId, portType, position);
  };

  const handlePortMouseUp = (e, portId, portType) => {
    e.stopPropagation();
    onConnectionEnd(node.id, portId, portType);
  };

  const handlePortMouseEnter = (portId, portType) => {
    onPortHover({ nodeId: node.id, portId, portType });
  };

  const handlePortMouseLeave = () => {
    onPortHover(null);
  };

  // Get node color based on type
  const getNodeColor = () => {
    const typeColors = {
      trigger: '#ff6b6b',
      condition: '#f7b731',
      action: '#ee5a24',
      calc: '#0fb9b1',
      flow: '#a55eea'
    };
    
    const typePrefix = node.type.split('_')[0];
    return typeColors[typePrefix] || '#95a5a6';
  };

  // Get node icon based on type
  const getNodeIcon = () => {
    const icons = {
      trigger_action: '⚡',
      trigger_damage: '💔',
      trigger_time: '⏰',
      condition_check: '❓',
      condition_compare: '⚖️',
      condition_roll: '🎲',
      action_damage: '⚔️',
      action_heal: '❤️',
      action_effect: '✨',
      action_resource: '💎',
      calc_math: '🧮',
      calc_random: '🎲',
      flow_split: '🔀',
      flow_merge: '🔃'
    };
    
    return icons[node.type] || '📦';
  };

  const nodeColor = getNodeColor();

  return (
    <div
      ref={nodeRef}
      id={`node-${node.id}`}
      className={`logic-node ${isSelected ? 'selected' : ''} ${isStartNode ? 'start-node' : ''}`}
      style={{
        left: node.position.x,
        top: node.position.y,
        borderColor: nodeColor
      }}
      {...getSelectableProps((e) => {
        e.stopPropagation();
        onSelect();
      }, isSelected)}
    >
      <div 
        className="node-header" 
        style={{ backgroundColor: nodeColor + '20' }}
        onMouseDown={handleNodeDragStart}
      >
        <FaGripVertical className="node-drag-handle" style={{ color: nodeColor }} />
        <span className="node-icon">{getNodeIcon()}</span>
        <span className="node-title">{node.subtype || node.type}</span>
        <div className="node-actions">
          {!isStartNode && (
            <button 
              className="node-action"
              onClick={(e) => {
                e.stopPropagation();
                onSetAsStart();
              }}
              title="Set as start node"
            >
              <FaStar />
            </button>
          )}
          <button 
            className="node-action delete"
            onClick={(e) => {
              e.stopPropagation();
              onDelete();
            }}
            title="Delete node"
          >
            <FaTrash />
          </button>
        </div>
      </div>

      <div className="node-body">
        {/* Display key properties */}
        {Object.entries(node.properties || {}).slice(0, 3).map(([key, value]) => (
          <div key={key} className="node-property">
            <div className="node-property-label">{key}:</div>
            <div className="node-property-value">{String(value)}</div>
          </div>
        ))}
      </div>

      {/* Input ports */}
      {node.inputs && node.inputs.length > 0 && (
        <div className="input-ports">
          {node.inputs.map(port => (
            <div
              key={port.id}
              className={`node-port input-port ${port.connected ? 'connected' : ''}`}
              data-port-id={port.id}
              onMouseDown={(e) => handlePortMouseDown(e, port.id, 'input')}
              onMouseUp={(e) => handlePortMouseUp(e, port.id, 'input')}
              onMouseEnter={() => handlePortMouseEnter(port.id, 'input')}
              onMouseLeave={handlePortMouseLeave}
              title={port.name}
            />
          ))}
        </div>
      )}

      {/* Output ports */}
      {node.outputs && node.outputs.length > 0 && (
        <div className="output-ports">
          {node.outputs.map(port => (
            <div
              key={port.id}
              className={`node-port output-port ${port.connected ? 'connected' : ''}`}
              data-port-id={port.id}
              onMouseDown={(e) => handlePortMouseDown(e, port.id, 'output')}
              onMouseUp={(e) => handlePortMouseUp(e, port.id, 'output')}
              onMouseEnter={() => handlePortMouseEnter(port.id, 'output')}
              onMouseLeave={handlePortMouseLeave}
              title={port.name}
            />
          ))}
        </div>
      )}
    </div>
  );
};

export default LogicNode;