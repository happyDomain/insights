package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"git.happydns.org/happyDomain/model"
	_ "github.com/go-sql-driver/mysql"
)

// DSNGenerator returns DSN filed with values from environment
func DSNGenerator() string {
	db_user := "hd_insights"
	db_password := "hd_insights"
	db_host := ""
	db_db := "hd_insights"

	if v, exists := os.LookupEnv("MYSQL_HOST"); exists {
		if strings.HasPrefix(v, "/") {
			db_host = "unix(" + v + ")"
		} else {
			db_host = "tcp(" + v + ":"
			if p, exists := os.LookupEnv("MYSQL_PORT"); exists {
				db_host += p + ")"
			} else {
				db_host += "3306)"
			}
		}
	}
	if v, exists := os.LookupEnv("MYSQL_PASSWORD"); exists {
		db_password = v
	} else if v, exists := os.LookupEnv("MYSQL_ROOT_PASSWORD"); exists {
		db_user = "root"
		db_password = v
	}
	if v, exists := os.LookupEnv("MYSQL_USER"); exists {
		db_user = v
	}
	if v, exists := os.LookupEnv("MYSQL_DATABASE"); exists {
		db_db = v
	}

	return db_user + ":" + db_password + "@" + db_host + "/" + db_db + "?parseTime=true"
}

func openDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}

	// Create table if not exists
	createTableQuery := `
CREATE TABLE IF NOT EXISTS insights (
	id VARCHAR(255) NOT NULL,
	time DATETIME default CURRENT_TIMESTAMP,
	data JSON,
	PRIMARY KEY (id, time)
);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	return db, nil
}

func saveToDB(db *sql.DB, data happydns.Insights) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	query := `INSERT INTO insights (id, data) VALUES (?, ?)`
	_, err = db.Exec(query, data.InsightsID, dataJSON)
	return err
}

func purgeOldEntries(db *sql.DB) error {
	// Delete entries older than 30 days
	query := `DELETE FROM insights WHERE time < ?`
	cnt, err := db.Exec(query, time.Now().Add(-30*24*time.Hour))
	if err != nil {
		return err
	}
	deleted, _ := cnt.RowsAffected()
	log.Printf("Deleted %d old entries\n", deleted)
	return nil
}
