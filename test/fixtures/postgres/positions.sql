INSERT INTO positions (id, org_id, name) VALUES
                                             ('22222222-0000-0000-0000-000000000001', '7eef9e70-5ec9-48f5-a451-dd9901870071', 'Software Engineer'),
                                             ('22222222-0000-0000-0000-000000000002', '7eef9e70-5ec9-48f5-a451-dd9901870071', 'Senior Software Engineer'),
                                             ('22222222-0000-0000-0000-000000000003', '7eef9e70-5ec9-48f5-a451-dd9901870071', 'HR Manager'),
                                             ('22222222-0000-0000-0000-000000000004', '7eef9e70-5ec9-48f5-a451-dd9901870071', 'Financial Analyst'),
                                             ('22222222-0000-0000-0000-000000000005', '7eef9e70-5ec9-48f5-a451-dd9901870071', 'Marketing Specialist')
    ON CONFLICT (id) DO NOTHING;
 