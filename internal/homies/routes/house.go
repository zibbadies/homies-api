package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/zibbadies/homies/internal/homies/checks"
	"github.com/zibbadies/homies/internal/homies/db"
	"github.com/zibbadies/homies/internal/homies/models"
)

type JHouse struct {
	Name string `json:"name"`
}

func createHouse(c *gin.Context) {
	jwtdata, _ := c.Get("data")

	user, err := db.GetUserMe(jwtdata.(jwt.MapClaims)["uid"].(string))
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	if user.HouseId != "null" {
		c.JSON(400, gin.H{"error": models.DBError{
			Message:   "You already have a house!",
			ErrorCode: models.UserInHouse,
		}})
		return
	}

	var house JHouse
	err = c.ShouldBind(&house)
	if err != nil {
		c.JSON(400, gin.H{"error": models.DBError{
			Message:   "Invalid JSON Data!",
			ErrorCode: models.JsonFormatError,
		}})
		return
	}

	err = checks.Check("house_name", house.Name)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	houseid, invite, err := db.NewHouse(house.Name, jwtdata.(jwt.MapClaims)["uid"].(string))
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	err = db.ChangeHouse(jwtdata.(jwt.MapClaims)["uid"].(string), houseid)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	c.JSON(200, gin.H{
		"invite": invite,
	})
}

func userHouse(c *gin.Context) {
	if c.Param("id") != "me" {
		c.JSON(403, gin.H{"error": models.DBError{
			Message:   "You can't see other people houses!",
			ErrorCode: models.NotAuthorized,
		}})
		return
	}

	jwtdata, _ := c.Get("data")

	house, err := db.GetUserHouse(jwtdata.(jwt.MapClaims)["uid"].(string))
	if err != nil {
		if dberr, ok := err.(models.DBError); ok {
			if dberr.ErrorCode == models.UserNotFound {
				c.JSON(404, gin.H{"error": &models.DBError{
					Message:   "Your user was not found in the database!",
					ErrorCode: models.NotAuthenticated,
				}})
				return
			}
		}
		c.JSON(400, gin.H{"error": err})
		return
	}

	c.JSON(200, gin.H{
		"name":    house.Name,
		"invite":  house.Invite,
		"members": house.Members,
	})
}

func joinHouse(c *gin.Context) {
	jwtdata, _ := c.Get("data")
	invite := c.Param("invite")

	user, err := db.GetUserMe(jwtdata.(jwt.MapClaims)["uid"].(string))
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	if user.HouseId != "null" {
		c.JSON(400, gin.H{"error": models.DBError{
			Message:   "You already have a house!",
			ErrorCode: models.UserInHouse,
		}})
		return
	}

	houseid, err := db.HouseIDByInvite(invite)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	err = db.ChangeHouse(jwtdata.(jwt.MapClaims)["uid"].(string), houseid)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	c.JSON(200, gin.H{})
}

func inviteInfo(c *gin.Context) {
	invite := c.Param("invite")

	houseid, err := db.HouseIDByInvite(invite)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	house, err := db.GetHouse(houseid, "")
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	c.JSON(200, house)
}

func leaveHouse(c *gin.Context) {
	jwtdata, _ := c.Get("data")

	house, err := db.GetUserHouse(jwtdata.(jwt.MapClaims)["uid"].(string))
	if err != nil {
		if dberr, ok := err.(models.DBError); ok {
			if dberr.ErrorCode == models.UserNotFound {
				c.JSON(404, gin.H{"error": &models.DBError{
					Message:   "Your user was not found in the database!",
					ErrorCode: models.NotAuthenticated,
				}})
				return
			}
		}
		c.JSON(400, gin.H{"error": err})
		return
	}

	if house.Owner == jwtdata.(jwt.MapClaims)["uid"].(string) {
		c.JSON(400, gin.H{"error": models.DBError{
			Message:   "You are the owner of this house, you can't leave it!",
			ErrorCode: models.HouseCantLeaveOwner,
		}})
		return
	}

	if len(house.Members) == 1 {
		c.JSON(400, gin.H{"error": models.DBError{
			Message:   "You can't leave a house with only you inside! delete it first",
			ErrorCode: models.HouseCantLeaveMembers,
		}})
		return
	}

	err = db.LeaveHouse(jwtdata.(jwt.MapClaims)["uid"].(string))
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	c.JSON(200, gin.H{})
}
