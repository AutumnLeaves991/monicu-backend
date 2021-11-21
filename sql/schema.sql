create table if not exists "user"
(
    id         serial
        constraint user_pk
            primary key,
    discord_id bigint not null
);

alter table "user"
    owner to monicu;

create unique index if not exists user_discord_id_uindex
    on "user" (discord_id);

create unique index if not exists user_id_uindex
    on "user" (id);

create table if not exists guild
(
    id         serial
        constraint guild_pk
            primary key,
    discord_id bigint not null
);

alter table guild
    owner to monicu;

create unique index if not exists guild_discord_id_uindex
    on guild (discord_id);

create unique index if not exists guild_id_uindex
    on guild (id);

create table if not exists channel
(
    id         serial,
    discord_id bigint  not null,
    guild_id   integer not null
        constraint channel_guild_id_fk
            references guild
            on update cascade on delete cascade
);

alter table channel
    owner to monicu;

create unique index if not exists channel_discord_id_uindex
    on channel (discord_id);

create unique index if not exists channel_id_uindex
    on channel (id);

create table if not exists post
(
    id         serial
        constraint post_pk
            primary key,
    discord_id bigint        not null,
    channel_id integer       not null
        constraint post_channel_id_fk
            references channel (id)
            on update cascade on delete cascade,
    user_id    integer       not null
        constraint post_user_id_fk
            references "user"
            on update cascade on delete cascade,
    message    varchar(2000) not null
);

alter table post
    owner to monicu;

create unique index if not exists post_discord_id_uindex
    on post (discord_id);

create unique index if not exists post_id_uindex
    on post (id);

create table if not exists image
(
    id      serial
        constraint image_pk
            primary key,
    post_id integer not null
        constraint image_post_id_fk
            references post
            on update cascade on delete cascade,
    url     text    not null,
    width   integer not null,
    height  integer not null,
    size    bigint  not null
);

alter table image
    owner to monicu;

create unique index if not exists image_id_uindex
    on image (id);

create table if not exists emoji
(
    id         serial
        constraint emoji_pk
            primary key,
    discord_id bigint,
    name       varchar(32) not null
);

alter table emoji
    owner to monicu;

create unique index if not exists emoji_discord_id_uindex
    on emoji (discord_id);

create unique index if not exists emoji_id_uindex
    on emoji (id);

create unique index if not exists emoji_name_uindex
    on emoji (name)
    where (discord_id IS NULL);

create table if not exists reaction
(
    id       serial
        constraint reaction_pk
            primary key,
    post_id  integer not null
        constraint reaction_post_id_fk
            references post
            on update cascade on delete cascade,
    emoji_id integer not null
        constraint reaction_emoji_id_fk
            references emoji
            on update cascade on delete cascade
);

alter table reaction
    owner to monicu;

create unique index if not exists reaction_id_uindex
    on reaction (id);

create unique index if not exists reaction_post_id_emoji_id_uindex
    on reaction (post_id, emoji_id);

create table if not exists user_reaction
(
    id          serial
        constraint user_reaction_pk
            primary key,
    reaction_id integer not null
        constraint user_reaction_reaction_id_fk
            references reaction
            on update cascade on delete cascade,
    user_id     integer not null
        constraint user_reaction_user_id_fk
            references "user"
            on update cascade on delete cascade
);

alter table user_reaction
    owner to monicu;

create unique index if not exists user_reaction_id_uindex
    on user_reaction (id);

create unique index if not exists user_reaction_reaction_id_user_id_uindex
    on user_reaction (reaction_id, user_id);

