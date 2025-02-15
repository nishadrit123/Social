package main

import (
	"net/http"
	"social/internal/auth"
	"social/internal/env"
	"social/internal/mailer"
	"social/internal/store"
	"social/internal/store/cache"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type config struct {
	addr        string
	db          dbConfig
	redisCfg    redisConfig
	env         string
	apiURL      string
	mail        mailConfig
	frontendURL string
	auth        authConfig
}

type redisConfig struct {
	addr string
	pw   string
	db   int
}

type mailConfig struct {
	sendGrid  sendGridConfig
	mailTrap  mailTrapConfig
	fromEmail string
	exp       time.Duration
}

type mailTrapConfig struct {
	apiKey string
}

type sendGridConfig struct {
	apiKey string
}

type authConfig struct {
	token tokenConfig
}

type tokenConfig struct {
	secret string
	exp    time.Duration
	iss    string
}

type application struct {
	config        config
	store         store.Storage
	cacheStorage  cache.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{env.GetString("CORS_ALLOWED_ORIGIN", "http://localhost:5174")},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Use(middleware.RequestID)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)

			r.Post("/", app.createPostHandler)
			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.postsContextMiddleware)
				r.Get("/", app.getPostHandler)
				r.Get("/user", app.getUserofPostHandler)

				r.Patch("/", app.checkOwnership("moderator", "post", app.updatePostHandler))
				r.Delete("/", app.checkOwnership("admin", "post", app.deletePostHandler))
			})
		})

		r.Route("/likedislike", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Route("/post/{postID}", func(r chi.Router) {
				r.Use(app.postsContextMiddleware)
				r.Post("/", app.likedislikeHandler)
			})
		})

		r.Route("/comment", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)

			r.Route("/post/{postID}", func(r chi.Router) {
				r.Get("/", app.getCommentHandler)
				r.Post("/", app.createCommentHandler)
			})
			r.Route("/{commentID}", func(r chi.Router) {
				r.Use(app.commentContextMiddleware)

				r.Patch("/", app.checkOwnership("moderator", "comment", app.updateCommentHandler))
				r.Delete("/", app.checkOwnership("admin", "comment", app.deleteCommentHandler))
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler) // activate the registered user
			r.Post("/", app.createUserHandler)                  // not used

			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				// r.Use(app.usersContextMiddleware)
				r.Get("/", app.getUserHandler)
				r.Get("/allposts", app.getUserAllPostsHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})
			r.Route("/logout", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Post("/", app.logoutUserHandler) // log out and inactivate user jwt toekn
			})

			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)

				r.Get("/feed", app.getUserFeedHandler)
			})
		})

		// public routes
		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler) // registers the users and send them invites
			r.Post("/token", app.createTokenHandler) // jwt based stateless authentication
		})

	})
	return r
}

func (app *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Millisecond,
	}
	app.logger.Info("Started server on %v\n", app.config.addr)
	err := srv.ListenAndServe()
	app.logger.Info("ListenAndServe is a blocking call and wint be executed unless it throws ant err %v", err)
	if err != nil {
		return err
	}
	return nil
}
