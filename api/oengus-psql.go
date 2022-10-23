package api

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/jackc/pgx/v4"
    "os"
    "strings"
)

func GetAcceptedRunnerDiscordIds(marathonId string) ([]string, error) {
    db := getConnection()
    defer closeConnection(db)

    categoryIds, err := getAcceptedCategoryIds(marathonId, db)

    if err != nil {
        fmt.Println("Failed to fetch category ids for", marathonId, err)
        return nil, err
    }

    userIDs, err := getUserIdsForCategoryIds(categoryIds, db)

    if err != nil {
        fmt.Println("Failed to fetch user ids for", marathonId, err)
        return nil, err
    }

    fmt.Println(userIDs)

    strs, _ := json.Marshal(userIDs)
    userIdsParsed := strings.Trim(string(strs), "[]")
    sql := fmt.Sprintf("SELECT username, discord_id FROM users WHERE id IN (%s)", userIdsParsed)

    rows, err := db.Query(context.Background(), sql)

    if err != nil {
        fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
        return nil, err
    }

    var usernameTmp string
    var discordIdTmp string
    var finalDiscordIds []string

    for rows.Next() {
        err := rows.Scan(&usernameTmp, &discordIdTmp)

        if err != nil {
            fmt.Println("Error scanning user rows", err)
            continue
        }

        if discordIdTmp == "" {
            fmt.Println("NO DISCORD ID FOR USER", usernameTmp)
            continue
        } else {
            fmt.Println("Discord id for user", usernameTmp, "is", discordIdTmp)
        }

        finalDiscordIds = append(finalDiscordIds, discordIdTmp)
    }

    return finalDiscordIds, nil
}

func getAcceptedCategoryIds(marathonId string, db *pgx.Conn) ([]int, error) {
    // Todo = 0
    // Rejected = 1
    // Bonus = 2
    // Validated = 3
    // Backup = 4

    sql := "SELECT category_id FROM selection WHERE marathon_id = $1 AND status IN (2, 3, 4)"

    rows, err := db.Query(context.Background(), sql, marathonId)

    if err != nil {
        fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
        return nil, err
    }

    var categoryIds []int
    var categoryId int

    for rows.Next() {
        // Scan is positional, not name based
        err := rows.Scan(&categoryId)

        if err != nil {
            fmt.Println("Error scanning rows", err)
            continue
        }

        categoryIds = append(categoryIds, categoryId)
    }

    fmt.Println(categoryIds)

    return categoryIds, nil
}

func getUserIdsForCategoryIds(categoryIds []int, db *pgx.Conn) ([]int, error) {
    // TODO: KEEP MULTIPLAYER RUNS IN MIND
    // Step 1: main games (hell of a query)
    // https://stackoverflow.com/a/54126847
    strs, _ := json.Marshal(categoryIds)
    catIdsParsed := strings.Trim(string(strs), "[]")
    catSql := fmt.Sprintf("SELECT game_id FROM category WHERE id IN (%s)", catIdsParsed)
    gameSql := fmt.Sprintf("SELECT submission_id FROM game WHERE id IN (%s)", catSql)
    sql := fmt.Sprintf("SELECT user_id FROM submission WHERE id IN (%s)", gameSql)

    fmt.Println(sql)

    rows, err := db.Query(context.Background(), sql)

    if err != nil {
        fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
        return nil, err
    }

    var userIds []int
    var userIdTmp int

    for rows.Next() {
        // Scan is positional, not name based
        err := rows.Scan(&userIdTmp)
        fmt.Println(userIdTmp)

        if err != nil {
            fmt.Println("Error scanning row", err)
            continue
        }

        userIds = append(userIds, userIdTmp)
    }

    // Step 2: searching for opponent users
    sql = fmt.Sprintf(
        "select user_id from submission where id in (SELECT opponent_submission_id FROM opponent WHERE category_id in (%s))",
        catIdsParsed)

    rows, err = db.Query(context.Background(), sql)

    if err != nil {
        fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
        return nil, err
    }

    for rows.Next() {
        // Scan is positional, not name based
        err := rows.Scan(&userIdTmp)
        fmt.Println(userIdTmp)

        if err != nil {
            fmt.Println("Error scanning row 2", err)
            continue
        }

        userIds = append(userIds, userIdTmp)
    }

    return userIds, nil
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
