package db

import (
	"database/sql"
	"time"

	"github.com/lib/pq"

	"github.com/zibbadies/homies/internal/homies/db/execers"
	"github.com/zibbadies/homies/internal/homies/logger"
	"github.com/zibbadies/homies/internal/homies/models"
)

func NewList(houseId string, name string, inOverview bool) error {
	err := execers.NewListEx(db, houseId, name, inOverview)
	if err == nil {
		return nil
	}

	if pqErr, ok := err.(*pq.Error); ok {
		logger.Logger.Error("list insert error", "err", err.Error(), "sql_err", pqErr.Code)
	} else {
		logger.Logger.Error("list insert error", "err", err.Error())
	}

	return &models.DBError{
		Message:   "General error, please try again later!",
		ErrorCode: models.InternalError,
	}
}

func GetLists(houseId string) ([]models.List, error) {
	lists, err := execers.GetListsEx(db, houseId)
	if err == nil {
		return lists, nil
	}

	if pqErr, ok := err.(*pq.Error); ok {
		logger.Logger.Error("lists get error", "err", err.Error(), "sql_err", pqErr.Code)
	} else {
		logger.Logger.Error("lists get error", "err", err.Error())
	}

	return []models.List{}, &models.DBError{
		Message:   "General error, please try again later!",
		ErrorCode: models.InternalError,
	}
}

func GetListHID(listId string) (string, error) {
	houseId, err := execers.GetListHIDEx(db, listId)
	if err == nil {
		return houseId, nil
	}

	if err == sql.ErrNoRows {
		return "", &models.DBError{
			Message:   "We didn't find your list in the database!",
			ErrorCode: models.ListNotFound,
		}
	}

	if pqErr, ok := err.(*pq.Error); ok {
		logger.Logger.Error("list HID get error", "err", err.Error(), "sql_err", pqErr.Code)
	} else {
		logger.Logger.Error("list HID get error", "err", err.Error())
	}

	return "", &models.DBError{
		Message:   "General error, please try again later!",
		ErrorCode: models.InternalError,
	}
}

func GetItems(listId string, from time.Time, to time.Time, limit int) ([]models.Item, error) {
	items, err := execers.GetItemsEx(db, listId, from, to, limit)
	if err == nil {
		return items, nil
	}

	if pqErr, ok := err.(*pq.Error); ok {
		logger.Logger.Error("items get error", "err", err.Error(), "sql_err", pqErr.Code)
	} else {
		logger.Logger.Error("items get error", "err", err.Error())
	}

	return []models.Item{}, &models.DBError{
		Message:   "General error, please try again later!",
		ErrorCode: models.InternalError,
	}
}

func NewItem(text string, listId string, authorId string) (error, string) {
	err, id := execers.NewItemEx(db, text, listId, authorId)
	if err == nil {
		return nil, id
	}

	if pqErr, ok := err.(*pq.Error); ok {
		logger.Logger.Error("item insert error", "err", err.Error(), "sql_err", pqErr.Code)
	} else {
		logger.Logger.Error("item insert error", "err", err.Error())
	}

	return &models.DBError{
		Message:   "General error, please try again later!",
		ErrorCode: models.InternalError,
	}, ""
}

func UpdateItem(listId string, itemId string, text string, authorId string) error {
	err := execers.UpdateItemEx(db, listId, itemId, text, authorId)
	if err == nil {
		return nil
	}

	if err == sql.ErrNoRows {
		return &models.DBError{
			Message: "Item not found!",
			ErrorCode: models.ItemNotFound,
		}
	}

	if pqErr, ok := err.(*pq.Error); ok {
		logger.Logger.Error("item update error", "err", err.Error(), "sql_err", pqErr.Code)
	} else {
		logger.Logger.Error("item update error", "err", err.Error())
	}

	return &models.DBError{
		Message:   "General error, please try again later!",
		ErrorCode: models.InternalError,
	}
}

func SetItemCompleted(listId string, itemId string, completed bool) error {
	err := execers.SetItemCompletedEx(db, listId, itemId, completed)
	if err == nil {
		return nil
	}

	if pqErr, ok := err.(*pq.Error); ok {
		logger.Logger.Error("item update completed error", "err", err.Error(), "sql_err", pqErr.Code)
	} else {
		logger.Logger.Error("item update completed error", "err", err.Error())
	}

	return &models.DBError{
		Message:   "General error, please try again later!",
		ErrorCode: models.InternalError,
	}
}

func DeleteItem(listId string, itemId string) error {
	err := execers.DeleteItemEx(db, listId, itemId)
	if err == nil {
		return nil
	}

	if pqErr, ok := err.(*pq.Error); ok {
		logger.Logger.Error("item delete error", "err", err.Error(), "sql_err", pqErr.Code)
	} else {
		logger.Logger.Error("item delete error", "err", err.Error())
	}

	return &models.DBError{
		Message:   "General error, please try again later!",
		ErrorCode: models.InternalError,
	}
}
