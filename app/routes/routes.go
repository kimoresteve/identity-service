package routes

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/kimoresteve/identity-service/app/controllers"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"os"
)

type App struct {
	Router     *mux.Router
	Controller *controllers.Controller
}

func (a *App) Initialize() {

	a.Router = mux.NewRouter()
	//db := database.GetDBConnection()

	// Add Swagger handler
	a.Router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	a.setRoutes()

}

func (a *App) setRoutes() {

	//status
	a.Router.HandleFunc("/", a.Controller.Status).Methods("POST")
	a.Router.HandleFunc("/", a.Controller.Status).Methods("GET")

	//auth

	authRouter := a.Router.PathPrefix("/auth").Subrouter()
	//a.Router.HandleFunc("/auth/register", a.Controller.Register).Methods("POST")
	authRouter.HandleFunc("/register/agency", a.Controller.RegisterAgency).Methods("POST")
	authRouter.HandleFunc("/register/agency/landlord", a.Controller.Login).Methods("POST")
	authRouter.HandleFunc("/register/landlord", a.Controller.RegisterLandlord).Methods("POST")
	a.Router.HandleFunc("/auth/login", a.Controller.Login).Methods("POST")
	a.Router.HandleFunc("/auth/verify", a.Controller.Verify).Methods("POST")
	a.Router.HandleFunc("/auth/forgot-password", a.Controller.ForgotPassword).Methods("POST")
	a.Router.HandleFunc("/auth/reset-password", a.Controller.ResetPassword).Methods("POST")

	////auth
	//a.Router.HandleFunc("/auth/register", a.Controller.RegisterLandlord).Methods("POST")
	//a.Router.HandleFunc("/auth/login", a.Controller.LoginLandlord).Methods("POST")
	//a.Router.HandleFunc("/auth/verification", a.Controller.VerifyLandlord).Methods("POST")
	//a.Router.HandleFunc("/auth/forgot", a.Controller.ForgotPassword).Methods("POST")
	//a.Router.HandleFunc("/auth/resetPassword", a.Controller.ResetPassword).Methods("POST")
	//
	//a.Router.HandleFunc("/auth/resendOtp", a.Controller.ResendOTP).Methods("POST")
	//
	////landlords
	//a.Router.Handle("/landlord/{landlordID}/units", middleware.JWTMiddleware(http.HandlerFunc(a.Controller.GetUnitsByLandlordID))).Methods("GET")
	//
	////flats
	//a.Router.Handle("/flat", middleware.JWTMiddleware(http.HandlerFunc(a.Controller.AddFlat))).Methods("POST")
	//a.Router.Handle("/flat/{landlordID}", middleware.JWTMiddleware(http.HandlerFunc(a.Controller.GetFlatsByLandlordID))).Methods("GET")
	//a.Router.Handle("flat/update/{id}", middleware.JWTMiddleware(http.HandlerFunc(a.Controller.UpdateFlat))).Methods("PATCH")
	//a.Router.Handle("/flat/{id}", middleware.JWTMiddleware(http.HandlerFunc(a.Controller.RemoveFlat))).Methods("DELETE")
	//
	////
	//a.Router.Handle("/flat/{flatID}/units", middleware.JWTMiddleware(http.HandlerFunc(a.Controller.GetUnitsByFlatID))).Methods("GET")
	//
	////units
	//a.Router.Handle("/unit", middleware.JWTMiddleware(http.HandlerFunc(a.Controller.AddUnit))).Methods("POST")
	//a.Router.Handle("/unit/update/{id}", middleware.JWTMiddleware(http.HandlerFunc(a.Controller.UpdateUnit))).Methods("PATCH")
	//a.Router.Handle("/unit/delete/{id}", middleware.JWTMiddleware(http.HandlerFunc(a.Controller.DeleteUnit))).Methods("DELETE")

}

func (a *App) Run() {

	envErr := godotenv.Load()

	if envErr != nil {
		log.Fatal("Error loading .env file")
	}
	host := os.Getenv("SYSTEM_HOST")
	port := os.Getenv("SYSTEM_PORT")
	fmt.Printf("serving on %s:%s\n", host, port)
	server := fmt.Sprintf("%s:%s", host, port)
	err := http.ListenAndServe(server, a.Router)
	panic(err)

}
