create table goschema_migration_history
(
    id         int auto_increment not null,
    version    varchar(255) not null,
    action     enum ('migrating_up', 'migrating_down', 'migrated_up', 'migrated_down', 'migration_error') not null,
    created_at timestamp    not null,
    primary key (id)
);

