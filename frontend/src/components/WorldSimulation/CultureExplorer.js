import React, { useState, useEffect } from 'react';
import { 
  FaGlobeAsia, FaLanguage, FaPalette, FaMusic, 
  FaUtensils, FaHome, FaPray, FaUsers, FaPlus,
  FaBook, FaTshirt, FaHandshake
} from 'react-icons/fa';
import api from '../../services/api';
import { getClickableProps, getSelectableProps } from '../../utils/accessibility';

const CultureExplorer = ({ sessionId, isDM }) => {
  const [cultures, setCultures] = useState([]);
  const [selectedCulture, setSelectedCulture] = useState(null);
  const [activeTab, setActiveTab] = useState('overview');
  const [loading, setLoading] = useState(false);
  const [showGenerateModal, setShowGenerateModal] = useState(false);
  const [generationParams, setGenerationParams] = useState({
    environment: 'forest',
    historical_context: '',
    neighboring_cultures: [],
    special_traits: []
  });

  useEffect(() => {
    loadCultures();
  }, [sessionId]);

  const loadCultures = async () => {
    try {
      const response = await api.get(`/sessions/${sessionId}/cultures`);
      setCultures(response.data);
      if (response.data.length > 0 && !selectedCulture) {
        setSelectedCulture(response.data[0]);
      }
    } catch (err) {
      console.error('Failed to load cultures:', err);
    }
  };

  const generateCulture = async () => {
    setLoading(true);
    try {
      const response = await api.post(`/sessions/${sessionId}/cultures/generate`, generationParams);
      setCultures([response.data, ...cultures]);
      setSelectedCulture(response.data);
      setShowGenerateModal(false);
      setGenerationParams({
        environment: 'forest',
        historical_context: '',
        neighboring_cultures: [],
        special_traits: []
      });
    } catch (err) {
      console.error('Failed to generate culture:', err);
    } finally {
      setLoading(false);
    }
  };

  const interactWithCulture = async (actionType, approach) => {
    if (!selectedCulture || !isDM) return;

    try {
      const response = await api.post(`/cultures/${selectedCulture.id}/interact`, {
        type: actionType,
        approach: approach,
        magnitude: 0.5,
        description: `Players ${actionType} with ${approach} approach`
      });
      
      if (response.data.culture) {
        setSelectedCulture(response.data.culture);
        // Update in list
        setCultures(cultures.map(c => 
          c.id === response.data.culture.id ? response.data.culture : c
        ));
      }
    } catch (err) {
      console.error('Failed to interact with culture:', err);
    }
  };

  const renderLanguage = () => {
    if (!selectedCulture) return null;
    const lang = selectedCulture.language;

    return (
      <div className="culture-language">
        <h4><FaLanguage /> Language: {lang.name}</h4>
        
        <div className="language-section">
          <h5>Common Words</h5>
          <div className="word-grid">
            {Object.entries(lang.common_words || {}).map(([english, translation]) => (
              <div key={english} className="word-pair">
                <span className="english">{english}</span>
                <span className="translation">{translation}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="language-section">
          <h5>Greetings</h5>
          <div className="greetings-list">
            {Object.entries(selectedCulture.greetings || {}).map(([context, greeting]) => (
              <div key={context} className="greeting-item">
                <span className="context">{context}:</span>
                <span className="greeting">"{greeting}"</span>
              </div>
            ))}
          </div>
        </div>

        {lang.idioms && lang.idioms.length > 0 && (
          <div className="language-section">
            <h5>Idioms</h5>
            {lang.idioms.map((idiom, idx) => (
              <div key={idx} className="idiom">
                <p className="expression">"{idiom.expression}"</p>
                <p className="meaning">Meaning: {idiom.meaning}</p>
              </div>
            ))}
          </div>
        )}
      </div>
    );
  };

  const renderBeliefs = () => {
    if (!selectedCulture) return null;
    const beliefs = selectedCulture.belief_system;

    return (
      <div className="culture-beliefs">
        <h4><FaPray /> {beliefs.name}</h4>
        <p className="belief-type">Type: {beliefs.type}</p>

        {beliefs.deities && beliefs.deities.length > 0 && (
          <div className="deities-section">
            <h5>Deities</h5>
            <div className="deity-cards">
              {beliefs.deities.map((deity, idx) => (
                <div key={idx} className="deity-card">
                  <h6>{deity.name}</h6>
                  <p className="deity-title">{deity.title}</p>
                  <p className="deity-domains">Domains: {deity.domain.join(', ')}</p>
                  <p className="deity-symbol">Symbol: {deity.symbol}</p>
                </div>
              ))}
            </div>
          </div>
        )}

        <div className="beliefs-section">
          <h5>Core Beliefs</h5>
          <ul>
            {beliefs.core_beliefs.map((belief, idx) => (
              <li key={idx}>{belief}</li>
            ))}
          </ul>
        </div>

        <div className="afterlife-section">
          <h5>View of Afterlife</h5>
          <p>{beliefs.afterlife}</p>
        </div>
      </div>
    );
  };

  const renderCustoms = () => {
    if (!selectedCulture) return null;

    return (
      <div className="culture-customs">
        <h4><FaHandshake /> Customs & Traditions</h4>
        
        <div className="customs-grid">
          {selectedCulture.customs.map((custom, idx) => (
            <div key={idx} className="custom-card">
              <h5>{custom.name}</h5>
              <span className="custom-type">{custom.type}</span>
              <p>{custom.description}</p>
              <div className="custom-details">
                <span><strong>Frequency:</strong> {custom.frequency}</span>
                <span><strong>Participants:</strong> {custom.participants}</span>
              </div>
            </div>
          ))}
        </div>

        <div className="taboos-section">
          <h5>Cultural Taboos</h5>
          <ul className="taboo-list">
            {selectedCulture.taboos.map((taboo, idx) => (
              <li key={idx} className="taboo-item">{taboo}</li>
            ))}
          </ul>
        </div>
      </div>
    );
  };

  const renderArtAndCulture = () => {
    if (!selectedCulture) return null;

    return (
      <div className="culture-arts">
        <div className="art-section">
          <h4><FaPalette /> Art Style</h4>
          <p className="style-description">{selectedCulture.art_style.style_description}</p>
          
          <div className="art-details">
            <div className="detail-group">
              <h6>Primary Mediums</h6>
              <div className="tag-list">
                {selectedCulture.art_style.primary_mediums.map((medium, idx) => (
                  <span key={idx} className="tag">{medium}</span>
                ))}
              </div>
            </div>

            <div className="detail-group">
              <h6>Common Motifs</h6>
              <div className="tag-list">
                {selectedCulture.art_style.common_motifs.map((motif, idx) => (
                  <span key={idx} className="tag">{motif}</span>
                ))}
              </div>
            </div>

            <div className="detail-group">
              <h6>Color Palette</h6>
              <div className="color-palette">
                {selectedCulture.art_style.color_palette.map((color, idx) => (
                  <div 
                    key={idx} 
                    className="color-swatch"
                    style={{ backgroundColor: color.toLowerCase() }}
                    title={color}
                  />
                ))}
              </div>
            </div>
          </div>
        </div>

        <div className="music-section">
          <h4><FaMusic /> Music Style</h4>
          <div className="music-details">
            <p><strong>Style:</strong> {selectedCulture.music_style.name}</p>
            <p><strong>Instruments:</strong> {selectedCulture.music_style.instruments.join(', ')}</p>
            <p><strong>Common Themes:</strong> {selectedCulture.music_style.themes.join(', ')}</p>
            <p><strong>Dance Styles:</strong> {selectedCulture.music_style.dance_styles.join(', ')}</p>
          </div>
        </div>

        <div className="cuisine-section">
          <h4><FaUtensils /> Cuisine</h4>
          <div className="dishes-grid">
            {selectedCulture.cuisine.map((dish, idx) => (
              <div key={idx} className="dish-card">
                <h6>{dish.name}</h6>
                <span className="dish-type">{dish.type}</span>
                <p className="ingredients">Ingredients: {dish.ingredients.join(', ')}</p>
                <p className="occasion">Served: {dish.occasion}</p>
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  };

  const renderSocialStructure = () => {
    if (!selectedCulture) return null;
    const social = selectedCulture.social_structure;

    return (
      <div className="culture-social">
        <h4><FaUsers /> Social Structure</h4>
        <p className="structure-type">Type: {social.type}</p>
        <p className="mobility">Social Mobility: {social.mobility}</p>

        <div className="social-classes">
          <h5>Social Classes</h5>
          {social.classes.map((socialClass, idx) => (
            <div key={idx} className="social-class">
              <div className="class-header">
                <h6>{socialClass.name}</h6>
                <span className="class-rank">Rank {socialClass.rank}</span>
              </div>
              <div className="class-details">
                <p><strong>Privileges:</strong> {socialClass.privileges.join(', ')}</p>
                <p><strong>Restrictions:</strong> {socialClass.restrictions.join(', ')}</p>
                <p><strong>Occupations:</strong> {socialClass.occupations.join(', ')}</p>
              </div>
            </div>
          ))}
        </div>

        <div className="social-details">
          <p><strong>Leadership:</strong> {social.leadership}</p>
          <p><strong>Family Unit:</strong> {social.family_unit}</p>
          <p><strong>Treatment of Outsiders:</strong> {social.outsiders}</p>
        </div>
      </div>
    );
  };

  return (
    <div className="culture-explorer">
      <div className="culture-list">
        <div className="list-header">
          <h3>Procedural Cultures</h3>
          {isDM && (
            <button 
              className="btn-generate"
              onClick={() => setShowGenerateModal(true)}
            >
              <FaPlus /> Generate Culture
            </button>
          )}
        </div>

        <div className="culture-cards">
          {cultures.map(culture => (
            <div 
              key={culture.id}
              className={`culture-card ${selectedCulture?.id === culture.id ? 'selected' : ''}`}
              onClick={() => setSelectedCulture(culture)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  setSelectedCulture(culture);
                }
              }}
              role="button"
              tabIndex={0}
              aria-pressed={selectedCulture?.id === culture.id}
            >
              <div className="culture-icon">
                <FaGlobeAsia />
              </div>
              <h4>{culture.name}</h4>
              <div className="culture-values">
                {Object.entries(culture.values)
                  .sort(([,a], [,b]) => b - a)
                  .slice(0, 3)
                  .map(([value, score]) => (
                    <span key={value} className="value-tag">
                      {value} ({Math.round(score * 100)}%)
                    </span>
                  ))}
              </div>
            </div>
          ))}
        </div>
      </div>

      {selectedCulture && (
        <div className="culture-details">
          <div className="detail-header">
            <h3>{selectedCulture.name} Culture</h3>
            {isDM && (
              <div className="interaction-buttons">
                <button 
                  onClick={() => interactWithCulture('trade', 'respectful')}
                  className="btn-interact"
                >
                  Trade Peacefully
                </button>
                <button 
                  onClick={() => interactWithCulture('diplomacy', 'respectful')}
                  className="btn-interact"
                >
                  Diplomatic Contact
                </button>
                <button 
                  onClick={() => interactWithCulture('influence', 'subversive')}
                  className="btn-interact"
                >
                  Cultural Influence
                </button>
              </div>
            )}
          </div>

          <div className="culture-tabs">
            <button 
              className={activeTab === 'overview' ? 'active' : ''}
              onClick={() => setActiveTab('overview')}
            >
              Overview
            </button>
            <button 
              className={activeTab === 'language' ? 'active' : ''}
              onClick={() => setActiveTab('language')}
            >
              <FaLanguage /> Language
            </button>
            <button 
              className={activeTab === 'beliefs' ? 'active' : ''}
              onClick={() => setActiveTab('beliefs')}
            >
              <FaPray /> Beliefs
            </button>
            <button 
              className={activeTab === 'customs' ? 'active' : ''}
              onClick={() => setActiveTab('customs')}
            >
              <FaHandshake /> Customs
            </button>
            <button 
              className={activeTab === 'arts' ? 'active' : ''}
              onClick={() => setActiveTab('arts')}
            >
              <FaPalette /> Arts
            </button>
            <button 
              className={activeTab === 'social' ? 'active' : ''}
              onClick={() => setActiveTab('social')}
            >
              <FaUsers /> Society
            </button>
          </div>

          <div className="tab-content">
            {activeTab === 'overview' && (
              <div className="culture-overview">
                <div className="value-chart">
                  <h4>Cultural Values</h4>
                  <div className="values-list">
                    {Object.entries(selectedCulture.values).map(([value, score]) => (
                      <div key={value} className="value-item">
                        <span className="value-name">{value}</span>
                        <div className="value-bar">
                          <div 
                            className="value-fill"
                            style={{ width: `${score * 100}%` }}
                          />
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                <div className="origin-story">
                  <h4><FaBook /> Origin Story</h4>
                  <p>{selectedCulture.metadata.origin_story || 'The origins of this culture are shrouded in mystery...'}</p>
                </div>
              </div>
            )}
            {activeTab === 'language' && renderLanguage()}
            {activeTab === 'beliefs' && renderBeliefs()}
            {activeTab === 'customs' && renderCustoms()}
            {activeTab === 'arts' && renderArtAndCulture()}
            {activeTab === 'social' && renderSocialStructure()}
          </div>
        </div>
      )}

      {showGenerateModal && (
        <div className="modal-overlay" {...getClickableProps(() => setShowGenerateModal(false))}>
          <div className="modal-content" {...getClickableProps(e => e.stopPropagation())}>
            <h3>Generate New Culture</h3>
            
            <div className="form-group">
              <label>Environment</label>
              <select
                value={generationParams.environment}
                onChange={(e) => setGenerationParams({...generationParams, environment: e.target.value})}
              >
                <option value="forest">Forest</option>
                <option value="mountain">Mountain</option>
                <option value="desert">Desert</option>
                <option value="coastal">Coastal</option>
                <option value="plains">Plains</option>
                <option value="swamp">Swamp</option>
                <option value="tundra">Tundra</option>
                <option value="volcanic">Volcanic</option>
              </select>
            </div>

            <div className="form-group">
              <label>Historical Context</label>
              <textarea
                value={generationParams.historical_context}
                onChange={(e) => setGenerationParams({...generationParams, historical_context: e.target.value})}
                placeholder="Describe significant historical events..."
                rows="3"
              />
            </div>

            <div className="form-group">
              <label>Special Traits</label>
              <input
                type="text"
                placeholder="e.g., magical affinity, warrior culture, peaceful traders"
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    setGenerationParams({
                      ...generationParams,
                      special_traits: [...generationParams.special_traits, e.target.value]
                    });
                    e.target.value = '';
                  }
                }}
              />
              <div className="trait-tags">
                {generationParams.special_traits.map((trait, idx) => (
                  <span key={idx} className="trait-tag">
                    {trait}
                    <button onClick={() => {
                      setGenerationParams({
                        ...generationParams,
                        special_traits: generationParams.special_traits.filter((_, i) => i !== idx)
                      });
                    }}>&times;</button>
                  </span>
                ))}
              </div>
            </div>

            <div className="modal-actions">
              <button onClick={() => setShowGenerateModal(false)}>Cancel</button>
              <button 
                onClick={generateCulture}
                disabled={loading}
                className="btn-primary"
              >
                {loading ? 'Generating...' : 'Generate Culture'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default CultureExplorer;