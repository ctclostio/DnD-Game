import React, { useState, useRef, useEffect } from 'react';
import { useDrop } from 'react-dnd';
import LogicNode from './LogicNode';
import ConnectionLayer from './ConnectionLayer';

const VisualLogicEditor = ({
  logicGraph,
  selectedNode,
  onNodeSelect,
  onNodeUpdate,
  onNodeDelete,
  onConnectionAdd,
  onConnectionDelete,
  onStartNodeSet
}) => {
  const canvasRef = useRef(null);
  const [canvasOffset, setCanvasOffset] = useState({ x: 0, y: 0 });
  const [zoom, setZoom] = useState(1);
  const [isPanning, setIsPanning] = useState(false);
  const [panStart, setPanStart] = useState({ x: 0, y: 0 });
  const [draggedConnection, setDraggedConnection] = useState(null);
  const [hoveredPort, setHoveredPort] = useState(null);

  // Drop target for new nodes from palette
  const [{ isOver }, drop] = useDrop({
    accept: 'node-template',
    drop: (item, monitor) => {
      const offset = monitor.getClientOffset();
      const canvasRect = canvasRef.current.getBoundingClientRect();
      
      // Calculate position relative to canvas with zoom and pan
      const x = (offset.x - canvasRect.left - canvasOffset.x) / zoom;
      const y = (offset.y - canvasRect.top - canvasOffset.y) / zoom;
      
      // This will trigger the parent's handleNodeAdd
      if (item.onDrop) {
        item.onDrop({ x, y });
      }
    },
    collect: (monitor) => ({
      isOver: monitor.isOver()
    })
  });

  // Combine refs
  const setRefs = (el) => {
    canvasRef.current = el;
    drop(el);
  };

  // Handle canvas panning
  useEffect(() => {
    const handleMouseMove = (e) => {
      if (isPanning) {
        const dx = e.clientX - panStart.x;
        const dy = e.clientY - panStart.y;
        setCanvasOffset({
          x: canvasOffset.x + dx,
          y: canvasOffset.y + dy
        });
        setPanStart({ x: e.clientX, y: e.clientY });
      }
    };

    const handleMouseUp = () => {
      setIsPanning(false);
    };

    if (isPanning) {
      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);
      
      return () => {
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
      };
    }
  }, [isPanning, panStart, canvasOffset]);

  const handleCanvasMouseDown = (e) => {
    // Only pan if clicking on empty canvas
    if (e.target === canvasRef.current) {
      setIsPanning(true);
      setPanStart({ x: e.clientX, y: e.clientY });
      onNodeSelect(null); // Deselect nodes
    }
  };

  const handleWheel = (e) => {
    e.preventDefault();
    const delta = e.deltaY * -0.001;
    const newZoom = Math.min(Math.max(0.5, zoom + delta), 2);
    setZoom(newZoom);
  };

  // Handle connection dragging
  const handleConnectionStart = (nodeId, portId, portType, position) => {
    setDraggedConnection({
      fromNodeId: nodeId,
      fromPortId: portId,
      fromPortType: portType,
      startPosition: position,
      endPosition: position
    });
  };

  const handleConnectionDrag = (e) => {
    if (draggedConnection) {
      const canvasRect = canvasRef.current.getBoundingClientRect();
      const x = (e.clientX - canvasRect.left - canvasOffset.x) / zoom;
      const y = (e.clientY - canvasRect.top - canvasOffset.y) / zoom;
      
      setDraggedConnection({
        ...draggedConnection,
        endPosition: { x, y }
      });
    }
  };

  const handleConnectionEnd = (toNodeId, toPortId, toPortType) => {
    if (draggedConnection && toNodeId && toPortId) {
      // Validate connection
      if (draggedConnection.fromPortType === 'output' && toPortType === 'input') {
        onConnectionAdd({
          from_node_id: draggedConnection.fromNodeId,
          from_port_id: draggedConnection.fromPortId,
          to_node_id: toNodeId,
          to_port_id: toPortId
        });
      } else if (draggedConnection.fromPortType === 'input' && toPortType === 'output') {
        onConnectionAdd({
          from_node_id: toNodeId,
          from_port_id: toPortId,
          to_node_id: draggedConnection.fromNodeId,
          to_port_id: draggedConnection.fromPortId
        });
      }
    }
    setDraggedConnection(null);
    setHoveredPort(null);
  };

  useEffect(() => {
    const handleMouseMove = (e) => {
      handleConnectionDrag(e);
    };

    const handleMouseUp = () => {
      if (draggedConnection && !hoveredPort) {
        setDraggedConnection(null);
      }
    };

    if (draggedConnection) {
      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);
      
      return () => {
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
      };
    }
  }, [draggedConnection, hoveredPort]);

  const getNodePosition = (nodeId) => {
    const node = logicGraph.nodes.find(n => n.id === nodeId);
    return node ? node.position : { x: 0, y: 0 };
  };

  const getPortPosition = (nodeId, portId, portType) => {
    const node = logicGraph.nodes.find(n => n.id === nodeId);
    if (!node) return { x: 0, y: 0 };

    const nodeElement = document.getElementById(`node-${nodeId}`);
    if (!nodeElement) return { x: 0, y: 0 };

    const portElement = nodeElement.querySelector(`[data-port-id="${portId}"]`);
    if (!portElement) return { x: 0, y: 0 };

    const nodeRect = nodeElement.getBoundingClientRect();
    const portRect = portElement.getBoundingClientRect();
    const canvasRect = canvasRef.current.getBoundingClientRect();

    const x = (portRect.left + portRect.width / 2 - canvasRect.left - canvasOffset.x) / zoom;
    const y = (portRect.top + portRect.height / 2 - canvasRect.top - canvasOffset.y) / zoom;

    return { x, y };
  };

  return (
    <div className="visual-logic-editor" ref={setRefs} onMouseDown={handleCanvasMouseDown} onWheel={handleWheel}>
      <div 
        className="logic-canvas"
        style={{
          transform: `translate(${canvasOffset.x}px, ${canvasOffset.y}px) scale(${zoom})`,
          transformOrigin: '0 0',
          cursor: isPanning ? 'grabbing' : isOver ? 'copy' : 'grab'
        }}
      >
        {/* Render connections */}
        <ConnectionLayer
          connections={logicGraph.connections}
          nodes={logicGraph.nodes}
          getPortPosition={getPortPosition}
          onConnectionDelete={onConnectionDelete}
          draggedConnection={draggedConnection}
          zoom={zoom}
          offset={canvasOffset}
        />

        {/* Render nodes */}
        {logicGraph.nodes.map(node => (
          <LogicNode
            key={node.id}
            node={node}
            isSelected={selectedNode?.id === node.id}
            isStartNode={logicGraph.start_node_id === node.id}
            onSelect={() => onNodeSelect(node)}
            onUpdate={(updates) => onNodeUpdate(node.id, updates)}
            onDelete={() => onNodeDelete(node.id)}
            onSetAsStart={() => onStartNodeSet(node.id)}
            onConnectionStart={handleConnectionStart}
            onConnectionEnd={handleConnectionEnd}
            onPortHover={setHoveredPort}
          />
        ))}
      </div>

      {/* Canvas controls */}
      <div className="canvas-controls">
        <button onClick={() => setZoom(zoom + 0.1)} title="Zoom In">+</button>
        <span>{Math.round(zoom * 100)}%</span>
        <button onClick={() => setZoom(zoom - 0.1)} title="Zoom Out">-</button>
        <button onClick={() => { setZoom(1); setCanvasOffset({ x: 0, y: 0 }); }} title="Reset View">
          ⟲
        </button>
      </div>

      {/* Help text */}
      {logicGraph.nodes.length === 0 && (
        <div className="empty-canvas-message">
          <p>Drag nodes from the palette to start building your rule</p>
        </div>
      )}
    </div>
  );
};

export default VisualLogicEditor;