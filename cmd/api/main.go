package main

import (
	"social/internal/auth"
	"social/internal/db"
	"social/internal/env"
	"social/internal/mailer"
	"social/internal/store"
	"social/internal/store/cache"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	err := godotenv.Load()
	if err != nil {
		logger.Fatal(err)
	}
	mail := &mailConfig{
		exp:       time.Hour * 24 * 3,
		fromEmail: env.GetString("FROM_EMAIL", ""),
		sendGrid: sendGridConfig{
			apiKey: env.GetString("SENDGRID_API_KEY", ""),
		},
		mailTrap: mailTrapConfig{
			apiKey: env.GetString("MAILTRAP_API_KEY", ""),
		},
	}
	cfg := &config{
		addr:        env.GetString("ADDR", ":8080"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:5173"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/socialnetwork?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		redisCfg: redisConfig{
			addr: env.GetString("REDIS_ADDR", "localhost:6379"),
			pw:   env.GetString("REDIS_PW", ""),
			db:   env.GetInt("REDIS_DB", 0),
		},
		mail: *mail,
		auth: authConfig{
			token: tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", "example"),
				exp:    time.Hour * 24 * 3, // 3 days
				iss:    "nishadsocial",     // jwt issuer
			},
		},
	}

	// Main Database
	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		logger.Fatal(err)
	}

	mailtrap, err := mailer.NewMailTrapClient(cfg.mail.mailTrap.apiKey, cfg.mail.fromEmail)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("database connection pool established")

	// var rdb *redis.Client
	rdb := cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.pw, cfg.redisCfg.db)
	logger.Info("redis cache connection established")

	defer rdb.Close()

	store := store.NewStorage(db)
	cacheStorage := cache.NewRedisStorage(rdb) // redis

	jwtAuthenticator := auth.NewJWTAuthenticator(
		cfg.auth.token.secret,
		cfg.auth.token.iss, // jwt issuer
		cfg.auth.token.iss,
	)
	app := &application{
		config:        *cfg,
		store:         store,
		cacheStorage:  cacheStorage,
		logger:        logger,
		mailer:        mailtrap,
		authenticator: jwtAuthenticator,
	}
	mux := app.mount()

	logger.Fatal(app.run(mux))
}
