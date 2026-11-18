CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_delivery_events_failed_branch_occurred ON delivery_events (branch_id, occurred_at) WHERE status = 'failed';
