package execers

import (
	"strconv"
	"crypto/rand"
	"math/big"
	"database/sql"
	"github.com/lib/pq"

	"github.com/zibbadies/homies/internal/homies/logger"
	"github.com/zibbadies/homies/internal/homies/models"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateCode() (string, error) {
	code := make([]byte, 6)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[num.Int64()]
	}
	return string(code), nil
}

// Returns the house ID & invite code
// NO EXECER FOR THIS ONE
func NewHouseEx(exec *sql.DB, name string, owner string) (string, string, error) {
    tx, err := exec.Begin()
    if err != nil {return "", "", err}
    defer func (){
		if p := recover(); p != nil {
			tx.Rollback()
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	var invite string
	var houseId int64
	for range 5 {
		invite, err = GenerateCode()
		if (err != nil) {
			logger.Logger.Error("house insert error", "err", err.Error())
			return "", "", err
		}

		err = tx.QueryRow(`
			INSERT INTO houses (name, owner, invite)
			VALUES ($1, $2, $3)
			RETURNING id`, name, owner, invite,
		).Scan(&houseId)

		if err == nil {
			break;
		}


		if pqErr, ok := err.(*pq.Error); ok {
			if (pqErr.Code == "23505") {
				logger.Logger.Error("invite duplicate!?", "err", err.Error())
				continue
			}
		}

		logger.Logger.Error("house insert error", "err", err.Error())
		break;
	}

	houseIdStr := strconv.FormatInt(houseId, 10)

	err = NewListEx(tx, houseIdStr, "shopping", true);
	if err != nil {
		logger.Logger.Error("shopping list insert error", "err", err.Error())
		return "", "", err
	}

	err = NewListEx(tx, houseIdStr, "todo", true);
	if err != nil {
		logger.Logger.Error("todo list insert error", "err", err.Error())
		return "", "", err
	}

	return strconv.FormatInt(houseId, 10), invite, nil;
}

func GetHouseEx(exec Execer, house string, skipUser string) (models.House, error) {
	var resHouse models.House
	var houseid int64

	logger.Logger.Info("get house", "house", house)
	err := exec.QueryRow("SELECT id, invite, name, owner FROM houses WHERE id = $1", house).Scan(&houseid, &resHouse.Invite, &resHouse.Name, &resHouse.Owner)
	if (err != nil) {
		logger.Logger.Error("house get error", "err", err.Error())
		return models.House{}, err
	}

	// Retrive house Members
	// Why this must be a mess every fucking time?

	query := `
		SELECT id, name, bg_color, face_color, face_x, face_y, 
			   left_eye_x, left_eye_y, right_eye_x, right_eye_y, bezier
		FROM users
		WHERE house = $1
	`
	args := []any{houseid}

	if skipUser != "" {
		query += " AND id != $2"
		args = append(args, skipUser)
	}

	rows, err := exec.Query(query, args...)

	//rows, err := exec.Query(`SELECT id, name, bg_color, face_color, face_x, face_y, left_eye_x, left_eye_y, right_eye_x, right_eye_y, bezier FROM users WHERE house = ?`, houseid);
	defer rows.Close()

	if err != nil {
		logger.Logger.Error("list DB select error", "err", err.Error(), "houseId", houseid)
		return models.House{}, err
	}

	var users = make([]models.User, 0);
	for rows.Next() {
		var user models.User;

		if err := rows.Scan(&user.UID, &user.Username, &user.Avatar.BgColor, &user.Avatar.FaceColor, &user.Avatar.FaceX, &user.Avatar.FaceY, &user.Avatar.LeX, &user.Avatar.LeY, &user.Avatar.ReX, &user.Avatar.ReY, &user.Avatar.Bezier); err != nil {
			logger.Logger.Error("list row scan error", "err", err.Error(), "houseId", houseid)
			return models.House{}, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		logger.Logger.Error("list rows error", "err", err.Error(), "houseId", houseid)
		return models.House{}, err
	}

	resHouse.Members = users

	return resHouse, nil;
}

