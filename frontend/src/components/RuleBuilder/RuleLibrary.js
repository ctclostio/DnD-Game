import React, { useState, useEffect } from 'react';
import { FaBook, FaSearch, FaStar, FaDownload, FaFilter, FaTags, FaUser, FaClock } from 'react-icons/fa';
import api from '../../services/api';
import { getClickableProps } from '../../utils/accessibility';

const RuleLibrary = ({ onImportRule, currentRuleId }) => {
  const [rules, setRules] = useState([]);
  const [filteredRules, setFilteredRules] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('all');
  const [selectedComplexity, setSelectedComplexity] = useState('all');
  const [sortBy, setSortBy] = useState('popularity');
  const [showFilters, setShowFilters] = useState(false);
  const [selectedRule, setSelectedRule] = useState(null);

  useEffect(() => {
    loadRules();
  }, []);

  useEffect(() => {
    filterAndSortRules();
  }, [rules, searchTerm, selectedCategory, selectedComplexity, sortBy]);

  const loadRules = async () => {
    try {
      setLoading(true);
      const response = await api.get('/api/rules/templates');
      setRules(response.data);
    } catch (error) {
      console.error('Failed to load rule library:', error);
    } finally {
      setLoading(false);
    }
  };

  const filterAndSortRules = () => {
    let filtered = [...rules];

    // Apply search filter
    if (searchTerm) {
      filtered = filtered.filter(rule => 
        rule.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        rule.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
        rule.tags.some(tag => tag.toLowerCase().includes(searchTerm.toLowerCase()))
      );
    }

    // Apply category filter
    if (selectedCategory !== 'all') {
      filtered = filtered.filter(rule => rule.category === selectedCategory);
    }

    // Apply complexity filter
    if (selectedComplexity !== 'all') {
      filtered = filtered.filter(rule => rule.complexity === selectedComplexity);
    }

    // Sort rules
    filtered.sort((a, b) => {
      switch (sortBy) {
        case 'popularity':
          return b.usage_count - a.usage_count;
        case 'rating':
          return b.average_rating - a.average_rating;
        case 'newest':
          return new Date(b.created_at) - new Date(a.created_at);
        case 'alphabetical':
          return a.name.localeCompare(b.name);
        default:
          return 0;
      }
    });

    setFilteredRules(filtered);
  };

  const categories = [
    { value: 'all', label: 'All Categories', icon: '📚' },
    { value: 'combat', label: 'Combat', icon: '⚔️' },
    { value: 'magic', label: 'Magic', icon: '✨' },
    { value: 'skill', label: 'Skills', icon: '🎯' },
    { value: 'class_feature', label: 'Class Features', icon: '🛡️' },
    { value: 'racial', label: 'Racial Abilities', icon: '🧬' },
    { value: 'environmental', label: 'Environmental', icon: '🌍' },
    { value: 'social', label: 'Social', icon: '💬' },
    { value: 'custom', label: 'Custom Mechanics', icon: '🔧' }
  ];

  const complexityLevels = [
    { value: 'all', label: 'All Complexities' },
    { value: 'simple', label: 'Simple', color: '#27ae60' },
    { value: 'moderate', label: 'Moderate', color: '#f39c12' },
    { value: 'complex', label: 'Complex', color: '#e74c3c' },
    { value: 'advanced', label: 'Advanced', color: '#8e44ad' }
  ];

  const getComplexityColor = (complexity) => {
    const level = complexityLevels.find(l => l.value === complexity);
    return level?.color || '#95a5a6';
  };

  const renderRulePreview = (rule) => {
    return (
      <div className="rule-preview-modal" {...getClickableProps(() => setSelectedRule(null))}>
        <div className="rule-preview-content" {...getClickableProps((e) => e.stopPropagation())}>
          <div className="preview-header">
            <h3>{rule.name}</h3>
            <button className="close-btn" onClick={() => setSelectedRule(null)}>×</button>
          </div>

          <div className="preview-body">
            <p className="rule-description">{rule.description}</p>

            <div className="preview-stats">
              <div className="stat">
                <span className="stat-label">Category</span>
                <span className="stat-value">{rule.category}</span>
              </div>
              <div className="stat">
                <span className="stat-label">Complexity</span>
                <span 
                  className="stat-value complexity-badge"
                  style={{ color: getComplexityColor(rule.complexity) }}
                >
                  {rule.complexity}
                </span>
              </div>
              <div className="stat">
                <span className="stat-label">Uses</span>
                <span className="stat-value">{rule.usage_count}</span>
              </div>
              <div className="stat">
                <span className="stat-label">Rating</span>
                <span className="stat-value">
                  <FaStar style={{ color: '#f1c40f' }} /> {rule.average_rating.toFixed(1)}
                </span>
              </div>
            </div>

            {rule.parameters && rule.parameters.length > 0 && (
              <div className="preview-parameters">
                <h4>Customizable Parameters</h4>
                {rule.parameters.map((param, index) => (
                  <div key={index} className="parameter-item">
                    <span className="param-name">{param.display_name || param.name}</span>
                    <span className="param-type">{param.type}</span>
                  </div>
                ))}
              </div>
            )}

            {rule.tags && rule.tags.length > 0 && (
              <div className="preview-tags">
                {rule.tags.map((tag, index) => (
                  <span key={index} className="tag">{tag}</span>
                ))}
              </div>
            )}

            <div className="preview-actions">
              <button 
                className="btn-primary"
                onClick={() => {
                  onImportRule(rule);
                  setSelectedRule(null);
                }}
                disabled={rule.id === currentRuleId}
              >
                <FaDownload /> {rule.id === currentRuleId ? 'Current Rule' : 'Import Rule'}
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  };

  return (
    <div className="rule-library">
      <div className="library-header">
        <h3><FaBook /> Rule Library</h3>
        <button 
          className="filter-toggle"
          onClick={() => setShowFilters(!showFilters)}
        >
          <FaFilter /> Filters
        </button>
      </div>

      {/* Search Bar */}
      <div className="library-search">
        <FaSearch />
        <input
          type="text"
          placeholder="Search rules by name, description, or tags..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
        />
      </div>

      {/* Filters */}
      {showFilters && (
        <div className="library-filters">
          <div className="filter-group">
            <label>Category</label>
            <select
              value={selectedCategory}
              onChange={(e) => setSelectedCategory(e.target.value)}
            >
              {categories.map(cat => (
                <option key={cat.value} value={cat.value}>
                  {cat.icon} {cat.label}
                </option>
              ))}
            </select>
          </div>

          <div className="filter-group">
            <label>Complexity</label>
            <select
              value={selectedComplexity}
              onChange={(e) => setSelectedComplexity(e.target.value)}
            >
              {complexityLevels.map(level => (
                <option key={level.value} value={level.value}>
                  {level.label}
                </option>
              ))}
            </select>
          </div>

          <div className="filter-group">
            <label>Sort By</label>
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value)}
            >
              <option value="popularity">Most Popular</option>
              <option value="rating">Highest Rated</option>
              <option value="newest">Newest First</option>
              <option value="alphabetical">Alphabetical</option>
            </select>
          </div>
        </div>
      )}

      {/* Results Count */}
      <div className="results-info">
        Showing {filteredRules.length} of {rules.length} rules
      </div>

      {/* Rule Grid */}
      {loading ? (
        <div className="library-loading">
          <div className="spinner" />
          <p>Loading rule library...</p>
        </div>
      ) : (
        <div className="rules-grid">
          {filteredRules.map(rule => (
            <div 
              key={rule.id} 
              className={`rule-card ${rule.id === currentRuleId ? 'current' : ''}`}
              {...getClickableProps(() => setSelectedRule(rule))}
            >
              <div className="rule-card-header">
                <h4>{rule.name}</h4>
                <span 
                  className="complexity-indicator"
                  style={{ backgroundColor: getComplexityColor(rule.complexity) }}
                  title={`Complexity: ${rule.complexity}`}
                />
              </div>

              <p className="rule-card-description">{rule.description}</p>

              <div className="rule-card-meta">
                <div className="meta-item">
                  <FaUser />
                  <span>{rule.created_by || 'Community'}</span>
                </div>
                <div className="meta-item">
                  <FaStar />
                  <span>{rule.average_rating.toFixed(1)}</span>
                </div>
                <div className="meta-item">
                  <FaDownload />
                  <span>{rule.usage_count}</span>
                </div>
              </div>

              {rule.tags && rule.tags.length > 0 && (
                <div className="rule-card-tags">
                  {rule.tags.slice(0, 3).map((tag, index) => (
                    <span key={index} className="tag">{tag}</span>
                  ))}
                  {rule.tags.length > 3 && (
                    <span className="tag more">+{rule.tags.length - 3}</span>
                  )}
                </div>
              )}

              {rule.id === currentRuleId && (
                <div className="current-indicator">Current Rule</div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Empty State */}
      {!loading && filteredRules.length === 0 && (
        <div className="library-empty">
          <p>No rules found matching your criteria</p>
          {searchTerm && (
            <button 
              className="btn-secondary"
              onClick={() => setSearchTerm('')}
            >
              Clear Search
            </button>
          )}
        </div>
      )}

      {/* Rule Preview Modal */}
      {selectedRule && renderRulePreview(selectedRule)}
    </div>
  );
};

export default RuleLibrary;