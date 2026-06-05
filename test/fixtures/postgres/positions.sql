INSERT INTO positions (id, org_id, name) VALUES
                                             ('22222222-0000-0000-0000-000000000001', '6ca38f34-e671-4a4e-a293-c4f80c109b87', 'Software Engineer'),
                                             ('22222222-0000-0000-0000-000000000002', '6ca38f34-e671-4a4e-a293-c4f80c109b87', 'Senior Software Engineer'),
                                             ('22222222-0000-0000-0000-000000000003', '6ca38f34-e671-4a4e-a293-c4f80c109b87', 'HR Manager'),
                                             ('22222222-0000-0000-0000-000000000004', '6ca38f34-e671-4a4e-a293-c4f80c109b87', 'Financial Analyst'),
                                             ('22222222-0000-0000-0000-000000000005', '6ca38f34-e671-4a4e-a293-c4f80c109b87', 'Marketing Specialist')
    ON CONFLICT (name) DO NOTHING;
 
