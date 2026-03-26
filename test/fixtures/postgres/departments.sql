INSERT INTO departments (id, org_id, name) VALUES
                                               ('11111111-0000-0000-0000-000000000001', '7eef9e70-5ec9-48f5-a451-dd9901870071', 'Engineering'),
                                               ('11111111-0000-0000-0000-000000000002', '7eef9e70-5ec9-48f5-a451-dd9901870071', 'Human Resources'),
                                               ('11111111-0000-0000-0000-000000000003', '7eef9e70-5ec9-48f5-a451-dd9901870071', 'Finance'),
                                               ('11111111-0000-0000-0000-000000000004', '7eef9e70-5ec9-48f5-a451-dd9901870071', 'Marketing')
    ON CONFLICT (id) DO NOTHING;