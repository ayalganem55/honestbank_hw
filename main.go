package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	env "github.com/joho/godotenv"
)

const envFile = ".env"
const dataFile = "data/forms.json"

var loadEnv = env.Load
var templates = template.Must(template.ParseGlob("templates/*"))

type formInput struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

func (f formInput) validate() error {
	if f.FirstName == "" || f.LastName == "" || f.Email == "" || f.PhoneNumber == "" {
		return errors.New("invalid input")
	}
	return nil
}

func getData() ([]formInput, error) {
	file, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return nil, err
	}
	var forms []formInput
	err = json.Unmarshal(file, &forms)
	if err != nil {
		return nil, err
	}

	return forms, nil
}

func (f formInput) save() error {
	forms, err := getData()
	if err != nil {
		return err
	}

	toSave, err := json.Marshal(append(forms, f))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(dataFile, toSave, os.ModeAppend)
	return err
}

func handleFormFunc(resp http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		err := req.ParseForm()
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprint(resp, err.Error())
			return
		}
		f := formInput{
			FirstName:   req.FormValue("first_name"),
			LastName:    req.FormValue("last_name"),
			Email:       req.FormValue("email"),
			PhoneNumber: req.FormValue("phone_number"),
		}

		err = f.validate()
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprint(resp, err.Error())
			return
		}
		err = f.save()
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(resp, err.Error())
			return
		}
		resp.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(resp, "form saved")
	case http.MethodGet:
		resp.WriteHeader(http.StatusOK)
		renderTemplate(resp, "form.html", nil)
	default:
		log.Println("error no 404")
		resp.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(resp, "not found")
	}
}

func handleDataFunc(resp http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		forms, err := getData()
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp.WriteHeader(http.StatusOK)
		renderTemplate(resp, "table.html", forms)
	default:
		log.Println("error no 404")
		resp.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(resp, "not found")
	}
}

func run() (s *http.Server) {
	err := loadEnv(envFile)
	if err != nil {
		log.Fatal(err)
	}
	port, exist := os.LookupEnv("PORT")
	if !exist {
		log.Fatal("no port specified")
	}
	port = fmt.Sprintf(":%s", port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleFormFunc)
	mux.HandleFunc("/data", handleDataFunc)

	s = &http.Server{
		Addr:           port,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        mux,
	}

	go func() {
		fmt.Printf("Starting the server at port %s\n", port)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	return
}

func renderTemplate(resp http.ResponseWriter, templateName string, data interface{}) {
	err := templates.ExecuteTemplate(resp, templateName, data)
	if err != nil {
		resp.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(resp, err.Error())
		return
	}
}

func main() {
	s := run()
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown")
	}
	log.Println("Server exiting")
}
