import React, { useState, useEffect } from 'react';

const SettlementDetails = ({ settlement, sessionId, onUpdate }) => {
    const [activeSection, setActiveSection] = useState('overview');
    const [market, setMarket] = useState(null);
    const [priceCheck, setPriceCheck] = useState({ basePrice: 0, itemType: '', result: null });
    
    // Helper render methods to reduce complexity
    const renderOverviewSection = () => (
        <div className="overview-section">
            <div className="description">
                <h4>Description</h4>
                <p>{settlement.description}</p>
            </div>

            <div className="history">
                <h4>History</h4>
                <p>{settlement.history}</p>
            </div>

            <div className="government">
                <h4>Government</h4>
                <p><strong>Type:</strong> {settlement.governmentType}</p>
                <p><strong>Alignment:</strong> {settlement.alignment}</p>
            </div>

            <div className="notable-locations">
                <h4>Notable Locations</h4>
                {settlement.notableLocations?.map((location, index) => (
                    <div key={index} className="location-item">
                        <h5>{location.name}</h5>
                        <p>{location.description}</p>
                    </div>
                ))}
            </div>

            <div className="problems">
                <h4>Current Problems</h4>
                {settlement.problems?.map((problem, index) => (
                    <div key={index} className="problem-item">
                        <h5>{problem.title}</h5>
                        <p>{problem.description}</p>
                    </div>
                ))}
            </div>

            {settlement.secrets?.length > 0 && (
                <div className="secrets">
                    <h4>Secrets (DM Only)</h4>
                    {settlement.secrets.map((secret, index) => (
                        <div key={index} className="secret-item">
                            <h5>{secret.title}</h5>
                            <p>{secret.description}</p>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
    
    const renderNPCSection = () => (
        <div className="npcs-section">
            {settlement.npcs?.length === 0 ? (
                <p className="empty">No NPCs yet</p>
            ) : (
                settlement.npcs?.map(npc => (
                    <div key={npc.id} className="npc-card">
                        <div className="npc-header">
                            <h5>{npc.name}</h5>
                            <span className="npc-role">{npc.role}</span>
                        </div>
                        <div className="npc-details">
                            <p><strong>Race:</strong> {npc.race}</p>
                            <p><strong>Occupation:</strong> {npc.occupation}</p>
                            {npc.ancientKnowledge && (
                                <span className="ancient-knowledge">📜 Ancient Knowledge</span>
                            )}
                            {npc.corruptionTouched && (
                                <span className="corruption-touched">🌑 Corruption Touched</span>
                            )}
                        </div>
                    </div>
                ))
            )}
        </div>
    );
    
    const renderShopSection = () => (
        <div className="shops-section">
            {settlement.shops?.length === 0 ? (
                <p className="empty">No shops yet</p>
            ) : (
                settlement.shops?.map(shop => (
                    <div key={shop.id} className="shop-card">
                        <div className="shop-header">
                            <h5>{shop.name}</h5>
                            <span className="shop-type">{shop.type}</span>
                        </div>
                        <div className="shop-details">
                            <p><strong>Quality:</strong> {shop.qualityLevel}/10</p>
                            <p><strong>Price Modifier:</strong> {Math.round(shop.priceModifier * 100)}%</p>
                            {shop.blackMarket && (
                                <span className="black-market">🕵️ Black Market</span>
                            )}
                            {shop.ancientArtifacts && (
                                <span className="ancient-artifacts">🗿 Ancient Artifacts</span>
                            )}
                        </div>
                        {shop.currentRumors?.length > 0 && (
                            <div className="shop-rumors">
                                <strong>Rumors:</strong>
                                <ul>
                                    {shop.currentRumors.map((rumor, index) => (
                                        <li key={index}>{rumor}</li>
                                    ))}
                                </ul>
                            </div>
                        )}
                    </div>
                ))
            )}
        </div>
    );
    
    const renderMarketSection = () => {
        if (!market) {
            return <p className="empty">Loading market data...</p>;
        }
        
        return (
            <>
                <div className="market-conditions">
                    <h4>Market Conditions</h4>
                    <div className="price-modifiers">
                        <div className="modifier-item">
                            <span>Food:</span>
                            <span className={market.foodPriceModifier > 1 ? 'high' : 'low'}>
                                {Math.round(market.foodPriceModifier * 100)}%
                            </span>
                        </div>
                        <div className="modifier-item">
                            <span>Common Goods:</span>
                            <span className={market.commonGoodsModifier > 1 ? 'high' : 'low'}>
                                {Math.round(market.commonGoodsModifier * 100)}%
                            </span>
                        </div>
                        <div className="modifier-item">
                            <span>Weapons & Armor:</span>
                            <span className={market.weaponsArmorModifier > 1 ? 'high' : 'low'}>
                                {Math.round(market.weaponsArmorModifier * 100)}%
                            </span>
                        </div>
                        <div className="modifier-item">
                            <span>Magical Items:</span>
                            <span className={market.magicalItemsModifier > 1 ? 'high' : 'low'}>
                                {Math.round(market.magicalItemsModifier * 100)}%
                            </span>
                        </div>
                    </div>
                </div>
            </>
        );
    };

    useEffect(() => {
        if (settlement) {
            loadMarketData();
        }
    }, [settlement]);

    const loadMarketData = async () => {
        try {
            const response = await fetch(`/api/v1/settlements/${settlement.id}/market`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (response.ok) {
                const marketData = await response.json();
                setMarket(marketData);
            }
        } catch (err) {
            console.error('Failed to load market data:', err);
        }
    };

    const calculatePrice = async () => {
        if (!priceCheck.basePrice || !priceCheck.itemType) return;

        try {
            const response = await fetch(`/api/v1/settlements/${settlement.id}/calculate-price`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    basePrice: parseFloat(priceCheck.basePrice),
                    itemType: priceCheck.itemType
                })
            });

            if (response.ok) {
                const result = await response.json();
                setPriceCheck(prev => ({ ...prev, result }));
            }
        } catch (err) {
            console.error('Failed to calculate price:', err);
        }
    };

    return (
        <div className="settlement-details">
            <div className="details-header">
                <h3>{settlement.name}</h3>
                <span className="settlement-age">{settlement.ageCategory}</span>
            </div>

            <div className="details-tabs">
                <button
                    className={activeSection === 'overview' ? 'active' : ''}
                    onClick={() => setActiveSection('overview')}
                >
                    Overview
                </button>
                <button
                    className={activeSection === 'npcs' ? 'active' : ''}
                    onClick={() => setActiveSection('npcs')}
                >
                    NPCs ({settlement.npcs?.length || 0})
                </button>
                <button
                    className={activeSection === 'shops' ? 'active' : ''}
                    onClick={() => setActiveSection('shops')}
                >
                    Shops ({settlement.shops?.length || 0})
                </button>
                <button
                    className={activeSection === 'market' ? 'active' : ''}
                    onClick={() => setActiveSection('market')}
                >
                    Market
                </button>
            </div>

            <div className="details-content">
                {activeSection === 'overview' && renderOverviewSection()}
                {activeSection === 'npcs' && renderNPCSection()}
                {activeSection === 'shops' && renderShopSection()}
                {activeSection === 'market' && (
                    <div className="market-section">
                        {renderMarketSection()}
                        {market ? (
                            <>
                                <div className="market-conditions">
                                    <h4>Market Conditions</h4>
                                    <div className="price-modifiers">
                                        <div className="modifier-item">
                                            <span>Food:</span>
                                            <span className={market.foodPriceModifier > 1 ? 'high' : 'low'}>
                                                {Math.round(market.foodPriceModifier * 100)}%
                                            </span>
                                        </div>
                                        <div className="modifier-item">
                                            <span>Common Goods:</span>
                                            <span className={market.commonGoodsModifier > 1 ? 'high' : 'low'}>
                                                {Math.round(market.commonGoodsModifier * 100)}%
                                            </span>
                                        </div>
                                        <div className="modifier-item">
                                            <span>Weapons & Armor:</span>
                                            <span className={market.weaponsArmorModifier > 1 ? 'high' : 'low'}>
                                                {Math.round(market.weaponsArmorModifier * 100)}%
                                            </span>
                                        </div>
                                        <div className="modifier-item">
                                            <span>Magical Items:</span>
                                            <span className={market.magicalItemsModifier > 1 ? 'high' : 'low'}>
                                                {Math.round(market.magicalItemsModifier * 100)}%
                                            </span>
                                        </div>
                                        <div className="modifier-item">
                                            <span>Ancient Artifacts:</span>
                                            <span className={market.ancientArtifactsModifier > 1 ? 'high' : 'low'}>
                                                {Math.round(market.ancientArtifactsModifier * 100)}%
                                            </span>
                                        </div>
                                    </div>

                                    {market.economicBoom && (
                                        <div className="market-status boom">📈 Economic Boom</div>
                                    )}
                                    {market.economicDepression && (
                                        <div className="market-status depression">📉 Economic Depression</div>
                                    )}
                                    {market.blackMarketActive && (
                                        <div className="market-status black-market">🕵️ Black Market Active</div>
                                    )}
                                    {market.artifactDealerPresent && (
                                        <div className="market-status artifacts">🗿 Artifact Dealer Present</div>
                                    )}
                                </div>

                                <div className="price-calculator">
                                    <h4>Price Calculator</h4>
                                    <div className="calculator-inputs">
                                        <input
                                            type="number"
                                            placeholder="Base Price (gp)"
                                            value={priceCheck.basePrice}
                                            onChange={(e) => setPriceCheck(prev => ({
                                                ...prev,
                                                basePrice: e.target.value
                                            }))}
                                        />
                                        <select
                                            value={priceCheck.itemType}
                                            onChange={(e) => setPriceCheck(prev => ({
                                                ...prev,
                                                itemType: e.target.value
                                            }))}
                                        >
                                            <option value="">Select Item Type</option>
                                            <option value="food">Food/Rations</option>
                                            <option value="weapon">Weapon</option>
                                            <option value="armor">Armor</option>
                                            <option value="magic">Magic Item</option>
                                            <option value="artifact">Ancient Artifact</option>
                                            <option value="common">Common Goods</option>
                                        </select>
                                        <button onClick={calculatePrice}>Calculate</button>
                                    </div>

                                    {priceCheck.result && (
                                        <div className="price-result">
                                            <p>Base Price: {priceCheck.result.basePrice} gp</p>
                                            <p className="adjusted-price">
                                                Adjusted Price: {Math.round(priceCheck.result.adjustedPrice * 10) / 10} gp
                                            </p>
                                        </div>
                                    )}
                                </div>
                            </>
                        ) : (
                            <p>Loading market data...</p>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
};

export default SettlementDetails;