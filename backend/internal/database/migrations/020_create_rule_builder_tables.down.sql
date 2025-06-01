-- Drop triggers
DROP TRIGGER IF EXISTS update_rule_templates_updated_at ON rule_templates;

-- Drop function
DROP FUNCTION IF EXISTS update_rule_builder_updated_at();

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS rule_template_versions;
DROP TABLE IF EXISTS node_templates;
DROP TABLE IF EXISTS rule_execution_log;
DROP TABLE IF EXISTS conditional_contexts;
DROP TABLE IF EXISTS rule_ratings;
DROP TABLE IF EXISTS rule_library;
DROP TABLE IF EXISTS balance_simulations;
DROP TABLE IF EXISTS rule_instances;
DROP TABLE IF EXISTS rule_templates;