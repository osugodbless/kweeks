-- Instructors now carry explicit first_name/last_name so the BMONI user
-- identity is built from what the host typed on signup, never from re-splitting
-- a free-form name (a name field is ambiguous when the host types surname
-- first). Existing rows are backfilled best-effort from the stored name;
-- identity code falls back to that split only for pre-existing accounts.

alter table instructors add column if not exists first_name text not null default '';
alter table instructors add column if not exists last_name text not null default '';

update instructors
set first_name = split_part(name, ' ', 1),
    last_name = case
        when position(' ' in name) > 0 then substring(name from position(' ' in name) + 1)
        else ''
    end
where first_name = '';
