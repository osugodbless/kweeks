-- Kweeks schema. Plain Postgres (no Supabase features). Pool + claims are
-- exact integer amounts (kobo). Correct answers live here and nowhere else.

create extension if not exists pgcrypto;

create table if not exists quizzes (
    id           text primary key,
    instructor_id text not null,
    title        text not null,
    pool_kobo    bigint not null,
    winner_count int not null,
    pacing       text not null check (pacing in ('auto','manual')),
    default_duration_ms bigint not null,
    questions    jsonb not null,   -- [{id,prompt,options,correct_index,duration_ms}]
    created_at   timestamptz not null default now()
);

create index if not exists quizzes_instructor_idx on quizzes (instructor_id);

create table if not exists rooms (
    id          text primary key,
    quiz_id     text not null references quizzes (id),
    state       text not null check (state in ('lobby','live','podium','ended')),
    host_id     text not null,
    current_question_idx int not null default -1,
    question_started_at  timestamptz,
    started_at  timestamptz,
    created_at  timestamptz not null default now()
);

create index if not exists rooms_quiz_idx on rooms (quiz_id);

create table if not exists participants (
    id          text primary key,
    room_id     text not null references rooms (id) on delete cascade,
    email       text not null,
    nickname    text not null,
    avatar      text not null,
    joined_at   timestamptz not null default now(),
    unique (room_id, email)
);

create index if not exists participants_room_idx on participants (room_id);

create table if not exists answers (
    id                text primary key,
    room_id           text not null references rooms (id) on delete cascade,
    participant_id    text not null references participants (id) on delete cascade,
    question_id       text not null,
    option_index      int not null,
    correct           boolean not null,
    question_started_at timestamptz not null,
    received_at       timestamptz not null,
    unique (participant_id, question_id)
);

create index if not exists answers_room_idx on answers (room_id);

create table if not exists claims (
    id         text primary key,
    quiz_id    text not null references quizzes (id),
    room_id    text not null references rooms (id),
    email      text not null,
    amount_kobo bigint not null,
    claim_code text not null,
    state      text not null check (state in ('created','invited','onboarded','paid','failed')),
    created_at timestamptz not null default now(),
    paid_at    timestamptz,
    unique (quiz_id, email)
);

create index if not exists claims_quiz_idx on claims (quiz_id);
