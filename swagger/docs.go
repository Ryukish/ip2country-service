package swagger

import (
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title IP to Country Service API
// @version 1.1
// @description This is a service to map IP addresses to their corresponding country and city.

// @contact.name Mathuranath Metivier
// @contact.url ---------
// @contact.email --------

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

func RegisterSwagger(router *mux.Router) {
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
}
