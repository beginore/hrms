INSERT INTO documents (id, type, version, url, is_active, created_at)
VALUES
    (gen_random_uuid(), 'PRIVACY_POLICY', '0.9', 'https://s3.amazonaws.com/formedic/privacy_policy_v1.0.pdf', true, NOW()),
    (gen_random_uuid(), 'TERMS_AND_CONDITIONS', '0.9', 'https://s3.amazonaws.com/formedic/terms_v1.0.pdf', true, NOW());
