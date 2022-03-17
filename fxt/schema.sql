create table events (
  position bigserial,
  recorded_at timestamptz
);

create table samples (
  id text,
  label text
);

create table users (
  id text,
  email text,
  token text
);
