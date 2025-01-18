create table goschema_migration_version
(
    version    varchar(255) not null,
    is_current tinyint(1) default 0 not null,
    created_at timestamp    not null,
    primary key (version)
);

