create database if not exists ai_gateway;

create table ai_gateway.prompts
(
    id          bigint auto_increment primary key,
    name        varchar(32) not null ,
    owner       bigint not null,
    owner_type  enum('personal', 'organization ') not null,
    content     text        not null ,
    description text        not null ,
    status      tinyint(1)  unsigned default '1' null,
    ctime       bigint      null,
    utime       bigint      null,
    key idx_owner_owner_type(owner, owner_type)
);