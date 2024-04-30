package handler

import (
	"net/http"
	"site/models"
	"site/view/posts"
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

func (p PostsHandler) ListBlogPosts(w http.ResponseWriter, r *http.Request) error {
	return posts.Posts(p.BlogPosts).Render(r.Context(), w)
}