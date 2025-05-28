import React, { useState, useEffect } from 'react';
import { getCharacterInventory, equipItem, unequipItem, attuneItem, unattuneItem, 
         getCharacterCurrency, updateCharacterCurrency, getCharacterWeight } from '../services/api';

const InventoryView = ({ characterId }) => {
    const [inventory, setInventory] = useState([]);
    const [currency, setCurrency] = useState({
        copper: 0,
        silver: 0,
        electrum: 0,
        gold: 0,
        platinum: 0
    });
    const [weight, setWeight] = useState({
        current_weight: 0,
        carry_capacity: 0,
        encumbered: false,
        heavily_encumbered: false
    });
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [selectedTab, setSelectedTab] = useState('all');

    useEffect(() => {
        if (characterId) {
            loadInventoryData();
        }
    }, [characterId]);

    const loadInventoryData = async () => {
        setLoading(true);
        try {
            const [invResponse, currResponse, weightResponse] = await Promise.all([
                getCharacterInventory(characterId),
                getCharacterCurrency(characterId),
                getCharacterWeight(characterId)
            ]);
            
            setInventory(invResponse.data || []);
            setCurrency(currResponse.data);
            setWeight(weightResponse.data);
            setError('');
        } catch (err) {
            setError('Failed to load inventory data');
            console.error('Error loading inventory:', err);
        } finally {
            setLoading(false);
        }
    };

    const handleEquip = async (itemId, currentlyEquipped) => {
        try {
            if (currentlyEquipped) {
                await unequipItem(characterId, itemId);
            } else {
                await equipItem(characterId, itemId);
            }
            await loadInventoryData();
        } catch (err) {
            setError('Failed to equip/unequip item');
        }
    };

    const handleAttune = async (itemId, currentlyAttuned) => {
        try {
            if (currentlyAttuned) {
                await unattuneItem(characterId, itemId);
            } else {
                await attuneItem(characterId, itemId);
            }
            await loadInventoryData();
        } catch (err) {
            setError(err.response?.data?.error || 'Failed to attune/unattune item');
        }
    };

    const handleCurrencyChange = async (type, amount) => {
        const newCurrency = { ...currency };
        newCurrency[type] = Math.max(0, newCurrency[type] + amount);
        
        try {
            const response = await updateCharacterCurrency(characterId, {
                copper: newCurrency.copper - currency.copper,
                silver: newCurrency.silver - currency.silver,
                electrum: newCurrency.electrum - currency.electrum,
                gold: newCurrency.gold - currency.gold,
                platinum: newCurrency.platinum - currency.platinum
            });
            setCurrency(response.data);
        } catch (err) {
            setError('Failed to update currency');
        }
    };

    const getTotalValue = () => {
        return currency.copper + 
               (currency.silver * 10) + 
               (currency.electrum * 50) + 
               (currency.gold * 100) + 
               (currency.platinum * 1000);
    };

    const filterItems = (items) => {
        switch(selectedTab) {
            case 'equipped':
                return items.filter(item => item.equipped);
            case 'weapons':
                return items.filter(item => item.item?.type === 'weapon');
            case 'armor':
                return items.filter(item => item.item?.type === 'armor');
            case 'magic':
                return items.filter(item => item.item?.rarity !== 'common');
            default:
                return items;
        }
    };

    if (loading) return <div className="loading">Loading inventory...</div>;

    return (
        <div className="inventory-view">
            <h2>Inventory</h2>
            
            {error && <div className="error">{error}</div>}
            
            <div className="inventory-stats">
                <div className="weight-info">
                    <h3>Weight</h3>
                    <div className={`weight-bar ${weight.encumbered ? 'encumbered' : ''} ${weight.heavily_encumbered ? 'heavily-encumbered' : ''}`}>
                        <div 
                            className="weight-fill" 
                            style={{ width: `${Math.min(100, (weight.current_weight / weight.carry_capacity) * 100)}%` }}
                        />
                    </div>
                    <p>{weight.current_weight.toFixed(1)} / {weight.carry_capacity} lbs</p>
                    {weight.encumbered && <p className="warning">Encumbered!</p>}
                    {weight.heavily_encumbered && <p className="warning">Heavily Encumbered!</p>}
                </div>
                
                <div className="currency-info">
                    <h3>Currency</h3>
                    <div className="currency-grid">
                        <div className="currency-item">
                            <label>CP</label>
                            <input 
                                type="number" 
                                value={currency.copper} 
                                onChange={(e) => handleCurrencyChange('copper', parseInt(e.target.value) - currency.copper)}
                            />
                        </div>
                        <div className="currency-item">
                            <label>SP</label>
                            <input 
                                type="number" 
                                value={currency.silver} 
                                onChange={(e) => handleCurrencyChange('silver', parseInt(e.target.value) - currency.silver)}
                            />
                        </div>
                        <div className="currency-item">
                            <label>EP</label>
                            <input 
                                type="number" 
                                value={currency.electrum} 
                                onChange={(e) => handleCurrencyChange('electrum', parseInt(e.target.value) - currency.electrum)}
                            />
                        </div>
                        <div className="currency-item">
                            <label>GP</label>
                            <input 
                                type="number" 
                                value={currency.gold} 
                                onChange={(e) => handleCurrencyChange('gold', parseInt(e.target.value) - currency.gold)}
                            />
                        </div>
                        <div className="currency-item">
                            <label>PP</label>
                            <input 
                                type="number" 
                                value={currency.platinum} 
                                onChange={(e) => handleCurrencyChange('platinum', parseInt(e.target.value) - currency.platinum)}
                            />
                        </div>
                    </div>
                    <p className="total-value">Total: {getTotalValue()} CP</p>
                </div>
            </div>
            
            <div className="inventory-tabs">
                <button 
                    className={selectedTab === 'all' ? 'active' : ''} 
                    onClick={() => setSelectedTab('all')}
                >
                    All Items
                </button>
                <button 
                    className={selectedTab === 'equipped' ? 'active' : ''} 
                    onClick={() => setSelectedTab('equipped')}
                >
                    Equipped
                </button>
                <button 
                    className={selectedTab === 'weapons' ? 'active' : ''} 
                    onClick={() => setSelectedTab('weapons')}
                >
                    Weapons
                </button>
                <button 
                    className={selectedTab === 'armor' ? 'active' : ''} 
                    onClick={() => setSelectedTab('armor')}
                >
                    Armor
                </button>
                <button 
                    className={selectedTab === 'magic' ? 'active' : ''} 
                    onClick={() => setSelectedTab('magic')}
                >
                    Magic Items
                </button>
            </div>
            
            <div className="inventory-items">
                {filterItems(inventory).length === 0 ? (
                    <p className="empty">No items in this category</p>
                ) : (
                    filterItems(inventory).map(invItem => (
                        <div key={invItem.id} className={`inventory-item ${invItem.equipped ? 'equipped' : ''} ${invItem.attuned ? 'attuned' : ''}`}>
                            <div className="item-header">
                                <h4>{invItem.item.name}</h4>
                                <span className={`rarity ${invItem.item.rarity}`}>{invItem.item.rarity}</span>
                            </div>
                            
                            <div className="item-details">
                                <p className="item-type">{invItem.item.type}</p>
                                <p className="item-weight">Weight: {invItem.item.weight * invItem.quantity} lbs</p>
                                {invItem.quantity > 1 && <p className="item-quantity">Quantity: {invItem.quantity}</p>}
                            </div>
                            
                            {invItem.item.description && (
                                <p className="item-description">{invItem.item.description}</p>
                            )}
                            
                            <div className="item-properties">
                                {Object.entries(invItem.item.properties || {}).map(([key, value]) => (
                                    <span key={key} className="property">
                                        {key}: {typeof value === 'boolean' ? (value ? 'Yes' : 'No') : value}
                                    </span>
                                ))}
                            </div>
                            
                            <div className="item-actions">
                                {(invItem.item.type === 'weapon' || invItem.item.type === 'armor') && (
                                    <button 
                                        onClick={() => handleEquip(invItem.item_id, invItem.equipped)}
                                        className={invItem.equipped ? 'equipped' : ''}
                                    >
                                        {invItem.equipped ? 'Unequip' : 'Equip'}
                                    </button>
                                )}
                                
                                {invItem.item.requires_attunement && (
                                    <button 
                                        onClick={() => handleAttune(invItem.item_id, invItem.attuned)}
                                        className={invItem.attuned ? 'attuned' : ''}
                                        disabled={!invItem.attuned && weight.attunement_slots_used >= weight.attunement_slots_max}
                                    >
                                        {invItem.attuned ? 'End Attunement' : 'Attune'}
                                    </button>
                                )}
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};

export default InventoryView;