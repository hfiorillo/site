package handler

import (
	"net/http"

	home "github.com/hfiorillo/site/view/home"
)

// func HandleLongProcess(w http.ResponseWriter, r *http.Request) error {
// 	time.Sleep(time.Second * 5)
// 	return home.UserLikes(10000).Render(r.Context(), w)
// }

func HandleHomeIndex(w http.ResponseWriter, r *http.Request) error {
	return home.Index().Render(r.Context(), w)
}
