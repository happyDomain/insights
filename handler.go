package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"git.happydns.org/happyDomain/model"
)

func handler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data happydns.Insights

		err := decodeJSONBody(w, r, &data)
		if err != nil {
			var mr *malformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.msg, mr.status)
			} else {
				log.Printf("error decoding payload: %s", err.Error())
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			}
			return
		}

		err = saveToDB(db, data)
		if err != nil {
			log.Printf("Error handling request: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
