package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ameske/csvtoprism"
)

func main() {
	http.HandleFunc("/upload", ImportFile)
	http.HandleFunc("/generate", CreateCSV)
	log.Fatal(http.ListenAndServe("localhost:41586", nil))
}

type RawExperiment struct {
	Name    string                 `json:"name"`
	Samples []csvtoprism.RawSample `json:"experiment"`
}

// ImportFile accepts a raw XLS file and returns the representation of the data
func ImportFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	file, header, err := r.FormFile("xlsfile")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer

	n, err := io.Copy(&buf, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Importing file:", header.Filename)

	samples, err := csvtoprism.ParseRawSamplesXLS(bytes.NewReader(buf.Bytes()), n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	parts := strings.Split(header.Filename, ".")
	if len(parts) != 2 {
		http.Error(w, "malformed experiment file, must be of type <name>.xlsx", http.StatusBadRequest)
		return
	}

	e := RawExperiment{
		Name:    parts[0],
		Samples: samples,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")

	err = enc.Encode(&e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CSV takes in the JSON representation of an experiment and returns links
// to the two resulting files
func CreateCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Add("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		w.Header().Add("Access-Control-Allow-Headers", "Origin, Content-Type, X-Auth-Token")
		w.Header().Add("Access-Control-Allow-Origin", "*")
	} else {
		w.Header().Add("Access-Control-Allow-Origin", "*")
	}

	var exp csvtoprism.Experiment

	err := json.NewDecoder(r.Body).Decode(&exp)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	adjusted := exp.Adjusted()

	originalFD, err := os.Create(exp.Name + ".csv")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	adjustedFD, err := os.Create(adjusted.Name + ".csv")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = exp.WriteCSV(originalFD)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Created: " + exp.Name + ".csv")

	err = adjusted.WriteCSV(adjustedFD)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Created: " + adjusted.Name + ".csv")

	w.Write([]byte("Your experiements have been succesfully converted to CSV. Please check your home directory\n"))
}
