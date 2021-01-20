drop database if exists hw_db;
drop user if exists hw_db_user;

create user hw_db_user
with superuser
password 'hw_db_password';

create database hw_db
with owner = hw_db_user
    encoding = 'UTF-8'
    template = template0
    LC_COLLATE = 'C'
    LC_CTYPE = 'C'
    connection limit = -1;

