import React from 'react';
import { Link } from 'react-router-dom';
import { FaBook, FaCog, FaGlobe, FaUsers, FaDice, FaPalette, FaHammer } from 'react-icons/fa';

const DMTools: React.FC = () => {
  const tools = [
    {
      title: 'Rule Builder',
      description: 'Create custom mechanics with visual logic programming',
      icon: <FaHammer />,
      link: '/dm-tools/rule-builder',
      color: '#e74c3c'
    },
    {
      title: 'DM Assistant',
      description: 'AI-powered assistant for dynamic storytelling',
      icon: <FaBook />,
      link: '/dm-tools/assistant',
      color: '#3498db'
    },
    {
      title: 'World Builder',
      description: 'Create and manage your campaign world',
      icon: <FaGlobe />,
      link: '/dm-tools/world-builder',
      color: '#2ecc71'
    },
    {
      title: 'Campaign Manager',
      description: 'Track campaign progress and player actions',
      icon: <FaUsers />,
      link: '/dm-tools/campaign',
      color: '#f39c12'
    },
    {
      title: 'Encounter Builder',
      description: 'Design balanced encounters for your party',
      icon: <FaDice />,
      link: '/dm-tools/encounters',
      color: '#9b59b6'
    },
    {
      title: 'Narrative Engine',
      description: 'AI-powered dynamic storytelling system',
      icon: <FaPalette />,
      link: '/dm-tools/narrative',
      color: '#1abc9c'
    }
  ];

  return (
    <div className="dm-tools-hub">
      <div className="dm-tools-header">
        <h1>Dungeon Master Tools</h1>
        <p>Powerful tools to enhance your game mastering experience</p>
      </div>

      <div className="dm-tools-grid">
        {tools.map((tool, index) => (
          <Link
            key={index}
            to={tool.link}
            className="dm-tool-card"
            style={{ borderColor: tool.color }}
          >
            <div className="tool-icon" style={{ color: tool.color }}>
              {tool.icon}
            </div>
            <h3>{tool.title}</h3>
            <p>{tool.description}</p>
          </Link>
        ))}
      </div>
    </div>
  );
};

export default DMTools;