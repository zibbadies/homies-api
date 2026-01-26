package execers

import (
	"fmt"
	"time"

	"github.com/zibbadies/homies/internal/homies/logger"
	"github.com/zibbadies/homies/internal/homies/models"

	"strconv"
)

func NewListEx(exec Execer, houseId string, name string, inOverview bool) error {
	houseIdInt, err := strconv.Atoi(houseId)
	if (err != nil) {
		logger.Logger.Error("list ID atoi error", "err", err.Error(), "houseId", houseId)
		return err
	}

	_, err = exec.Exec(`
		INSERT INTO lists (house_id, name, in_overview)
		VALUES ($1, $2, $3)`,
		houseIdInt, name, inOverview,
	)
	if err != nil {return err;}

	return nil;
}

func GetListsEx(exec Execer, houseId string) ([]models.List, error) {
	houseIdInt, err := strconv.Atoi(houseId)
	if (err != nil) {
		logger.Logger.Error("list ID atoi error", "err", err.Error(), "houseId", houseId)
		return nil, err
	}

	rows, err := exec.Query(`SELECT id, name FROM lists WHERE house_id = $1`, houseIdInt);
	defer rows.Close()

	if (err != nil) {
		logger.Logger.Error("list get error", "err", err.Error())
		return nil, err
	}

	var lists []models.List;
	for rows.Next() {
		var list models.List;
		var id uint;

		if err := rows.Scan(&id, &list.Name); err != nil {
            logger.Logger.Error("list get error", "err", err.Error())
			return nil, err
        }
		list.Id = strconv.FormatUint(uint64(id), 10);
		lists = append(lists, list)
	}

    if err := rows.Err(); err != nil {
		logger.Logger.Error("list get error", "err", err.Error())
		return nil, err
    }

	return lists, nil;
}

func GetListHIDEx(exec Execer, listId string) (string, error) {
	var houseId int64;
	
	err := exec.QueryRow(`SELECT house_id FROM lists WHERE id = $1`, listId).Scan(&houseId);
	if (err != nil) {
		logger.Logger.Error("user house ID retrival error", "err", err.Error())
		return "", err
	}

	return strconv.FormatInt(houseId, 10), nil;
}

func GetItemsEx(exec Execer, listId string, from time.Time, to time.Time, limit int) ([]models.Item, error) {
	b_id, err := strconv.Atoi(listId)
	if (err != nil) { 
		logger.Logger.Error("list ID atoi error", "err", err.Error(), "listId", listId)
		return nil, err
	}

	// NOTE: ORDER BY created_at isn't necessary, maybe make it optional using function arguments?

	args := []any{b_id}
	argsNum := 2

	query := `
		SELECT id, text, completed, author, created_at
		FROM todos
		WHERE list_id = $1
	`

	if !from.IsZero() {
		query += fmt.Sprintf(" AND created_at >= $%d", argsNum)
		args = append(args, from.UTC())
		argsNum++
	}

	if !to.IsZero() {
		query += fmt.Sprintf(" AND created_at <= $%d", argsNum)
		args = append(args, to.UTC())
		argsNum++
	}

	query += ` ORDER BY created_at`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argsNum)
		args = append(args, limit)
		argsNum++
	}

	rows, err := exec.Query(query, args...)
	defer rows.Close()

	if err != nil {
		logger.Logger.Error("list DB select error", "err", err.Error(), "listId", listId)
		return nil, err
	}

	items := make([]models.Item, 0);
	for rows.Next() {
		var item models.Item;
		var iid int64;
		var createdAt time.Time; // TODO: implement this in &item.

		if err := rows.Scan(&iid, &item.Text, &item.Completed, &item.Author, &createdAt); err != nil {
			logger.Logger.Error("list row scan error", "err", err.Error(), "listId", listId)
			return nil, err
		}

		fmt.Println(createdAt)
		item.Id = strconv.FormatInt(iid, 10);

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		logger.Logger.Error("list rows error", "err", err.Error(), "listId", listId)
		return nil, err
	}

	return items, nil;
}


func NewItemEx(exec Execer, text string, listId string, authorId string) error {
	l_id, err := strconv.Atoi(listId)
	if err != nil {
		logger.Logger.Error("list ID atoi error", "err", err.Error(), "listId", listId)
		return err
	}

	_, err = exec.Exec(`UPDATE lists SET items = items + 1 WHERE id = $1`, l_id)
	if err != nil {
		logger.Logger.Error("list update error", "err", err.Error(), "authorId", authorId)
		return err
	}

	_, err = exec.Exec(`
		INSERT INTO todos (text, list_id, author)
		VALUES ($1, $2, $3)`, text, l_id, authorId)
	if err != nil {
		logger.Logger.Error("list insert error", "err", err.Error(), "listId", listId)
		return err
	}

	return nil
}

func UpdateItemEx(exec Execer, listId string, itemId string, text string, authorId string) error {
	i_id, err := strconv.Atoi(itemId)
	if err != nil {
		logger.Logger.Error("list item ID atoi error", "err", err.Error(), "listId", itemId)
		return err
	}

	l_id, err := strconv.Atoi(listId)
	if err != nil {
		logger.Logger.Error("list ID atoi error", "err", err.Error(), "listId", listId)
		return err
	}

	_, err = exec.Exec(`UPDATE todos SET text = $1, author = $2 WHERE (id = $3 AND list_id = $4)`, text, authorId, i_id, l_id)
	if err != nil {
		logger.Logger.Error("list item update error", "err", err.Error(), "itemId", itemId)
		return err
	}

	return nil
}

