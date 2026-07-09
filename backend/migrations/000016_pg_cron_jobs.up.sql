DO $$BEGIN CREATE EXTENSION IF NOT EXISTS pg_cron; EXCEPTION WHEN OTHERS THEN NULL; END$$;
CREATE OR REPLACE FUNCTION _schedule_cleanup_jobs() RETURNS void AS $$
DECLARE
    jobs text[] := ARRAY[
        'delete-expired-refresh-tokens|0 * * * *|DELETE FROM refresh_tokens WHERE expires_at < NOW()',
        'delete-expired-verification-tokens|0 * * * *|DELETE FROM verification_tokens WHERE expires_at < NOW()',
        'delete-expired-oauth-states|0 * * * *|DELETE FROM oauth_states WHERE expires_at < NOW()',
        'delete-expired-revoked-tokens|0 * * * *|DELETE FROM revoked_tokens WHERE expires_at < NOW()',
        'delete-expired-invitations|0 0 * * *|DELETE FROM invitations WHERE expires_at < NOW() AND status = ''pending''',
        'delete-expired-webauthn-sessions|0 * * * *|DELETE FROM webauthn_sessions WHERE expires_at < NOW()',
        'delete-expired-auth-codes|0 * * * *|DELETE FROM auth_codes WHERE expires_at < NOW()',
        'delete-old-audit-logs|0 0 * * *|DELETE FROM audit_log WHERE created_at < NOW() - INTERVAL ''90 days''',
        'delete-old-system-metrics|0 0 * * *|DELETE FROM system_metrics WHERE timestamp < NOW() - INTERVAL ''30 days''',
        'delete-old-webhook-deliveries|0 0 * * *|DELETE FROM webhook_deliveries WHERE status = ''delivered'' AND created_at < NOW() - INTERVAL ''30 days'''
    ];
    job text; parts text[];
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_cron') THEN
        FOREACH job IN ARRAY jobs LOOP
            parts := string_to_array(job, '|');
            BEGIN PERFORM cron.schedule(parts[1], parts[2], parts[3]); EXCEPTION WHEN OTHERS THEN NULL; END;
        END LOOP;
    END IF;
END;
$$ LANGUAGE plpgsql;
SELECT _schedule_cleanup_jobs();
DROP FUNCTION _schedule_cleanup_jobs();
