package handler

import (
	"net/http"
	"site/models"
	"site/view/home"
)

// func HandleLongProcess(w http.ResponseWriter, r *http.Request) error {
// 	time.Sleep(time.Second * 5)
// 	return home.UserLikes(10000).Render(r.Context(), w)
// }

// Handles things
type Handler struct {
	BlogPosts []models.BlogPost
}

func NewIndexHandler(blogPosts []models.BlogPost) Handler {
	return Handler{
		BlogPosts: blogPosts,
	}
}

func HandleHomeIndex(w http.ResponseWriter, r *http.Request) error {
	return home.Index().Render(r.Context(), w)
}
