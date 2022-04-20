package postgresql

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"time"
)

func Ping() {
	db, err := sql.Open("postgres",
		"db.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		panic(err)
	}

}
