package handler

import (
	"log/slog"
	"net/http"

	"github.com/hfiorillo/site/models"
	"github.com/hfiorillo/site/pkg"
	"github.com/hfiorillo/site/view/home"
)

// func HandleLongProcess(w http.ResponseWriter, r *http.Request) error {
// 	time.Sleep(time.Second * 5)
// 	return home.UserLikes(10000).Render(r.Context(), w)
// }

const aboutMePath = "./content/aboutme"

type GeneralHandler struct {
	Logger  *slog.Logger
	AboutMe models.BlogPost
}

func NewGeneralHandler(logger *slog.Logger) GeneralHandler {

	aboutMe, err := pkg.LoadMarkdownPosts(aboutMePath)
	if err != nil {
		logger.Error("error loading aboutme dir: %s", postsPath)
	}

	return GeneralHandler{
		AboutMe: aboutMe[0],
	}
}

func (h GeneralHandler) HandleAboutMe(w http.ResponseWriter, r *http.Request) error {
	return home.AboutMe(h.AboutMe).Render(r.Context(), w)
}

func HandleHomeIndex(w http.ResponseWriter, r *http.Request) error {
	return home.Index().Render(r.Context(), w)
}
