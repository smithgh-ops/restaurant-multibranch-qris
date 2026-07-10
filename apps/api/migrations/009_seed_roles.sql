-- Migration: 009_seed_roles.sql
-- Description: Seed standard application roles (no user credentials here)

INSERT IGNORE INTO roles (name, description) VALUES
    ('super_admin',    'Full platform access across all organizations'),
    ('org_admin',      'Administrative access scoped to one organization'),
    ('branch_manager', 'Management access for a specific branch'),
    ('cashier',        'POS and payment operations at a branch'),
    ('kitchen_staff',  'Kitchen Display System view for a branch'),
    ('waiter',         'Order-taking and table management at a branch');
