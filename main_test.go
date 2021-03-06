package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestHandleFormFunc_POST_Success(t *testing.T) {
	w := httptest.NewRecorder()
	data := url.Values{}
	data.Set("first_name", "John")
	data.Set("last_name", "Doe")
	data.Set("email", "email@example.com")
	data.Set("phone_number", "0819999999")
	req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handleFormFunc(w, req)
	if w.Code != http.StatusOK {
		t.Fail()
	}
	err := ioutil.WriteFile(dataFile, []byte("[]"), os.ModeAppend)
	if err != nil {
		fmt.Printf("Failed to write to the file %s\n%s", dataFile, err.Error())
	}
}

func TestHandleFormFunc_GET_Success(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	handleFormFunc(w, req)
	if w.Code != http.StatusOK {
		t.Fail()
	}
}

func TestHandleFormFunc_GET_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/", nil)
	handleFormFunc(w, req)
	if w.Code != http.StatusNotFound {
		t.Fail()
	}
}

func TestHandleDataFunc_GET_Success(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/data", nil)
	handleDataFunc(w, req)
	if w.Code != http.StatusNotFound {
		t.Fail()
	}
}

func TestHandleDataFunc_POST_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/data", nil)
	handleDataFunc(w, req)
	if w.Code != http.StatusNotFound {
		t.Fail()
	}
}

func TestHandleDataFunc_PUT_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/data", nil)
	handleDataFunc(w, req)
	if w.Code != http.StatusNotFound {
		t.Fail()
	}
}

func TestHandleDataFunc_DELETE_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/data", nil)
	handleDataFunc(w, req)
	if w.Code != http.StatusNotFound {
		t.Fail()
	}
}

func TestRun(t *testing.T) {
	oLoadEnv := loadEnv
	loadEnv = func(filename ...string) (err error) {
		err = os.Setenv("PORT", "8080")
		if err != nil {
			fmt.Printf("Failed to set environment vriable PORT\n%s", err.Error())
		}
		return
	}
	defer func() {
		loadEnv = oLoadEnv
		r := recover()
		if r != nil {
			t.Fail()
		}
	}()
	srv := run()
	time.Sleep(1 * time.Second)
	err := srv.Shutdown(context.TODO())
	if err != nil {
		fmt.Printf("Failed to shut down\n%s", err.Error())
	}
}
