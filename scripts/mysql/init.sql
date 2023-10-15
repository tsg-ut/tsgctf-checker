create database if not exists tsgctf2023 default character set utf8mb4;

create user if not exists `tsgctf2023-checker`@'%' identified with mysql_native_password by 'testpass';
grant all on `tsgctf2023`.* to 'tsgctf2023-checker'@`%`;

use tsgctf2023;

create table if not exists `test_result`
(
  `name`        varchar(255)      not null,
  `result`      int               not null,
  `timestamp`   datetime           not null
);
