package handler

// const postsPath = "./content/posts"

// type PostsHandler struct {
// 	BlogPosts []models.BlogPost
// }

// func NewPostsHandler(logger *slog.Logger) PostsHandler {

// 	posts, err := markdown.LoadMarkdownPosts(postsPath)
// 	if err != nil {
// 		logger.Error("error loading markdown posts dir: %s", postsPath)
// 	}

// 	return PostsHandler{
// 		BlogPosts: posts,
// 	}
// }

// // Returns a list of blog posts
// func (p PostsHandler) ListBlogPosts(w http.ResponseWriter, r *http.Request) error {
// 	return posts.Posts(p.BlogPosts).Render(r.Context(), w)
// }

// func (p PostsHandler) DisplayBlogPosts(w http.ResponseWriter, r *http.Request) error {
// 	return posts.Posts(p.BlogPosts).Render(r.Context(), w)
// }
