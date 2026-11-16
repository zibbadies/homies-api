package routes

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/zibbadies/homies/internal/homies/checks"
	"github.com/zibbadies/homies/internal/homies/db"
	"github.com/zibbadies/homies/internal/homies/models"
)

type ItemInput struct {
	Text      string          `json:"text"`
	Completed bool            `json:"completed"`
	DueAt     json.RawMessage `json:"due_at"`
}

type ItemsFilter struct {
	Limit int    `form:"limit"`
	From  string `form:"from"`
	To    string `form:"to"`
	DueOn string `form:"due_on"`
}

// Parses the optional "due_at" field of an ItemInput.
// Returns (touchDue, dueAt, err):
//   - field absent  -> (false, nil, nil)   : leave the stored value untouched
//   - field null    -> (true, nil, nil)    : clear the due date
//   - field present -> (true, &time, nil)  : set the due date (RFC3339)
func parseDueAt(raw json.RawMessage) (bool, *time.Time, error) {
	if raw == nil {
		return false, nil, nil
	}

	if strings.TrimSpace(string(raw)) == "null" {
		return true, nil, nil
	}

	var dueStr string;
	if err := json.Unmarshal(raw, &dueStr); err != nil {
		return false, nil, &models.CheckError{
			Message:   "Your due date must be an RFC3339 string (e.g. 2026-11-15T18:00:00Z) or null!",
			ErrorCode: models.BasicCheckError,
		}
	}

	if err := checks.Check("list_item_due", dueStr); err != nil {
		return false, nil, err
	}

	due, err := time.Parse(time.RFC3339, dueStr)
	if err != nil {
		return false, nil, &models.CheckError{
			Message:   "Your due date is not a valid RFC3339 timestamp!",
			ErrorCode: models.BasicCheckError,
		}
	}

	due = due.UTC()
	return true, &due, nil
}

func getLists(c *gin.Context) {
	jwtdata, _ := c.Get("data")

	dbuser, err := db.GetUserMe(jwtdata.(jwt.MapClaims)["uid"].(string))
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	lists, err := db.GetLists(dbuser.HouseId)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	c.JSON(200, gin.H{
		"lists": lists,
	})
}

func newItem(c *gin.Context) {
	jwtdata, _ := c.Get("data")

	var item ItemInput
	err := c.ShouldBind(&item)
	if err != nil {
		c.JSON(400, gin.H{"error": models.DBError{
			Message:   "Invalid JSON Data!",
			ErrorCode: models.JsonFormatError,
		}})
		return
	}

	err = checks.Check("list_item_text", item.Text)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	_, dueAt, err := parseDueAt(item.DueAt)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	list_id := c.Param("id")

	list_hid, err := db.GetListHID(list_id)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	dbuser, err := db.GetUserMe(jwtdata.(jwt.MapClaims)["uid"].(string))
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	if dbuser.HouseId != list_hid {
		c.JSON(403, gin.H{"error": models.DBError{
			Message:   "You can't access other people lists!",
			ErrorCode: models.NotAuthorized,
		}})
		return
	}

	err, id := db.NewItem(item.Text, list_id, jwtdata.(jwt.MapClaims)["uid"].(string), dueAt)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	c.JSON(200, gin.H{"id": id})
}

func getItems(c *gin.Context) {
	jwtdata, _ := c.Get("data")
	id_param := c.Param("id")

	var filter ItemsFilter
	err := c.ShouldBind(&filter)
	if err != nil {
		c.JSON(400, gin.H{"error": models.DBError{
			Message:   "Invalid JSON Data!",
			ErrorCode: models.JsonFormatError,
		}})
		return
	}

	list_hid, err := db.GetListHID(id_param)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	dbuser, err := db.GetUserMe(jwtdata.(jwt.MapClaims)["uid"].(string))
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	if dbuser.HouseId != list_hid {
		c.JSON(403, gin.H{"error": models.DBError{
			Message:   "You can't access other people lists!",
			ErrorCode: models.NotAuthorized,
		}})
		return
	}

	fmt.Println(filter)

	if filter.Limit > 10 || filter.Limit == 0 {
		filter.Limit = 10
	}

	// TODO: Finish time filtering

	fromt, err := time.Parse(time.RFC3339, filter.From)
	if err != nil {
		fmt.Println(err)
		fromt = models.NoTime
	}

	tot, err := time.Parse(time.RFC3339, filter.To)
	if err != nil {
		fmt.Println(err)
		tot = models.NoTime
	}

	// due_on=YYYY-MM-DD filters items whose due_at falls in that UTC day
	dueOn, err := time.Parse("2006-01-02", filter.DueOn)
	if err != nil {
		dueOn = models.NoTime
	}

	fmt.Println("TIME FILTERING: ", fromt, tot, dueOn)

	items, err := db.GetItems(id_param, fromt, tot, dueOn, filter.Limit)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	c.JSON(200, gin.H{"items": items})
}

func updateItem(c *gin.Context) {
	jwtdata, _ := c.Get("data")
	list_id := c.Param("id")
	item_id := c.Param("item_id")

	var item ItemInput
	err := c.ShouldBind(&item)
	if err != nil {
		c.JSON(400, gin.H{"error": models.DBError{
			Message:   "Invalid JSON Data!",
			ErrorCode: models.JsonFormatError,
		}})
		return
	}

	touchDue, dueAt, err := parseDueAt(item.DueAt)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	/*
		Checks:
		- List ID in user's house (list_hid == user_hid)
		- Item ID in list (item_lid == list_id) ( db side? )
	*/

	dbuser, err := db.GetUserMe(jwtdata.(jwt.MapClaims)["uid"].(string))
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	list_hid, err := db.GetListHID(list_id)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	if dbuser.HouseId != list_hid {
		c.JSON(403, gin.H{"error": models.DBError{
			Message:   "You can't access this list!",
			ErrorCode: models.NotAuthorized,
		}})
		return
	}

	fmt.Println("--------")
	fmt.Println(item)

	if (item.Text == "") {
		err = db.SetItemCompleted(list_id, item_id, item.Completed)
		if err != nil {
			c.JSON(400, gin.H{"error": err})
			return
		}

		if (touchDue) {
			err = db.SetItemDue(list_id, item_id, dueAt)
			if err != nil {
				c.JSON(400, gin.H{"error": err})
				return
			}
		}
	} else {
		err = checks.Check("list_item_text", item.Text)
		if err != nil {
			c.JSON(400, gin.H{"error": err})
			return
		}

		err = db.UpdateItem(list_id, item_id, item.Text, jwtdata.(jwt.MapClaims)["uid"].(string), touchDue, dueAt)
		if err != nil {
			c.JSON(400, gin.H{"error": err})
			return
		}
	}

	c.JSON(200, gin.H{})
}

func deleteItem(c *gin.Context) {
	jwtdata, _ := c.Get("data")
	list_id := c.Param("id")
	item_id := c.Param("item_id")

	/*
		Checks:
		- List ID in user's house (list_hid == user_hid)
		- Item ID in list (item_lid == list_id) ( db side? )
	*/

	dbuser, err := db.GetUserMe(jwtdata.(jwt.MapClaims)["uid"].(string))
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	list_hid, err := db.GetListHID(list_id)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	if dbuser.HouseId != list_hid {
		c.JSON(403, gin.H{"error": models.DBError{
			Message:   "You can't access this list!",
			ErrorCode: models.NotAuthorized,
		}})
		return
	}

	err = db.DeleteItem(list_id, item_id)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	c.JSON(200, gin.H{})
}


// GET /user/:id/due-days?from=YYYY-MM-DD&to=YYYY-MM-DD
// Calendar preview: UTC days that have due tasks in the house's "todo" list,
// with total/open counts. Bounds are optional, "to" is exclusive.
// The :id param is ignored, the house is taken from the JWT (same as homeOverview).
func dueDays(c *gin.Context) {
	jwtdata, _ := c.Get("data")

	user, err := db.GetUserMe(jwtdata.(jwt.MapClaims)["uid"].(string))
	if (err != nil) {
		c.JSON(400, gin.H{"error": err})
		return
	}

	from, err := time.Parse("2006-01-02", c.Query("from"))
	if (err != nil) {
		from = models.NoTime
	}

	to, err := time.Parse("2006-01-02", c.Query("to"))
	if (err != nil) {
		to = models.NoTime
	}

	days, err := db.GetDueDays(user.HouseId, from, to)
	if (err != nil) {
		c.JSON(400, gin.H{"error": err})
		return
	}

	c.JSON(200, gin.H{"days": days})
}
