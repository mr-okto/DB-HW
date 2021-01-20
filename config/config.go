package config

import "github.com/jackc/pgx"

var ConnConfig = pgx.ConnConfig{
	Host:     "localhost",
	Port:     5432,
	Database: "hw_db",
	User:     "hw_db_user",
	Password: "hw_db_password",
}

var DBConfig = pgx.ConnPoolConfig{
	ConnConfig:     ConnConfig,
	MaxConnections: 1000,
}
