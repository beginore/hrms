INSERT INTO departments (id, org_id, name) VALUES
                                               ('11111111-0000-0000-0000-000000000001', '6ca38f34-e671-4a4e-a293-c4f80c109b87', 'Engineering'),
                                               ('11111111-0000-0000-0000-000000000002', '6ca38f34-e671-4a4e-a293-c4f80c109b87', 'Human Resources'),
                                               ('11111111-0000-0000-0000-000000000003', '6ca38f34-e671-4a4e-a293-c4f80c109b87', 'Finance'),
                                               ('11111111-0000-0000-0000-000000000004', '6ca38f34-e671-4a4e-a293-c4f80c109b87', 'Marketing')
    ON CONFLICT (name) DO NOTHING;
