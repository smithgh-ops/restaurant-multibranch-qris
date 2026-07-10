-- Migration: 011_alter_user_branch_roles_nullable_branch.sql
-- Description: Allow null branch_id to represent organization-wide (global) role assignments.
--              A null branch_id means the role applies across the whole organization.

ALTER TABLE user_branch_roles
    MODIFY COLUMN branch_id BIGINT UNSIGNED NULL COMMENT 'NULL = global org-level role';

-- Drop the old unique constraint and recreate without the branch_id NOT NULL assumption.
-- MySQL allows multiple NULL values in a unique index so a user can have multiple global
-- roles, but still cannot have the same (user_id, branch_id, role_id) combination.
DROP INDEX uq_user_branch_role ON user_branch_roles;
CREATE UNIQUE INDEX uq_user_branch_role ON user_branch_roles (user_id, branch_id, role_id);
