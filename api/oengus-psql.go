package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"os"
	"strings"
)

// TODO: https://pkg.go.dev/golang.org/x/exp/slices#Contains

func GetMarathonName(code string) (string, error) {
	db := getConnection()
	defer closeConnection(db)

	sql := "SELECT name FROM marathon WHERE id = $1"

	rows, err := db.Query(context.Background(), sql, code)

	if err != nil {
		return "", err
	}

	if rows.Next() {
		var name string

		// Scan is positional, not name based
		err := rows.Scan(&name)

		if err != nil {
			return "", err
		}

		return name, nil
	}

	return "", errors.New("database lookup failed")
}

func GetUserProfile(userId int) (ProfileDto, error) {
	db := getConnection()
	defer closeConnection(db)

	var profile ProfileDto

	sql := "SELECT id, username FROM users WHERE id = $1"

	rows, err := db.Query(context.Background(), sql, userId)

	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return profile, err
	}

	if rows.Next() {
		var userId int
		var username string

		// Scan is positional, not name based
		err := rows.Scan(&userId, &username)

		if err != nil {
			fmt.Println("Error scanning rows", err)
			return profile, err
		}

		profile = ProfileDto{
			Id:       userId,
			Username: username,
		}
	}

	return profile, nil
}

func GetGameById(gameId int) (GameDto, error) {
	db := getConnection()
	defer closeConnection(db)

	var game GameDto

	sql := "SELECT id, name, console FROM game WHERE id = $1"

	rows, err := db.Query(context.Background(), sql, gameId)

	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return game, err
	}

	if rows.Next() {
		var id int
		var name string
		var console string

		// Scan is positional, not name based
		err := rows.Scan(&id, &name, &console)

		if err != nil {
			fmt.Println("Error scanning rows", err)
			return game, err
		}

		game = GameDto{
			Id:      id,
			Name:    name,
			Console: console,
		}
	}

	return game, nil
}

func GetMarathonSelectionDone(marathonId string) (bool, error) {
	db := getConnection()
	defer closeConnection(db)

	sql := "SELECT is_selection_done FROM marathon WHERE id = $1"

	rows, err := db.Query(context.Background(), sql, marathonId)

	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return false, err
	}

	var isSelectionDone bool

	if rows.Next() {
		// Scan is positional, not name based
		err := rows.Scan(&isSelectionDone)

		if err != nil {
			fmt.Println("Error scanning rows", err)
			return false, err
		}
	}

	return isSelectionDone, nil
}

func GetModeratorsForMarathon(marathonId string) ([]string, error) {
	db := getConnection()
	defer closeConnection(db)

	sql := "SELECT discord_id FROM users WHERE id IN (SELECT user_id FROM moderator WHERE marathon_id = $1) OR id = (SELECT creator_id FROM marathon WHERE id = $1)"

	rows, err := db.Query(context.Background(), sql, marathonId)

	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return nil, err
	}

	var discordIdList []string
	var discordId string

	for rows.Next() {
		// Scan is positional, not name based
		err := rows.Scan(&discordId)

		if err != nil {
			fmt.Println("Error scanning rows", err)
			continue
		}

		discordIdList = append(discordIdList, discordId)
	}

	return discordIdList, nil
}

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

	// TODO: remove dupes

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
