package handler

import (
	"net/http"

	"github.com/hfiorillo/site/view/posts"

	"github.com/hfiorillo/site/models"
)

// Handles things
type PostsHandler struct {
	BlogPosts []models.BlogPost
}

func NewPostsHandler(blogPosts []models.BlogPost) PostsHandler {
	return PostsHandler{
		BlogPosts: blogPosts,
	}
}

// Returns a list of blog posts
func (p PostsHandler) ListBlogPosts(w http.ResponseWriter, r *http.Request) error {
	return posts.Posts(p.BlogPosts).Render(r.Context(), w)
}

func (p PostsHandler) DisplayBlogPosts(w http.ResponseWriter, r *http.Request) error {
	return posts.Posts(p.BlogPosts).Render(r.Context(), w)
}

// func (p PostsHandler) DisplayBlogPosts(w http.ResponseWriter, r *http.Request) error {
// 	return posts.Posts(p.BlogPosts).Render(r.Context(), w)
// }
