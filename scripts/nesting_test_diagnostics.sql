\set ON_ERROR_STOP on
BEGIN READ ONLY;
SET LOCAL statement_timeout='2s';
SET LOCAL lock_timeout='200ms';
DO $$ BEGIN
 IF current_database()<>'nesting_test' OR current_schema() !~ '^test_[0-9]+$' THEN
  RAISE EXCEPTION 'diagnostics require nesting_test and an explicitly selected test schema';
 END IF;
END $$;
SELECT current_database(),current_user,postgis_version();
SELECT 'save' AS kind,state,count(*) FROM nesting_operations GROUP BY state
UNION ALL SELECT 'preview',state,count(*) FROM nesting_preview_operations GROUP BY state;
SELECT count(*) AS ready_rows,coalesce(sum(pg_column_size(payload)),0) AS payload_bytes,
 coalesce(sum(pg_column_size(saved_result)),0) AS saved_result_bytes FROM nesting_previews;
SELECT relation,pg_total_relation_size(relation::regclass) AS total_bytes,
 pg_indexes_size(relation::regclass) AS index_bytes
FROM (VALUES ('nesting_operations'),('nesting_preview_operations'),('nesting_previews')) AS names(relation);
SELECT state,count(*) FROM pg_stat_activity
WHERE datname=current_database() AND
 (application_name LIKE 'stage5_%' OR application_name LIKE 'stage7_%') GROUP BY state;
ROLLBACK;
