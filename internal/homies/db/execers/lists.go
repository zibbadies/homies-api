package execers

import (
	"database/sql"
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

func GetItemsEx(exec Execer, listId string, from time.Time, to time.Time, dueOn time.Time, limit int) ([]models.Item, error) {
	b_id, err := strconv.Atoi(listId)
	if (err != nil) { 
		logger.Logger.Error("list ID atoi error", "err", err.Error(), "listId", listId)
		return nil, err
	}

	// NOTE: ORDER BY created_at isn't necessary, maybe make it optional using function arguments?

	args := []any{b_id}
	argsNum := 2

	query := `
		SELECT id, text, completed, author, created_at, due_at
		FROM todos
		WHERE list_id = $1
	`

	if !from.IsZero() {
		query += fmt.Sprintf(" AND created_at >= $%d", argsNum)
		args = append(args, from.UTC())
		argsNum++
	}

	if !to.IsZero() {
		query += fmt.Sprintf(" AND created_at < $%d", argsNum)
		args = append(args, to.UTC())
		argsNum++
	}

	if !dueOn.IsZero() {
		dueOn = dueOn.UTC()
		query += fmt.Sprintf(" AND due_at >= $%d AND due_at < $%d", argsNum, argsNum+1)
		args = append(args, dueOn, dueOn.Add(24*time.Hour))
		argsNum += 2
	}

	query += ` ORDER BY created_at DESC`

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
		var due sql.NullTime;

		if err := rows.Scan(&iid, &item.Text, &item.Completed, &item.Author, &item.CreatedAt, &due); err != nil {
			logger.Logger.Error("list row scan error", "err", err.Error(), "listId", listId)
			return nil, err
		}

		item.Id = strconv.FormatInt(iid, 10);

		if (due.Valid) {
			dueStr := due.Time.UTC().Format(time.RFC3339);
			item.DueAt = &dueStr;
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		logger.Logger.Error("list rows error", "err", err.Error(), "listId", listId)
		return nil, err
	}

	return items, nil;
}

// Returns the UTC days (in [from, to), bounds optional) that have due tasks
// in the house's "todo" list, with total and open (not completed) counts.
func GetDueDaysEx(exec Execer, houseId string, from time.Time, to time.Time) ([]models.DueDay, error) {
	houseIdInt, err := strconv.Atoi(houseId)
	if (err != nil) {
		logger.Logger.Error("house ID atoi error", "err", err.Error(), "houseId", houseId)
		return nil, err
	}

	args := []any{houseIdInt}
	argsNum := 2

	query := `
		SELECT (t.due_at AT TIME ZONE 'UTC')::date AS day,
		       COUNT(*) AS total,
		       COUNT(*) FILTER (WHERE NOT t.completed) AS open
		FROM todos t
		JOIN lists l ON t.list_id = l.id
		WHERE l.house_id = $1 AND l.name = 'todo' AND t.due_at IS NOT NULL
	`

	if !from.IsZero() {
		query += fmt.Sprintf(" AND t.due_at >= $%d", argsNum)
		args = append(args, from.UTC())
		argsNum++
	}

	if !to.IsZero() {
		query += fmt.Sprintf(" AND t.due_at < $%d", argsNum)
		args = append(args, to.UTC())
		argsNum++
	}

	query += ` GROUP BY day ORDER BY day`

	rows, err := exec.Query(query, args...)
	defer rows.Close()

	if err != nil {
		logger.Logger.Error("due days select error", "err", err.Error(), "houseId", houseId)
		return nil, err
	}

	days := make([]models.DueDay, 0);
	for rows.Next() {
		var day time.Time;
		var dueDay models.DueDay;

		if err := rows.Scan(&day, &dueDay.Total, &dueDay.Open); err != nil {
			logger.Logger.Error("due days row scan error", "err", err.Error(), "houseId", houseId)
			return nil, err
		}

		dueDay.Date = day.Format("2006-01-02");
		days = append(days, dueDay)
	}

	if err := rows.Err(); err != nil {
		logger.Logger.Error("due days rows error", "err", err.Error(), "houseId", houseId)
		return nil, err
	}

	return days, nil;
}


func NewItemEx(exec Execer, text string, listId string, authorId string, dueAt *time.Time) (error, string) {
	l_id, err := strconv.Atoi(listId)
	if err != nil {
		logger.Logger.Error("list ID atoi error", "err", err.Error(), "listId", listId)
		return err, ""
	}

	_, err = exec.Exec(`UPDATE lists SET items = items + 1 WHERE id = $1`, l_id)
	if err != nil {
		logger.Logger.Error("list update error", "err", err.Error(), "authorId", authorId)
		return err, ""
	}

	var item_id string;
	err = exec.QueryRow(`
		INSERT INTO todos (text, list_id, author, due_at)
		VALUES ($1, $2, $3, $4) RETURNING id`, text, l_id, authorId, dueAt).Scan(&item_id)
	if err != nil {
		logger.Logger.Error("list insert error", "err", err.Error(), "listId", listId)
		return err, ""
	}

	return nil, item_id
}

func UpdateItemEx(exec Execer, listId string, itemId string, text string, authorId string, touchDue bool, dueAt *time.Time) error {
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

	if (touchDue) {
		_, err = exec.Exec(`UPDATE todos SET text = $1, author = $2, due_at = $3 WHERE (id = $4 AND list_id = $5)`, text, authorId, dueAt, i_id, l_id)
	} else {
		_, err = exec.Exec(`UPDATE todos SET text = $1, author = $2 WHERE (id = $3 AND list_id = $4)`, text, authorId, i_id, l_id)
	}
	if err != nil {
		logger.Logger.Error("list item update error", "err", err.Error(), "itemId", itemId)
		return err
	}

	return nil
}

func SetItemDueEx(exec Execer, listId string, itemId string, dueAt *time.Time) error {
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

	_, err = exec.Exec(`UPDATE todos SET due_at = $1 WHERE (id = $2 AND list_id = $3)`, dueAt, i_id, l_id)
	if err != nil {
		logger.Logger.Error("list item due update error", "err", err.Error(), "itemId", itemId)
		return err
	}

	return nil
}

func SetItemCompletedEx(exec Execer, listId string, itemId string, completed bool) error {
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

	_, err = exec.Exec(`UPDATE todos SET completed = $1 WHERE (id = $2 AND list_id = $3)`, completed, i_id, l_id)
	if err != nil {
		logger.Logger.Error("list item update error", "err", err.Error(), "itemId", itemId)
		return err
	}

	return nil
}

func DeleteItemEx(exec Execer, listId string, itemId string) error {
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

	_, err = exec.Exec(`DELETE FROM todos WHERE (id = $1 AND list_id = $2)`, i_id, l_id)
	if err != nil {
		logger.Logger.Error("list item update error", "err", err.Error(), "itemId", itemId)
		return err
	}

	return nil
}
