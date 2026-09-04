-- Kweeks instructors, sessions, wallets, wallet transactions, room codes.

alter table rooms add column if not exists code text;
create unique index if not exists rooms_code_idx on rooms (code) where code is not null;

create table if not exists instructors (
    id            text primary key,
    name          text not null,
    email         text not null unique,
    password_hash text not null,
    avatar        text not null default 'AP',
    created_at    timestamptz not null default now()
);

create table if not exists sessions (
    token          text primary key,
    instructor_id  text not null references instructors (id) on delete cascade,
    created_at     timestamptz not null default now(),
    expires_at     timestamptz not null
);

create index if not exists sessions_instructor_idx on sessions (instructor_id);

create table if not exists wallets (
    id            text primary key,
    instructor_id text not null unique references instructors (id) on delete cascade,
    balance_kobo  bigint not null default 0,
    created_at    timestamptz not null default now()
);

create table if not exists wallet_transactions (
    id         text primary key,
    wallet_id  text not null references wallets (id) on delete cascade,
    kind       text not null check (kind in ('fund','pool','payout','credit')),
    amount_kobo bigint not null,
    note       text not null default '',
    created_at timestamptz not null default now()
);

create index if not exists wallet_tx_wallet_idx on wallet_transactions (wallet_id, created_at desc);
