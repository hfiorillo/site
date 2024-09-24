package handler

// func HandleLongProcess(w http.ResponseWriter, r *http.Request) error {
// 	time.Sleep(time.Second * 5)
// 	return home.UserLikes(10000).Render(r.Context(), w)
// }

// const aboutMePath = "./content/aboutme"

// type GeneralHandler struct {
// 	Logger  *slog.Logger
// 	AboutMe models.BlogPost
// }

// func NewGeneralHandler(logger *slog.Logger) GeneralHandler {

// 	aboutMe, err := markdown.LoadMarkdownPosts(aboutMePath)
// 	if err != nil {
// 		logger.Error("error loading aboutme dir: %s")
// 	}

// 	return GeneralHandler{
// 		AboutMe: aboutMe[0],
// 	}
// }

// func (h GeneralHandler) HandleAboutMe(w http.ResponseWriter, r *http.Request) error {
// 	return pages.AboutMe(h.AboutMe).Render(r.Context(), w)
// }

// func HandleHomeIndex(w http.ResponseWriter, r *http.Request) error {
// 	return pages.Index().Render(r.Context(), w)
// }
