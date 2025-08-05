package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/config"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/databases/mongodb"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/databases/postgres"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/handlers"
	mdw "gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/middleware"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/post"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/session"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/user"
	"go.uber.org/zap"
)

func index(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("../../static/html/index.html"))

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

	pgPool, err := postgres.ConnectPool(context.Background(), config)
	if err != nil {
		sugar.Errorf("failed connect to postgres: %s", err)
		return
	}
	defer pgPool.Close()
	fmt.Println("connected to PostgreSQL database!")

	mongoClient, err := mongodb.Connect(context.Background(), config)
	if err != nil {
		sugar.Errorf("failed connect to postgres: %s", err)
		return
	}
	defer mongoClient.Disconnect(context.Background())
	fmt.Println("connected to MongoDB!")

	postsCollection := mongoClient.Database(config.MongoDB.Name).Collection("posts")
	err = mongodb.SetIndex(context.Background(), postsCollection, "votes.user")
	if err != nil {
		sugar.Errorf("failed to create mongo index: %s", err)
		return
	}

	r := mux.NewRouter()
	http.NewServeMux()

	staticHandler := http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("../../static")),
	)
	r.PathPrefix("/static/").Handler(staticHandler)

	userRepo := user.NewUserDBRepo(pgPool)
	postRepo := post.NewPostMemoryRepo(postsCollection)

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
	log.Println("server started")

	err = http.ListenAndServe(addr, mux)
	if err != nil {
		sugar.Errorw("failed to run server")
	}
}
