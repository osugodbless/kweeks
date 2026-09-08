-- Wallet provisioning now follows the strict per-step BMONI flow
-- (create-user → KYC → wallet → rail). Add the stage flags alongside the
-- existing bmoni identity columns.

alter table wallets add column if not exists bmoni_kyc_submitted boolean not null default false;
alter table wallets add column if not exists bmoni_rail_active boolean not null default false;
