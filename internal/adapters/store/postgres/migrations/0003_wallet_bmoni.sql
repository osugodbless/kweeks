-- External BMONI rail identity per wallet (provisioned at signup when enabled).

alter table wallets add column if not exists bmoni_user_id   text not null default '';
alter table wallets add column if not exists bmoni_wallet_id text not null default '';
alter table wallets add column if not exists bmoni_wallet_addr text not null default '';
