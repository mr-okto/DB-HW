package handlers

import (
	"db-hw/internal/db"
	w "db-hw/internal/writers"
	"net/http"
)

func GetInfo(res http.ResponseWriter, _ *http.Request) {
	stats, err, status := db.GetDBCountData()
	if err != nil {
		w.WriteError(res, status, err.Error())
		return
	}
	w.WriteEasyJson(res, status, stats)
}

func ClearData(res http.ResponseWriter, _ *http.Request) {
	db.ClearDB()
	res.WriteHeader(http.StatusOK)
}
