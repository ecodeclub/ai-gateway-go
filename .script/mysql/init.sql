create database if not exists ai_gateway;

create table ai_gateway.prompts
(
    id          bigint auto_increment primary key,
    name        varchar(32) not null ,
    biz         varchar(32) not null,
    pattern     text        not null ,
    description text        not null ,
    status      tinyint(1)  unsigned default '1' null,
    ctime       bigint      null,
    utime       bigint      null,
    constraint uni_biz
        unique (`biz`)
);