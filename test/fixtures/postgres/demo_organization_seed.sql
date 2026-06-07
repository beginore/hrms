BEGIN;

CREATE OR REPLACE FUNCTION pg_temp.demo_uuid(seed text)
RETURNS uuid
LANGUAGE sql
AS $$
SELECT (
    substr(md5(seed), 1, 8) || '-' ||
    substr(md5(seed), 9, 4) || '-4' ||
    substr(md5(seed), 14, 3) || '-8' ||
    substr(md5(seed), 18, 3) || '-' ||
    substr(md5(seed), 21, 12)
)::uuid;
$$;

CREATE TEMP TABLE demo_people (
    email text PRIMARY KEY,
    role text NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    phone_number text NOT NULL,
    department_name text NOT NULL,
    position_name text NOT NULL,
    salary_rate numeric(10, 2) NOT NULL,
    employee_status text NOT NULL,
    hire_date date NOT NULL,
    work_start time NOT NULL DEFAULT '09:00',
    work_end time NOT NULL DEFAULT '18:00'
) ON COMMIT DROP;

INSERT INTO demo_people (
    email, role, first_name, last_name, phone_number,
    department_name, position_name, salary_rate, employee_status, hire_date, work_start, work_end
) VALUES
    ('testadmin@mail.ru', 'SysAdmin', 'Test', 'Admin', '+3920414261', 'Demo Executive', 'Demo CEO', 1500000, 'Active', '2026-01-10', '09:00', '18:00'),
    ('testmanager@mail.ru', 'Admin', 'Test', 'Manager', '+77474313223', 'Demo Engineering', 'Demo Engineering Manager', 950000, 'Active', '2026-02-01', '09:00', '18:00'),
    ('testemployee@mail.ru', 'Employee', 'Test', 'Employee', '+77474313123', 'Demo Engineering', 'Demo QA Engineer', 420000, 'Active', '2026-03-05', '09:00', '18:00'),
    ('hr.manager@demo.hrms.local', 'HR', 'Aliya', 'Nurgaliyeva', '+77010001001', 'Demo HR', 'Demo HR Manager', 780000, 'Active', '2026-01-15', '09:00', '18:00'),
    ('finance.admin@demo.hrms.local', 'Admin', 'Dina', 'Sarsenova', '+77010001002', 'Demo Finance', 'Demo Finance Admin', 860000, 'Active', '2026-01-20', '09:00', '18:00'),
    ('backend.one@demo.hrms.local', 'Employee', 'Timur', 'Kassenov', '+77010001003', 'Demo Engineering', 'Demo Backend Developer', 820000, 'Active', '2026-02-10', '09:00', '18:00'),
    ('backend.two@demo.hrms.local', 'Employee', 'Arman', 'Suleimenov', '+77010001004', 'Demo Engineering', 'Demo Backend Developer', 760000, 'Active', '2026-02-12', '10:00', '19:00'),
    ('frontend.one@demo.hrms.local', 'Employee', 'Dana', 'Akhmetova', '+77010001005', 'Demo Engineering', 'Demo Frontend Developer', 690000, 'Active', '2026-02-18', '09:00', '18:00'),
    ('devops.one@demo.hrms.local', 'Employee', 'Nurlan', 'Iskakov', '+77010001006', 'Demo Engineering', 'Demo DevOps Engineer', 880000, 'Active', '2026-02-20', '10:00', '19:00'),
    ('sales.one@demo.hrms.local', 'Employee', 'Madina', 'Yessenova', '+77010001007', 'Demo Sales', 'Demo Sales Representative', 610000, 'Active', '2026-03-01', '09:00', '18:00'),
    ('sales.two@demo.hrms.local', 'Employee', 'Zhanar', 'Beketova', '+77010001008', 'Demo Sales', 'Demo Sales Representative', 590000, 'Active', '2026-03-03', '09:00', '18:00'),
    ('accountant.one@demo.hrms.local', 'Employee', 'Aigerim', 'Omarova', '+77010001009', 'Demo Finance', 'Demo Accountant', 640000, 'Active', '2026-03-08', '09:00', '18:00'),
    ('ops.manager@demo.hrms.local', 'Admin', 'Rustem', 'Karimov', '+77010001010', 'Demo Operations', 'Demo Ops Manager', 830000, 'Active', '2026-02-05', '08:30', '17:30'),
    ('ops.one@demo.hrms.local', 'Employee', 'Miras', 'Tulegenov', '+77010001011', 'Demo Operations', 'Demo Operations Specialist', 560000, 'Active', '2026-03-10', '08:30', '17:30'),
    ('hr.specialist@demo.hrms.local', 'Employee', 'Aruzhan', 'Muratova', '+77010001012', 'Demo HR', 'Demo HR Specialist', 540000, 'Active', '2026-03-12', '09:00', '18:00'),
    ('qa.two@demo.hrms.local', 'Employee', 'Bekzat', 'Serikov', '+77010001013', 'Demo Engineering', 'Demo QA Engineer', 480000, 'Active', '2026-03-15', '09:00', '18:00');

INSERT INTO departments (id, org_id, name)
VALUES
    (pg_temp.demo_uuid('department:demo-executive'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo Executive'),
    (pg_temp.demo_uuid('department:demo-engineering'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo Engineering'),
    (pg_temp.demo_uuid('department:demo-hr'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo HR'),
    (pg_temp.demo_uuid('department:demo-finance'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo Finance'),
    (pg_temp.demo_uuid('department:demo-sales'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo Sales'),
    (pg_temp.demo_uuid('department:demo-operations'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo Operations')
ON CONFLICT (id) DO UPDATE
SET org_id = EXCLUDED.org_id,
    name = EXCLUDED.name;

INSERT INTO positions (id, org_id, name)
VALUES
    (pg_temp.demo_uuid('position:demo-ceo'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo CEO'),
    (pg_temp.demo_uuid('position:demo-engineering-manager'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo Engineering Manager'),
    (pg_temp.demo_uuid('position:demo-hr-manager'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo HR Manager'),
    (pg_temp.demo_uuid('position:demo-finance-admin'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo Finance Admin'),
    (pg_temp.demo_uuid('position:demo-backend-developer'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo Backend Developer'),
    (pg_temp.demo_uuid('position:demo-frontend-developer'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo Frontend Developer'),
    (pg_temp.demo_uuid('position:demo-qa-engineer'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo QA Engineer'),
    (pg_temp.demo_uuid('position:demo-devops-engineer'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo DevOps Engineer'),
    (pg_temp.demo_uuid('position:demo-sales-representative'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo Sales Representative'),
    (pg_temp.demo_uuid('position:demo-accountant'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo Accountant'),
    (pg_temp.demo_uuid('position:demo-ops-manager'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo Ops Manager'),
    (pg_temp.demo_uuid('position:demo-operations-specialist'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo Operations Specialist'),
    (pg_temp.demo_uuid('position:demo-hr-specialist'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'Demo HR Specialist')
ON CONFLICT (id) DO UPDATE
SET org_id = EXCLUDED.org_id,
    name = EXCLUDED.name;

INSERT INTO users (id, org_id, email, role, first_name, last_name, phone_number, verification_status, created_at)
SELECT
    CASE
        WHEN p.email = 'testadmin@mail.ru' THEN 'a498a4b8-3091-70b3-880d-1a6cf0f3b6f0'::uuid
        WHEN p.email = 'testmanager@mail.ru' THEN '94a86498-5041-7089-d728-d8764907de8e'::uuid
        WHEN p.email = 'testemployee@mail.ru' THEN '944804c8-b031-700a-a536-a31f91ab5236'::uuid
        ELSE pg_temp.demo_uuid('user:' || p.email)
    END,
    '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2',
    p.email,
    p.role,
    p.first_name,
    p.last_name,
    p.phone_number,
    'Verified',
    p.hire_date::timestamptz + time '09:00'
FROM demo_people p
ON CONFLICT (email) DO UPDATE
SET org_id = EXCLUDED.org_id,
    role = EXCLUDED.role,
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    phone_number = EXCLUDED.phone_number,
    verification_status = EXCLUDED.verification_status;

UPDATE employees e
SET department_id = d.id,
    position_id = pos.id,
    role = p.role,
    salary_rate = p.salary_rate,
    status = p.employee_status,
    hire_date = p.hire_date
FROM demo_people p
JOIN users u ON u.email = p.email
JOIN departments d ON d.name = p.department_name
JOIN positions pos ON pos.name = p.position_name
WHERE e.org_id = '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2'
  AND e.user_id = u.id
  AND p.role <> 'SysAdmin';

INSERT INTO employees (id, org_id, user_id, department_id, position_id, role, salary_rate, status, hire_date)
SELECT
    pg_temp.demo_uuid('employee:' || p.email),
    '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2',
    u.id,
    d.id,
    pos.id,
    p.role,
    p.salary_rate,
    p.employee_status,
    p.hire_date
FROM demo_people p
JOIN users u ON u.email = p.email
JOIN departments d ON d.name = p.department_name
JOIN positions pos ON pos.name = p.position_name
WHERE p.role <> 'SysAdmin'
  AND NOT EXISTS (
    SELECT 1
    FROM employees e
    WHERE e.org_id = '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2'
      AND e.user_id = u.id
);

INSERT INTO work_schedules (id, org_id, employee_id, work_start, work_end, late_threshold_minutes)
SELECT
    pg_temp.demo_uuid('schedule:' || p.email),
    '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2',
    e.id,
    p.work_start,
    p.work_end,
    15
FROM demo_people p
JOIN users u ON u.email = p.email
JOIN employees e ON e.user_id = u.id AND e.org_id = '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2'
ON CONFLICT (employee_id) DO UPDATE
SET work_start = EXCLUDED.work_start,
    work_end = EXCLUDED.work_end,
    late_threshold_minutes = EXCLUDED.late_threshold_minutes,
    updated_at = now();

INSERT INTO working_calendar (id, org_id, date, type, name)
SELECT
    pg_temp.demo_uuid('calendar:3db21a6d-abca-49a5-8f56-f3cbbc0f22e2:' || d::date),
    '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2',
    d::date,
    CASE
        WHEN d::date IN ('2026-05-01', '2026-05-07', '2026-05-09') THEN 'HOLIDAY'
        WHEN extract(isodow FROM d) IN (6, 7) THEN 'WEEKEND'
        ELSE 'WORKDAY'
    END,
    CASE
        WHEN d::date = '2026-05-01' THEN 'Unity Day'
        WHEN d::date = '2026-05-07' THEN 'Defender of the Fatherland Day'
        WHEN d::date = '2026-05-09' THEN 'Victory Day'
        WHEN extract(isodow FROM d) IN (6, 7) THEN 'Weekend'
        ELSE 'Regular workday'
    END
FROM generate_series('2026-05-01'::date, '2026-05-31'::date, interval '1 day') AS d
ON CONFLICT (org_id, date) DO UPDATE
SET type = EXCLUDED.type,
    name = EXCLUDED.name;

INSERT INTO attendance_records (id, org_id, employee_id, date, type, source, status, check_in, check_out, note)
SELECT
    pg_temp.demo_uuid('attendance:' || u.email || ':' || c.date),
    '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2',
    e.id,
    c.date,
    CASE
        WHEN u.email = 'sales.two@demo.hrms.local' AND c.date BETWEEN '2026-05-20' AND '2026-05-22' THEN 'VACATION'
        WHEN u.email = 'frontend.one@demo.hrms.local' AND c.date BETWEEN '2026-05-14' AND '2026-05-15' THEN 'SICK_LEAVE'
        WHEN u.email IN ('hr.manager@demo.hrms.local', 'backend.two@demo.hrms.local') AND c.date IN ('2026-05-18', '2026-05-19') THEN 'REMOTE'
        ELSE 'OFFICE'
    END,
    CASE
        WHEN u.email IN ('hr.manager@demo.hrms.local', 'backend.two@demo.hrms.local') AND c.date IN ('2026-05-18', '2026-05-19') THEN 'MANUAL'
        WHEN u.email = 'sales.two@demo.hrms.local' AND c.date BETWEEN '2026-05-20' AND '2026-05-22' THEN 'SYSTEM'
        WHEN u.email = 'frontend.one@demo.hrms.local' AND c.date BETWEEN '2026-05-14' AND '2026-05-15' THEN 'SYSTEM'
        ELSE 'SKUD'
    END,
    CASE
        WHEN u.email = 'qa.two@demo.hrms.local' AND c.date = '2026-05-13' THEN 'ABSENT'
        WHEN u.email = 'sales.two@demo.hrms.local' AND c.date BETWEEN '2026-05-20' AND '2026-05-22' THEN 'ON_LEAVE'
        WHEN u.email = 'frontend.one@demo.hrms.local' AND c.date BETWEEN '2026-05-14' AND '2026-05-15' THEN 'ON_LEAVE'
        WHEN extract(day FROM c.date)::int IN (6, 12, 21, 27)
             AND u.email IN ('backend.one@demo.hrms.local', 'testemployee@mail.ru', 'ops.one@demo.hrms.local') THEN 'LATE'
        ELSE 'PRESENT'
    END,
    CASE
        WHEN u.email = 'qa.two@demo.hrms.local' AND c.date = '2026-05-13' THEN NULL
        WHEN u.email = 'sales.two@demo.hrms.local' AND c.date BETWEEN '2026-05-20' AND '2026-05-22' THEN NULL
        WHEN u.email = 'frontend.one@demo.hrms.local' AND c.date BETWEEN '2026-05-14' AND '2026-05-15' THEN NULL
        WHEN extract(day FROM c.date)::int IN (6, 12, 21, 27)
             AND u.email IN ('backend.one@demo.hrms.local', 'testemployee@mail.ru', 'ops.one@demo.hrms.local')
            THEN c.date::timestamp + ws.work_start + interval '27 minutes'
        ELSE c.date::timestamp + ws.work_start + (extract(day FROM c.date)::int % 8) * interval '1 minute'
    END,
    CASE
        WHEN u.email = 'qa.two@demo.hrms.local' AND c.date = '2026-05-13' THEN NULL
        WHEN u.email = 'sales.two@demo.hrms.local' AND c.date BETWEEN '2026-05-20' AND '2026-05-22' THEN NULL
        WHEN u.email = 'frontend.one@demo.hrms.local' AND c.date BETWEEN '2026-05-14' AND '2026-05-15' THEN NULL
        ELSE c.date::timestamp + ws.work_end + (extract(day FROM c.date)::int % 6) * interval '5 minutes'
    END,
    CASE
        WHEN u.email = 'qa.two@demo.hrms.local' AND c.date = '2026-05-13' THEN 'No show, manager follow-up required'
        WHEN u.email = 'sales.two@demo.hrms.local' AND c.date BETWEEN '2026-05-20' AND '2026-05-22' THEN 'Approved vacation'
        WHEN u.email = 'frontend.one@demo.hrms.local' AND c.date BETWEEN '2026-05-14' AND '2026-05-15' THEN 'Approved sick leave'
        WHEN u.email IN ('hr.manager@demo.hrms.local', 'backend.two@demo.hrms.local') AND c.date IN ('2026-05-18', '2026-05-19') THEN 'Remote work day'
        ELSE NULL
    END
FROM users u
JOIN employees e ON e.user_id = u.id AND e.org_id = '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2'
JOIN work_schedules ws ON ws.employee_id = e.id
JOIN working_calendar c ON c.org_id = e.org_id AND c.type = 'WORKDAY'
WHERE u.org_id = '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2'
  AND c.date BETWEEN '2026-05-01' AND '2026-05-31'
ON CONFLICT (employee_id, date) DO UPDATE
SET type = EXCLUDED.type,
    source = EXCLUDED.source,
    status = EXCLUDED.status,
    check_in = EXCLUDED.check_in,
    check_out = EXCLUDED.check_out,
    note = EXCLUDED.note,
    updated_at = now();

INSERT INTO skud_events (id, org_id, employee_id, event_type, device_id, occurred_at, processed)
SELECT
    pg_temp.demo_uuid('skud-enter:' || ar.employee_id || ':' || ar.date),
    ar.org_id,
    ar.employee_id,
    'ENTER',
    'ASTANA-HQ-01',
    ar.check_in,
    true
FROM attendance_records ar
WHERE ar.org_id = '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2'
  AND ar.source = 'SKUD'
  AND ar.check_in IS NOT NULL
ON CONFLICT (id) DO UPDATE
SET occurred_at = EXCLUDED.occurred_at,
    processed = EXCLUDED.processed;

INSERT INTO skud_events (id, org_id, employee_id, event_type, device_id, occurred_at, processed)
SELECT
    pg_temp.demo_uuid('skud-exit:' || ar.employee_id || ':' || ar.date),
    ar.org_id,
    ar.employee_id,
    'EXIT',
    'ASTANA-HQ-01',
    ar.check_out,
    true
FROM attendance_records ar
WHERE ar.org_id = '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2'
  AND ar.source = 'SKUD'
  AND ar.check_out IS NOT NULL
ON CONFLICT (id) DO UPDATE
SET occurred_at = EXCLUDED.occurred_at,
    processed = EXCLUDED.processed;

INSERT INTO leave_requests (id, org_id, employee_id, type, start_date, end_date, reason, status, reviewed_by, reviewed_at)
SELECT *
FROM (
    SELECT
        pg_temp.demo_uuid('leave:sales-two:vacation') AS id,
        '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2'::uuid AS org_id,
        e.id AS employee_id,
        'VACATION'::varchar AS type,
        '2026-05-20'::date AS start_date,
        '2026-05-22'::date AS end_date,
        'Family trip'::text AS reason,
        'APPROVED'::varchar AS status,
        reviewer.id AS reviewed_by,
        '2026-05-15 14:00+05'::timestamptz AS reviewed_at
    FROM employees e
    JOIN users u ON u.id = e.user_id
    JOIN users reviewer ON reviewer.email = 'testmanager@mail.ru'
    WHERE u.email = 'sales.two@demo.hrms.local'
    UNION ALL
    SELECT
        pg_temp.demo_uuid('leave:frontend-one:sick'),
        '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2'::uuid,
        e.id,
        'SICK_LEAVE',
        '2026-05-14',
        '2026-05-15',
        'Medical certificate provided',
        'APPROVED',
        reviewer.id,
        '2026-05-14 10:30+05'::timestamptz
    FROM employees e
    JOIN users u ON u.id = e.user_id
    JOIN users reviewer ON reviewer.email = 'hr.manager@demo.hrms.local'
    WHERE u.email = 'frontend.one@demo.hrms.local'
    UNION ALL
    SELECT
        pg_temp.demo_uuid('leave:backend-two:remote'),
        '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2'::uuid,
        e.id,
        'REMOTE',
        '2026-05-18',
        '2026-05-19',
        'Remote work for release support',
        'APPROVED',
        reviewer.id,
        '2026-05-17 16:00+05'::timestamptz
    FROM employees e
    JOIN users u ON u.id = e.user_id
    JOIN users reviewer ON reviewer.email = 'testmanager@mail.ru'
    WHERE u.email = 'backend.two@demo.hrms.local'
) rows
ON CONFLICT (id) DO UPDATE
SET status = EXCLUDED.status,
    reviewed_by = EXCLUDED.reviewed_by,
    reviewed_at = EXCLUDED.reviewed_at;

INSERT INTO events (id, title, description, starts_at, ends_at, scope, department_id, created_by, created_by_role, organization_id)
SELECT
    pg_temp.demo_uuid(seed),
    title,
    description,
    starts_at,
    ends_at,
    scope,
    department_id,
    created_by,
    created_by_role,
    organization_id
FROM (
    SELECT
        'event:allhands:2026-06-10' AS seed,
        'Monthly all-hands'::varchar AS title,
        'Company updates, payroll timeline, and roadmap review'::text AS description,
        '2026-06-10 10:00+05'::timestamptz AS starts_at,
        '2026-06-10 11:00+05'::timestamptz AS ends_at,
        'global'::varchar AS scope,
        NULL::uuid AS department_id,
        u.id AS created_by,
        u.role AS created_by_role,
        u.org_id AS organization_id
    FROM users u WHERE u.email = 'testadmin@mail.ru'
    UNION ALL
    SELECT
        'event:engineering-planning:2026-06-12',
        'Engineering sprint planning',
        'Plan the next HRMS payroll and attendance sprint',
        '2026-06-12 15:00+05',
        '2026-06-12 16:30+05',
        'department',
        d.id,
        u.id,
        u.role,
        u.org_id
    FROM users u
    JOIN departments d ON d.name = 'Demo Engineering'
    WHERE u.email = 'testmanager@mail.ru'
    UNION ALL
    SELECT
        'event:finance-payroll-review:2026-06-17',
        'Payroll review meeting',
        'Finance review for May payroll cycle',
        '2026-06-17 11:00+05',
        '2026-06-17 12:00+05',
        'department',
        d.id,
        u.id,
        u.role,
        u.org_id
    FROM users u
    JOIN departments d ON d.name = 'Demo Finance'
    WHERE u.email = 'finance.admin@demo.hrms.local'
) rows
ON CONFLICT (id) DO UPDATE
SET title = EXCLUDED.title,
    description = EXCLUDED.description,
    starts_at = EXCLUDED.starts_at,
    ends_at = EXCLUDED.ends_at,
    updated_at = now();

INSERT INTO tasks (id, org_id, created_by, assigned_to, title, description, due_date, status, reviewed_by, reviewed_at, report_description, submitted_at)
SELECT
    pg_temp.demo_uuid(seed),
    '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2',
    creator.id,
    assignee_employee.id,
    t.title,
    t.description,
    t.due_date,
    t.status,
    reviewer.id,
    t.reviewed_at,
    t.report_description,
    t.submitted_at
FROM (
    VALUES
        ('task:qa:regression', 'testmanager@mail.ru', 'testemployee@mail.ru', 'Regression test payroll endpoints', 'Run smoke and regression checks for payroll API', '2026-06-14'::date, 'SUBMITTED', 'testmanager@mail.ru', '2026-06-14 17:20+05'::timestamptz, 'Regression completed, two UI issues reported', '2026-06-14 16:45+05'::timestamptz),
        ('task:backend:attendance-api', 'testmanager@mail.ru', 'backend.one@demo.hrms.local', 'Attendance API cleanup', 'Normalize attendance status handling and errors', '2026-06-18', 'PENDING', NULL, NULL, NULL, NULL),
        ('task:frontend:payroll-ui', 'testmanager@mail.ru', 'frontend.one@demo.hrms.local', 'Payroll dashboard widgets', 'Add summary cards for gross, net, and employer taxes', '2026-06-20', 'PENDING', NULL, NULL, NULL, NULL),
        ('task:hr:onboarding', 'hr.manager@demo.hrms.local', 'hr.specialist@demo.hrms.local', 'Prepare onboarding checklist', 'Create June onboarding checklist for new hires', '2026-06-13', 'APPROVED', 'hr.manager@demo.hrms.local', '2026-06-13 15:10+05', 'Checklist approved and shared with managers', '2026-06-13 14:30+05'),
        ('task:finance:payslip-audit', 'finance.admin@demo.hrms.local', 'accountant.one@demo.hrms.local', 'Audit May payslips', 'Check tax and employer cost calculations for May', '2026-06-19', 'SUBMITTED', 'finance.admin@demo.hrms.local', '2026-06-19 18:00+05', 'Audit prepared for finance admin review', '2026-06-19 17:35+05'),
        ('task:ops:access-cards', 'ops.manager@demo.hrms.local', 'ops.one@demo.hrms.local', 'Check office access cards', 'Reconcile SKUD events with access card list', '2026-06-16', 'PENDING', NULL, NULL, NULL, NULL),
        ('task:sales:crm-cleanup', 'ops.manager@demo.hrms.local', 'sales.one@demo.hrms.local', 'Clean demo CRM pipeline', 'Prepare sales demo pipeline for HRMS presentation', '2026-06-21', 'PENDING', NULL, NULL, NULL, NULL)
) AS t(seed, creator_email, assignee_email, title, description, due_date, status, reviewer_email, reviewed_at, report_description, submitted_at)
JOIN users creator ON creator.email = t.creator_email
JOIN users assignee_user ON assignee_user.email = t.assignee_email
JOIN employees assignee_employee ON assignee_employee.user_id = assignee_user.id
LEFT JOIN users reviewer ON reviewer.email = t.reviewer_email
ON CONFLICT (id) DO UPDATE
SET title = EXCLUDED.title,
    description = EXCLUDED.description,
    due_date = EXCLUDED.due_date,
    status = EXCLUDED.status,
    reviewed_by = EXCLUDED.reviewed_by,
    reviewed_at = EXCLUDED.reviewed_at,
    report_description = EXCLUDED.report_description,
    submitted_at = EXCLUDED.submitted_at,
    updated_at = now();

INSERT INTO salary_history (id, org_id, employee_id, salary_rate, effective_from, effective_to, created_by)
SELECT
    pg_temp.demo_uuid('salary-history:' || u.email || ':2026-05'),
    e.org_id,
    e.id,
    e.salary_rate,
    '2026-05-01',
    NULL,
    admin.id
FROM employees e
JOIN users u ON u.id = e.user_id
JOIN users admin ON admin.email = 'testadmin@mail.ru'
WHERE e.org_id = '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2'
ON CONFLICT (id) DO UPDATE
SET salary_rate = EXCLUDED.salary_rate,
    effective_from = EXCLUDED.effective_from,
    created_by = EXCLUDED.created_by;

INSERT INTO payroll_policies (id, org_id, missing_attendance_policy, late_penalty_mode, late_penalty_amount, rounding_mode)
VALUES (
    pg_temp.demo_uuid('payroll-policy:3db21a6d-abca-49a5-8f56-f3cbbc0f22e2'),
    '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2',
    'PAID_REVIEW_REQUIRED',
    'FIXED',
    2500,
    'CENT'
)
ON CONFLICT (org_id) DO UPDATE
SET missing_attendance_policy = EXCLUDED.missing_attendance_policy,
    late_penalty_mode = EXCLUDED.late_penalty_mode,
    late_penalty_amount = EXCLUDED.late_penalty_amount,
    rounding_mode = EXCLUDED.rounding_mode,
    updated_at = now();

INSERT INTO payroll_tax_rules (id, org_id, country, name, rate, applies_to, is_active, effective_from, payer)
VALUES
    (pg_temp.demo_uuid('tax:pit:demo-org'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'KZ', 'Personal income tax', 0.10, 'GROSS', true, '2026-01-01', 'EMPLOYEE'),
    (pg_temp.demo_uuid('tax:social:demo-org'), '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2', 'KZ', 'Employer social tax', 0.095, 'GROSS', true, '2026-01-01', 'EMPLOYER')
ON CONFLICT (id) DO UPDATE
SET rate = EXCLUDED.rate,
    is_active = EXCLUDED.is_active,
    updated_at = now();

INSERT INTO payroll_cycles (id, org_id, period_start, period_end, status, created_by, approved_by, approved_at, paid_at, currency)
SELECT
    pg_temp.demo_uuid('payroll-cycle:demo-org:2026-05'),
    '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2',
    '2026-05-01',
    '2026-05-31',
    'READY_FOR_PAYMENT',
    admin.id,
    finance.id,
    '2026-06-03 15:00+05',
    NULL,
    'KZT'
FROM users admin
JOIN users finance ON finance.email = 'finance.admin@demo.hrms.local'
WHERE admin.email = 'testadmin@mail.ru'
ON CONFLICT (id) DO UPDATE
SET status = EXCLUDED.status,
    approved_by = EXCLUDED.approved_by,
    approved_at = EXCLUDED.approved_at,
    updated_at = now();

INSERT INTO payroll_adjustments (id, org_id, employee_id, cycle_id, type, category, amount, is_taxable, reason, created_by)
SELECT
    pg_temp.demo_uuid(seed),
    '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2',
    e.id,
    pg_temp.demo_uuid('payroll-cycle:demo-org:2026-05'),
    type,
    category,
    amount,
    is_taxable,
    reason,
    creator.id
FROM (
    VALUES
        ('adjustment:backend-one:bonus', 'backend.one@demo.hrms.local', 'BONUS', 'PERFORMANCE', 50000::numeric, true, 'Release support bonus', 'testmanager@mail.ru'),
        ('adjustment:sales-one:bonus', 'sales.one@demo.hrms.local', 'BONUS', 'SALES', 45000::numeric, true, 'Sales demo bonus', 'ops.manager@demo.hrms.local'),
        ('adjustment:qa-two:deduction', 'qa.two@demo.hrms.local', 'DEDUCTION', 'ABSENCE', -18000::numeric, false, 'Unapproved absence adjustment', 'hr.manager@demo.hrms.local')
) AS a(seed, employee_email, type, category, amount, is_taxable, reason, creator_email)
JOIN users employee_user ON employee_user.email = a.employee_email
JOIN employees e ON e.user_id = employee_user.id
JOIN users creator ON creator.email = a.creator_email
ON CONFLICT (id) DO UPDATE
SET amount = EXCLUDED.amount,
    reason = EXCLUDED.reason;

WITH cycle AS (
    SELECT *
    FROM payroll_cycles
    WHERE id = pg_temp.demo_uuid('payroll-cycle:demo-org:2026-05')
),
attendance AS (
    SELECT
        employee_id,
        count(*) FILTER (WHERE status IN ('PRESENT', 'LATE', 'ON_LEAVE'))::numeric AS paid_days,
        count(*) FILTER (WHERE status = 'ABSENT')::numeric AS unpaid_days,
        count(*) FILTER (WHERE status = 'LATE')::int AS late_days,
        count(*) FILTER (WHERE status = 'ABSENT')::int AS absent_days
    FROM attendance_records
    WHERE org_id = '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2'
      AND date BETWEEN '2026-05-01' AND '2026-05-31'
    GROUP BY employee_id
),
working_days AS (
    SELECT count(*)::int AS total
    FROM working_calendar
    WHERE org_id = '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2'
      AND date BETWEEN '2026-05-01' AND '2026-05-31'
      AND type = 'WORKDAY'
),
adjustments AS (
    SELECT
        employee_id,
        coalesce(sum(amount) FILTER (WHERE amount > 0), 0) AS bonuses_total,
        abs(coalesce(sum(amount) FILTER (WHERE amount < 0), 0)) AS deductions_total
    FROM payroll_adjustments
    WHERE org_id = '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2'
      AND cycle_id = pg_temp.demo_uuid('payroll-cycle:demo-org:2026-05')
    GROUP BY employee_id
),
calc AS (
    SELECT
        e.id AS employee_id,
        e.org_id,
        e.salary_rate AS base_salary,
        wd.total AS working_days,
        coalesce(a.paid_days, wd.total) AS paid_days,
        coalesce(a.unpaid_days, 0) AS unpaid_days,
        coalesce(a.late_days, 0) AS late_days,
        coalesce(a.absent_days, 0) AS absent_days,
        CASE WHEN u.email IN ('backend.one@demo.hrms.local', 'devops.one@demo.hrms.local') THEN 120 ELSE 0 END AS overtime_minutes,
        round((e.salary_rate / wd.total) * coalesce(a.unpaid_days, 0) * -1, 2) AS attendance_adjustment,
        CASE WHEN u.email IN ('backend.one@demo.hrms.local', 'devops.one@demo.hrms.local') THEN 35000 ELSE 0 END AS overtime_amount,
        coalesce(adj.bonuses_total, 0) AS bonuses_total,
        coalesce(adj.deductions_total, 0) AS deductions_total
    FROM employees e
    JOIN users u ON u.id = e.user_id
    CROSS JOIN working_days wd
    LEFT JOIN attendance a ON a.employee_id = e.id
    LEFT JOIN adjustments adj ON adj.employee_id = e.id
    WHERE e.org_id = '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2'
),
totals AS (
    SELECT
        *,
        round(base_salary + attendance_adjustment + overtime_amount + bonuses_total - deductions_total, 2) AS gross_salary
    FROM calc
)
INSERT INTO payroll_items (
    id, cycle_id, org_id, employee_id, base_salary, attendance_adjustment, overtime_amount,
    bonuses_total, deductions_total, taxes_total, gross_salary, net_salary, working_days,
    paid_days, unpaid_days, late_days, absent_days, overtime_minutes, status,
    calculation_snapshot, missing_days, review_required, review_reasons,
    employer_taxes_total, total_employer_cost, currency, reviewed_by, reviewed_at, review_comment
)
SELECT
    pg_temp.demo_uuid('payroll-item:2026-05:' || t.employee_id),
    cycle.id,
    t.org_id,
    t.employee_id,
    t.base_salary,
    t.attendance_adjustment,
    t.overtime_amount,
    t.bonuses_total,
    t.deductions_total,
    round(t.gross_salary * 0.10, 2),
    t.gross_salary,
    round(t.gross_salary - (t.gross_salary * 0.10), 2),
    t.working_days,
    t.paid_days,
    t.unpaid_days,
    t.late_days,
    t.absent_days,
    t.overtime_minutes,
    CASE WHEN t.absent_days > 0 OR t.late_days > 1 THEN 'CALCULATED' ELSE 'READY_FOR_PAYMENT' END,
    jsonb_build_object(
        'seed', 'demo_organization_seed',
        'period', '2026-05',
        'baseSalary', t.base_salary,
        'paidDays', t.paid_days,
        'unpaidDays', t.unpaid_days,
        'lateDays', t.late_days
    ),
    0,
    (t.absent_days > 0 OR t.late_days > 1),
    CASE
        WHEN t.absent_days > 0 THEN jsonb_build_array('ABSENCE_REVIEW')
        WHEN t.late_days > 1 THEN jsonb_build_array('MULTIPLE_LATE_DAYS')
        ELSE '[]'::jsonb
    END,
    round(t.gross_salary * 0.095, 2),
    round(t.gross_salary + (t.gross_salary * 0.095), 2),
    'KZT',
    CASE WHEN t.absent_days > 0 OR t.late_days > 1 THEN reviewer.id ELSE NULL END,
    CASE WHEN t.absent_days > 0 OR t.late_days > 1 THEN '2026-06-04 16:00+05'::timestamptz ELSE NULL END,
    CASE WHEN t.absent_days > 0 THEN 'Requires HR review due to absence' WHEN t.late_days > 1 THEN 'Reviewed repeated late arrivals' ELSE NULL END
FROM totals t
CROSS JOIN cycle
LEFT JOIN users reviewer ON reviewer.email = 'finance.admin@demo.hrms.local'
ON CONFLICT (cycle_id, employee_id) DO UPDATE
SET base_salary = EXCLUDED.base_salary,
    attendance_adjustment = EXCLUDED.attendance_adjustment,
    overtime_amount = EXCLUDED.overtime_amount,
    bonuses_total = EXCLUDED.bonuses_total,
    deductions_total = EXCLUDED.deductions_total,
    taxes_total = EXCLUDED.taxes_total,
    gross_salary = EXCLUDED.gross_salary,
    net_salary = EXCLUDED.net_salary,
    paid_days = EXCLUDED.paid_days,
    unpaid_days = EXCLUDED.unpaid_days,
    late_days = EXCLUDED.late_days,
    absent_days = EXCLUDED.absent_days,
    status = EXCLUDED.status,
    calculation_snapshot = EXCLUDED.calculation_snapshot,
    review_required = EXCLUDED.review_required,
    review_reasons = EXCLUDED.review_reasons,
    employer_taxes_total = EXCLUDED.employer_taxes_total,
    total_employer_cost = EXCLUDED.total_employer_cost,
    reviewed_by = EXCLUDED.reviewed_by,
    reviewed_at = EXCLUDED.reviewed_at,
    review_comment = EXCLUDED.review_comment,
    updated_at = now();

INSERT INTO payslips (
    id, org_id, employee_id, payroll_cycle_id, payroll_item_id, period_start, period_end,
    status, currency, base_salary, overtime_amount, bonuses_total, deductions_total,
    taxes_total, gross_salary, net_salary, employer_taxes_total, total_employer_cost,
    payload_snapshot, sent_to_email, sent_at, generated_by
)
SELECT
    pg_temp.demo_uuid('payslip:2026-05:' || pi.employee_id),
    pi.org_id,
    pi.employee_id,
    pi.cycle_id,
    pi.id,
    '2026-05-01',
    '2026-05-31',
    'GENERATED',
    'KZT',
    pi.base_salary,
    pi.overtime_amount,
    pi.bonuses_total,
    pi.deductions_total,
    pi.taxes_total,
    pi.gross_salary,
    pi.net_salary,
    pi.employer_taxes_total,
    pi.total_employer_cost,
    jsonb_build_object('seed', 'demo_organization_seed', 'payrollItemId', pi.id, 'period', '2026-05'),
    u.email,
    '2026-06-05 10:00+05',
    generator.id
FROM payroll_items pi
JOIN employees e ON e.id = pi.employee_id
JOIN users u ON u.id = e.user_id
JOIN users generator ON generator.email = 'finance.admin@demo.hrms.local'
WHERE pi.cycle_id = pg_temp.demo_uuid('payroll-cycle:demo-org:2026-05')
ON CONFLICT (id) DO UPDATE
SET base_salary = EXCLUDED.base_salary,
    overtime_amount = EXCLUDED.overtime_amount,
    bonuses_total = EXCLUDED.bonuses_total,
    deductions_total = EXCLUDED.deductions_total,
    taxes_total = EXCLUDED.taxes_total,
    gross_salary = EXCLUDED.gross_salary,
    net_salary = EXCLUDED.net_salary,
    employer_taxes_total = EXCLUDED.employer_taxes_total,
    total_employer_cost = EXCLUDED.total_employer_cost,
    payload_snapshot = EXCLUDED.payload_snapshot,
    sent_to_email = EXCLUDED.sent_to_email,
    sent_at = EXCLUDED.sent_at,
    updated_at = now();

INSERT INTO notifications (id, user_id, org_id, type, title, message, metadata, is_read)
SELECT
    pg_temp.demo_uuid('notification:demo-welcome:' || u.email),
    u.id,
    u.org_id,
    'system',
    'Demo data ready',
    'Your demo workspace now includes employees, attendance, tasks, and payroll data for May 2026.',
    jsonb_build_object('seed', 'demo_organization_seed', 'period', '2026-05'),
    CASE WHEN u.email IN ('testadmin@mail.ru', 'testmanager@mail.ru') THEN true ELSE false END
FROM users u
WHERE u.org_id = '3db21a6d-abca-49a5-8f56-f3cbbc0f22e2'
ON CONFLICT (id) DO UPDATE
SET message = EXCLUDED.message,
    metadata = EXCLUDED.metadata,
    is_read = EXCLUDED.is_read;

COMMIT;
