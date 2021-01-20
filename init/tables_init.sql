create extension if not exists citext;

drop
    table if exists users cascade;
drop
    table if exists forums cascade;
drop
    table if exists threads cascade;
drop
    table if exists posts cascade;
drop
    table if exists votes cascade;
drop
    table if exists forum_users cascade;

create unlogged table if not exists users (
     nickname citext not null unique,
     fullname text,
     email citext not null unique,
     about text
);

create index if not exists nickname_idx
    on users (nickname);

create unlogged table if not exists forums (
    posts integer default 0,
    slug citext not null unique,
    threads integer default 0,
    title text not null,
    "user" citext references users (nickname) not null
);

create index if not exists forums_slug_idx
    on forums (slug);

create unlogged table if not exists forum_users (
    nickname citext references users (nickname) not null,
    slug citext references forums (slug) not null,
    unique (nickname, slug)
);

create unlogged table if not exists threads (
    author citext references users (nickname) not null,
    created timestamptz,
    forum citext references forums (slug) not null,
    id serial primary key,
    message text not null,
    slug citext default null unique,
    title text,
    votes integer not null default 0
    );

create index if not exists threads_slug_idx
    on threads (slug);

create index if not exists threads_author_idx
    on threads (author, forum);

create index if not exists threads_timestamp_idx
    on threads (forum, created);

create unlogged table if not exists posts (
    author citext references users (nickname) not null,
    created timestamptz default current_timestamp,
    forum citext references forums (slug),
    id serial primary key,
    isedited boolean default false,
    message text not null,
    parent integer default 0,
    thread integer references threads (id) not null,
    path integer[] default array [] :: int[]
    );

create index if not exists posts_path_idx
    on posts (path);

create index if not exists posts_path_suffix_idx
    on posts ((path [1]));

create index if not exists posts_thread_idx
    on posts (thread);

create index if not exists posts_thread_id_idx
    on posts (thread, id);

create index if not exists posts_thread_path_idx
    on posts (thread, path, id);

create index if not exists posts_id_path_suffix_idx
    on posts (id, (path [1]));

create index if not exists posts_author_idx
    on posts (author, forum);

create index if not exists posts_thread_suffix_parent_idx
    on posts (thread, id, (path[1]), parent);

create unlogged table if not exists votes (
    nickname citext references users (nickname) not null,
    voice smallint
    check (
              voice in (-1, 1)
    ),
    thread integer references threads (id) not null,
    unique (nickname, thread)
);

create index if not exists votes_name_idx
    on votes (nickname, thread);
