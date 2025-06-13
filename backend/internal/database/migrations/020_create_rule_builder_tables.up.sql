-- Rule Templates table for storing visual logic creations
CREATE TABLE IF NOT EXISTS rule_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL, -- spell, ability, item, environmental, condition
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_public BOOLEAN DEFAULT FALSE,
    version INTEGER DEFAULT 1,
    logic_graph JSONB NOT NULL,
    parameters JSONB DEFAULT '[]',
    balance_metrics JSONB DEFAULT '{}',
    conditional_rules JSONB DEFAULT '[]',
    tags TEXT[] DEFAULT '{}',
    usage_count INTEGER DEFAULT 0,
    approval_status VARCHAR(50) DEFAULT 'pending',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for rule_templates
CREATE INDEX idx_rule_templates_category ON rule_templates(category);
CREATE INDEX idx_rule_templates_created_by ON rule_templates(created_by);
CREATE INDEX idx_rule_templates_public ON rule_templates(is_public);
CREATE INDEX idx_rule_templates_approval ON rule_templates(approval_status);

-- Rule Instances table for active rules in play
CREATE TABLE IF NOT EXISTS rule_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES rule_templates(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL, -- Character, item, or location ID
    owner_type VARCHAR(50) NOT NULL, -- character, item, location, session
    session_id UUID REFERENCES game_sessions(id) ON DELETE CASCADE,
    parameter_values JSONB DEFAULT '{}',
    active_conditions TEXT[] DEFAULT '{}',
    state JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    activated_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for rule_instances
CREATE INDEX idx_rule_instances_template ON rule_instances(template_id);
CREATE INDEX idx_rule_instances_owner ON rule_instances(owner_id, owner_type);
CREATE INDEX idx_rule_instances_session ON rule_instances(session_id);
CREATE INDEX idx_rule_instances_active ON rule_instances(is_active);

-- Balance Simulations table for tracking AI balance analysis
CREATE TABLE IF NOT EXISTS balance_simulations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES rule_templates(id) ON DELETE CASCADE,
    simulation_type VARCHAR(50) NOT NULL,
    parameters JSONB NOT NULL,
    results JSONB NOT NULL,
    suggestions JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for balance_simulations
CREATE INDEX idx_balance_simulations_template ON balance_simulations(template_id);
CREATE INDEX idx_balance_simulations_created ON balance_simulations(created_at);

-- Rule Library table for community-shared rules
CREATE TABLE IF NOT EXISTS rule_library (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES rule_templates(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    author_id UUID NOT NULL REFERENCES users(id),
    category VARCHAR(50) NOT NULL,
    tags TEXT[] DEFAULT '{}',
    rating DECIMAL(3,2) DEFAULT 0,
    rating_count INTEGER DEFAULT 0,
    download_count INTEGER DEFAULT 0,
    featured BOOLEAN DEFAULT FALSE,
    published_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for rule_library
CREATE INDEX idx_rule_library_category ON rule_library(category);
CREATE INDEX idx_rule_library_author ON rule_library(author_id);
CREATE INDEX idx_rule_library_rating ON rule_library(rating DESC);
CREATE INDEX idx_rule_library_featured ON rule_library(featured);

-- Rule Ratings table
CREATE TABLE IF NOT EXISTS rule_ratings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    library_entry_id UUID NOT NULL REFERENCES rule_library(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    review TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(library_entry_id, user_id)
);

-- Conditional Context table for tracking active conditions
CREATE TABLE IF NOT EXISTS conditional_contexts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    context_type VARCHAR(50) NOT NULL, -- plane, emotion, environment, etc.
    context_value JSONB NOT NULL,
    affected_entities JSONB DEFAULT '[]', -- Characters, locations affected
    is_active BOOLEAN DEFAULT TRUE,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'
);

-- Create indexes for conditional_contexts
CREATE INDEX idx_conditional_contexts_session ON conditional_contexts(session_id);
CREATE INDEX idx_conditional_contexts_type ON conditional_contexts(context_type);
CREATE INDEX idx_conditional_contexts_active ON conditional_contexts(is_active);

-- Rule Execution Log for debugging and analysis
CREATE TABLE IF NOT EXISTS rule_execution_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID NOT NULL REFERENCES rule_instances(id) ON DELETE CASCADE,
    execution_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    trigger_type VARCHAR(100),
    input_values JSONB,
    output_values JSONB,
    nodes_executed JSONB,
    execution_duration_ms INTEGER,
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT
);

-- Create indexes for rule_execution_log
CREATE INDEX idx_rule_execution_instance ON rule_execution_log(instance_id);
CREATE INDEX idx_rule_execution_time ON rule_execution_log(execution_time);
CREATE INDEX idx_rule_execution_success ON rule_execution_log(success);

-- Node Templates table for reusable node configurations
CREATE TABLE IF NOT EXISTS node_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    node_type VARCHAR(50) NOT NULL,
    subtype VARCHAR(50),
    default_properties JSONB DEFAULT '{}',
    input_ports JSONB DEFAULT '[]',
    output_ports JSONB DEFAULT '[]',
    icon VARCHAR(50),
    color VARCHAR(7), -- Hex color
    category VARCHAR(50),
    is_system BOOLEAN DEFAULT FALSE, -- System nodes vs user-created
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for node_templates
CREATE INDEX idx_node_templates_type ON node_templates(node_type);
CREATE INDEX idx_node_templates_category ON node_templates(category);

-- Version History table for rule templates
CREATE TABLE IF NOT EXISTS rule_template_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES rule_templates(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    logic_graph JSONB NOT NULL,
    parameters JSONB,
    change_notes TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(template_id, version)
);

-- Create indexes for rule_template_versions
CREATE INDEX idx_rule_versions_template ON rule_template_versions(template_id);

-- Create update triggers
CREATE OR REPLACE FUNCTION update_rule_builder_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_rule_templates_updated_at
    BEFORE UPDATE ON rule_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_rule_builder_updated_at();

-- Insert default node templates
INSERT INTO node_templates (name, node_type, subtype, default_properties, input_ports, output_ports, icon, color, category, is_system) VALUES
-- Triggers
('On Action', 'trigger_action', 'action', '{"action_types": ["attack", "cast_spell", "move"]}', '[]', '[{"id": "out", "name": "trigger", "data_type": "trigger"}]', 'bolt', '#ff6b6b', 'triggers', true),
('On Damage', 'trigger_damage', 'damage', '{"damage_types": [], "threshold": 0}', '[]', '[{"id": "out", "name": "trigger", "data_type": "trigger"}, {"id": "damage", "name": "damage_amount", "data_type": "number"}]', 'heart-broken', '#ff6b6b', 'triggers', true),
('Every Turn', 'trigger_time', 'turn', '{"whose_turn": "self", "phase": "start"}', '[]', '[{"id": "out", "name": "trigger", "data_type": "trigger"}]', 'clock', '#4ecdc4', 'triggers', true),

-- Conditions
('If/Else', 'condition_check', 'boolean', '{}', '[{"id": "condition", "name": "condition", "data_type": "boolean"}]', '[{"id": "true", "name": "if true", "data_type": "any"}, {"id": "false", "name": "if false", "data_type": "any"}]', 'code-branch', '#f7b731', 'conditions', true),
('Compare Numbers', 'condition_compare', 'number', '{"operator": ">"}', '[{"id": "a", "name": "value A", "data_type": "number"}, {"id": "b", "name": "value B", "data_type": "number"}]', '[{"id": "result", "name": "result", "data_type": "boolean"}]', 'balance-scale', '#f7b731', 'conditions', true),
('Ability Check', 'condition_roll', 'check', '{"ability": "strength", "dc": 15}', '[{"id": "target", "name": "target", "data_type": "entity"}]', '[{"id": "success", "name": "success", "data_type": "boolean"}, {"id": "roll", "name": "roll_total", "data_type": "number"}]', 'dice-d20', '#f7b731', 'conditions', true),

-- Actions
('Deal Damage', 'action_damage', 'damage', '{"damage_dice": "1d6", "damage_type": "fire"}', '[{"id": "target", "name": "target", "data_type": "entity"}, {"id": "amount", "name": "damage", "data_type": "number"}]', '[{"id": "out", "name": "continue", "data_type": "any"}]', 'sword', '#ee5a24', 'actions', true),
('Apply Effect', 'action_effect', 'effect', '{"effect_type": "condition", "duration": "1_turn"}', '[{"id": "target", "name": "target", "data_type": "entity"}]', '[{"id": "out", "name": "continue", "data_type": "any"}]', 'magic', '#a55eea', 'actions', true),
('Modify Resource', 'action_resource', 'resource', '{"resource": "spell_slots", "operation": "subtract"}', '[{"id": "target", "name": "target", "data_type": "entity"}, {"id": "amount", "name": "amount", "data_type": "number"}]', '[{"id": "out", "name": "continue", "data_type": "any"}]', 'database', '#26de81', 'actions', true),

-- Calculations
('Math Operation', 'calc_math', 'math', '{"operation": "+"}', '[{"id": "a", "name": "value A", "data_type": "number"}, {"id": "b", "name": "value B", "data_type": "number"}]', '[{"id": "result", "name": "result", "data_type": "number"}]', 'calculator', '#0fb9b1', 'calculations', true),
('Roll Dice', 'calc_random', 'dice', '{"dice_notation": "1d20"}', '[]', '[{"id": "result", "name": "result", "data_type": "number"}]', 'dice', '#0fb9b1', 'calculations', true),
('Get Property', 'calc_property', 'property', '{"property_path": "abilities.strength.modifier"}', '[{"id": "entity", "name": "entity", "data_type": "entity"}]', '[{"id": "value", "name": "value", "data_type": "any"}]', 'tag', '#0fb9b1', 'calculations', true);