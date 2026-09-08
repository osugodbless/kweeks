-- Winner redemption payout details: the claim now carries the winner's
-- Nigerian bank payout state (verify -> register -> offramp) instead of the
-- no-app invite flow.

alter table claims add column if not exists bank_account_id text not null default '';
alter table claims add column if not exists payout_ref text not null default '';
alter table claims add column if not exists bank_account_number text not null default '';
alter table claims add column if not exists bank_name text not null default '';
alter table claims add column if not exists account_holder_name text not null default '';

-- State machine now: created -> bank_submitted -> paying -> paid (or failed).
alter table claims drop constraint if exists claims_state_check;
alter table claims add constraint claims_state_check
    check (state in ('created','bank_submitted','paying','paid','failed'));
