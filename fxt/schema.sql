create table events (
  position bigserial,
  label text
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
