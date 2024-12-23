package index

import (
	helpers "goserver/src/Helpers"
	post "goserver/src/Routes/Post"

	http "github.com/bogdanfinn/fhttp"
)

func RegisterRoutes(mux *http.ServeMux, logger *helpers.ColorizedLogger) {
	mux.HandleFunc("/token", post.TokenHandler(logger))
}
