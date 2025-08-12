package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/teatah/rclone/pkg/config"
	"github.com/teatah/rclone/pkg/databases/mongodb"
	"github.com/teatah/rclone/pkg/databases/postgres"
	"github.com/teatah/rclone/pkg/handlers"
	mdw "github.com/teatah/rclone/pkg/middleware"
	"github.com/teatah/rclone/pkg/post"
	"github.com/teatah/rclone/pkg/session"
	"github.com/teatah/rclone/pkg/user"
	"go.uber.org/zap"
)

func index(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./web/static/html/index.html"))

	err := tmpl.Execute(w, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	logger := zap.NewExample()
	defer func() {
		err := logger.Sync()
		if err != nil {
			log.Print(err)
		}
	}()
	sugar := logger.Sugar()

	config, err := config.LoadConfig()
	if err != nil {
		sugar.Errorf("failed to load config: %s", err)
		return
	}

	ctx := context.Background()
	pgPool, err := postgres.ConnectPool(ctx, config)
	if err != nil {
		sugar.Errorf("failed connect to postgres: %s", err)
		return
	}
	defer pgPool.Close()
	fmt.Println("connected to PostgreSQL database!")

	mongoClient, err := mongodb.Connect(ctx, config)
	if err != nil {
		sugar.Errorf("failed connect to mongo: %s", err)
		return
	}
	defer func() {
		disconErr := mongoClient.Disconnect(ctx)
		if disconErr != nil {
			sugar.Errorf("failed to disconnect from mongo: %s", disconErr)
		}
	}()
	fmt.Println("connected to MongoDB!")

	postsCollection := mongoClient.Database(config.MongoDB.Name).Collection("posts")
	err = mongodb.SetIndex(ctx, postsCollection, "votes.user")
	if err != nil {
		sugar.Errorf("failed to create mongo index: %s", err)
		return
	}

	r := mux.NewRouter()
	http.NewServeMux()

	staticHandler := http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./web/static")),
	)
	r.PathPrefix("/static/").Handler(staticHandler)

	userRepo := user.NewUserDBRepo(pgPool)
	postRepo := post.NewPostDBRepo(postsCollection)

	sm := session.NewDBSessionManager(pgPool)

	userHandler := handlers.UserHandler{
		SessionManager: sm,
		Logger:         sugar,
		UserRepo:       userRepo,
	}

	ph := handlers.PostHandler{
		SessionManager: sm,
		Logger:         sugar,
		PostRepo:       postRepo,
		UserRepo:       userRepo,
	}

	r.HandleFunc("/", index).Methods("GET")
	r.HandleFunc("/api/register", userHandler.Register).Methods(http.MethodPost)
	r.HandleFunc("/api/login", userHandler.Login).Methods(http.MethodPost)
	r.HandleFunc("/api/posts/", ph.Posts).Methods(http.MethodGet)
	r.HandleFunc("/api/post/{postID}", ph.GetPost).Methods(http.MethodGet)
	r.HandleFunc("/api/user/{username}", ph.PostsByUser).Methods(http.MethodGet)
	r.HandleFunc("/api/posts/{category}", ph.PostsByCategory).Methods(http.MethodGet)

	createPostHandler := mdw.AuthMiddleware(sm, sugar, http.HandlerFunc(ph.CreatePost))
	r.Handle("/api/posts", createPostHandler).Methods(http.MethodPost)

	deletePostHandler := mdw.AuthMiddleware(sm, sugar, http.HandlerFunc(ph.DeletePost))
	r.Handle("/api/post/{postID}", deletePostHandler).Methods(http.MethodDelete)

	createCommentHandler := mdw.AuthMiddleware(sm, sugar, http.HandlerFunc(ph.CreateComment))
	r.Handle("/api/post/{postID}", createCommentHandler).Methods(http.MethodPost)

	deleteCommentHandler := mdw.AuthMiddleware(sm, sugar, http.HandlerFunc(ph.DeleteComment))
	r.Handle("/api/post/{postID}/{commentID}", deleteCommentHandler).Methods(http.MethodDelete)

	upvoteHandler := mdw.AuthMiddleware(sm, sugar, http.HandlerFunc(ph.Upvote))
	r.Handle("/api/post/{postID}/upvote", upvoteHandler).Methods(http.MethodGet)

	downvoteHandler := mdw.AuthMiddleware(sm, sugar, http.HandlerFunc(ph.Downvote))
	r.Handle("/api/post/{postID}/downvote", downvoteHandler).Methods(http.MethodGet)

	unvoteHandler := mdw.AuthMiddleware(sm, sugar, http.HandlerFunc(ph.Unvote))
	r.Handle("/api/post/{postID}/unvote", unvoteHandler).Methods(http.MethodGet)

	mux := mdw.LogMiddleware(sugar, r)
	mux = mdw.PanicMiddleware(sugar, mux)

	addr := ":8080"

	ticker := time.NewTicker(time.Minute * 1 / 2)
	defer ticker.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-ticker.C:
				err := sm.RemoveExpiredSessions(ctx)
				if err != nil {
					sugar.Errorf("failed to remove expired sessions: %v", err)
				}
			case <-quit:
				return
			}
		}
	}()

	server := &http.Server{Addr: addr, Handler: mux}

	go func() {
		sugar.Infof("starting server at port %s", addr)

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			sugar.Errorw("failed to run server")
		}
	}()

	<-quit

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		sugar.Errorf("server shutdown failed: %+v", err)
	}
}
