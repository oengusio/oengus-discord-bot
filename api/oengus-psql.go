package api

import (
    "context"
    "fmt"
    "github.com/jackc/pgx/v4"
    "os"
)

func GetAcceptedRunnerDiscordIds(marathonId string) ([]string, error) {
    db := getConnection()
    defer closeConnection(db)

    getAcceptedCategoryIds(marathonId, db)

    sql := "SELECT * FROM users WHERE id IN ($1)"
    userIds := []string{"aa"}

    _, err := db.Query(context.Background(), sql, userIds)

    if err != nil {
        fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
        return nil, err
    }

    return []string{"BLA BLA DISCORD IDS"}, nil
}

func getAcceptedCategoryIds(marathonId string, db *pgx.Conn) {
    //
}

func getConnection() *pgx.Conn {
    conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URI"))

    if err != nil {
        fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
        os.Exit(1)
    }

    return conn
}

func closeConnection(db *pgx.Conn) {
    err := db.Close(context.Background())

    if err != nil {
        fmt.Println(err)
    }
}
