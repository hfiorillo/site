package handler

import (
	"net/http"

	"github.com/hfiorillo/site/view/home"
)

// func HandleLongProcess(w http.ResponseWriter, r *http.Request) error {
// 	time.Sleep(time.Second * 5)
// 	return home.UserLikes(10000).Render(r.Context(), w)
// }

func HandleAboutMe(w http.ResponseWriter, r *http.Request) error {
	return home.AboutMe().Render(r.Context(), w)
}
