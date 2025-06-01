import React, { useState } from 'react';

const ConnectionLayer = ({
  connections,
  nodes,
  getPortPosition,
  onConnectionDelete,
  draggedConnection,
  zoom,
  offset
}) => {
  const [selectedConnection, setSelectedConnection] = useState(null);

  // Calculate bezier curve path between two points
  const calculatePath = (start, end) => {
    const controlPointOffset = Math.abs(end.x - start.x) / 2;
    const d = `M ${start.x} ${start.y} C ${start.x + controlPointOffset} ${start.y}, ${end.x - controlPointOffset} ${end.y}, ${end.x} ${end.y}`;
    return d;
  };

  // Get connection endpoints
  const getConnectionEndpoints = (connection) => {
    const start = getPortPosition(connection.from_node_id, connection.from_port_id, 'output');
    const end = getPortPosition(connection.to_node_id, connection.to_port_id, 'input');
    return { start, end };
  };

  const handleConnectionClick = (e, connectionId) => {
    e.stopPropagation();
    setSelectedConnection(connectionId);
  };

  const handleDeleteKey = (e) => {
    if (e.key === 'Delete' && selectedConnection) {
      onConnectionDelete(selectedConnection);
      setSelectedConnection(null);
    }
  };

  React.useEffect(() => {
    document.addEventListener('keydown', handleDeleteKey);
    return () => {
      document.removeEventListener('keydown', handleDeleteKey);
    };
  }, [selectedConnection]);

  return (
    <svg 
      className="connections-layer"
      style={{
        position: 'absolute',
        top: 0,
        left: 0,
        width: '100%',
        height: '100%',
        pointerEvents: 'none'
      }}
    >
      <defs>
        {/* Arrow marker for connection endpoints */}
        <marker
          id="arrowhead"
          markerWidth="10"
          markerHeight="10"
          refX="9"
          refY="3"
          orient="auto"
        >
          <polygon
            points="0 0, 10 3, 0 6"
            fill="#3498db"
          />
        </marker>
      </defs>

      {/* Render existing connections */}
      {connections.map(connection => {
        const { start, end } = getConnectionEndpoints(connection);
        if (!start || !end) return null;

        return (
          <g key={connection.id}>
            {/* Invisible wider path for easier clicking */}
            <path
              d={calculatePath(start, end)}
              stroke="transparent"
              strokeWidth="10"
              fill="none"
              style={{ cursor: 'pointer', pointerEvents: 'stroke' }}
              onClick={(e) => handleConnectionClick(e, connection.id)}
            />
            {/* Visible connection line */}
            <path
              d={calculatePath(start, end)}
              className={`connection-line ${selectedConnection === connection.id ? 'selected' : ''}`}
              markerEnd="url(#arrowhead)"
              style={{ pointerEvents: 'none' }}
            />
          </g>
        );
      })}

      {/* Render dragged connection */}
      {draggedConnection && (
        <path
          d={calculatePath(
            draggedConnection.startPosition,
            draggedConnection.endPosition
          )}
          className="connection-line dragging"
          stroke="#3498db"
          strokeWidth="2"
          strokeDasharray="5,5"
          fill="none"
          markerEnd="url(#arrowhead)"
          style={{ pointerEvents: 'none' }}
        />
      )}

      {/* Delete button for selected connection */}
      {selectedConnection && (() => {
        const connection = connections.find(c => c.id === selectedConnection);
        if (!connection) return null;
        
        const { start, end } = getConnectionEndpoints(connection);
        if (!start || !end) return null;
        
        const midX = (start.x + end.x) / 2;
        const midY = (start.y + end.y) / 2;
        
        return (
          <g style={{ pointerEvents: 'all' }}>
            <circle
              cx={midX}
              cy={midY}
              r="15"
              fill="#e74c3c"
              onClick={(e) => {
                e.stopPropagation();
                onConnectionDelete(selectedConnection);
                setSelectedConnection(null);
              }}
              style={{ cursor: 'pointer' }}
            />
            <text
              x={midX}
              y={midY + 5}
              textAnchor="middle"
              fill="white"
              fontSize="14"
              fontWeight="bold"
              style={{ pointerEvents: 'none' }}
            >
              ×
            </text>
          </g>
        );
      })()}
    </svg>
  );
};

export default ConnectionLayer;