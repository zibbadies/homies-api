package routes

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/zibbadies/homies/internal/homies/checks"
	"github.com/zibbadies/homies/internal/homies/db"
	"github.com/zibbadies/homies/internal/homies/models"
)

type ItemInput struct {
	Text string `json:"text"`
}

type ItemsFilter struct {
	Limit int    `form:"limit"`
	From  string `form:"from"`
	To    string `form:"to"`
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

	err = db.NewItem(item.Text, list_id, jwtdata.(jwt.MapClaims)["uid"].(string))
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	c.JSON(200, gin.H{})
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

	fmt.Println("TIME FILTERING: ", fromt, tot)

	items, err := db.GetItems(id_param, fromt, tot, filter.Limit)
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

	err = checks.Check("list_item_text", item.Text)
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

	err = db.UpdateItem(list_id, item_id, item.Text, jwtdata.(jwt.MapClaims)["uid"].(string))
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	c.JSON(200, gin.H{})
}
