-- Instructors carry their own E.164 phone so each host provisions a distinct
-- BMONI user + wallet instead of sharing the persona phone (which caused every
-- signup on a shared key to resolve to the first user's BMONI identity).

alter table instructors add column if not exists phone text not null default '';
