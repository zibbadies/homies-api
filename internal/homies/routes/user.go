package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/zibbadies/homies/internal/homies/checks"
	"github.com/zibbadies/homies/internal/homies/db"
	"github.com/zibbadies/homies/internal/homies/models"
	"github.com/zibbadies/homies/pkg/homies/avatar"
)

func userInfo(c *gin.Context) {
	id_param := c.Param("id")
	jwtdata, _ := c.Get("data")

	uid := id_param
	if (id_param == "me") {
		uid = jwtdata.(jwt.MapClaims)["uid"].(string)
	} else { // let's avoid it for now...
		c.JSON(403, gin.H{"error": models.DBError{
			Message: "You can't see other people's info!",
			ErrorCode: models.NotAuthorized,
		}})
		return
	}

	dbuser, err := db.GetUserMe(uid)
	if (err != nil) {
		c.JSON(400, gin.H{"error": err})
		return
	}

	c.JSON(200, dbuser.Account);
}

func homeOverview(c *gin.Context) {
	jwtdata, _ := c.Get("data")	

	user, err := db.GetUserMe(jwtdata.(jwt.MapClaims)["uid"].(string))
	if (err != nil) {
		c.JSON(400, gin.H{"error": err})
		return
	}

	house, err := db.GetUserHouse(jwtdata.(jwt.MapClaims)["uid"].(string))
	if (err != nil) {
		c.JSON(400, gin.H{"error": err})
		return
	}

	lists, err := db.GetLists(user.HouseId)
	if (err != nil) {
		c.JSON(400, gin.H{"error": err})
		return
	}

	d_lists := make(map[string][]models.Item)

	for _, list := range lists {
		/* 
			TODO: Replace this if with the DB in_overview field
		*/
		if (list.Name == "todo" || list.Name == "shopping") {
			d_lists[list.Name], err = db.GetItems(list.Id, models.NoTime, models.NoTime, 10)

			if (err != nil) {
				c.JSON(400, gin.H{"error": err})
				return
			}
		}
	}

	overview := models.Overview{
		User: user.User,
		House: house,
		Lists: d_lists,
	}

	c.JSON(200, overview);
}

func generateAvatar(c *gin.Context) {
	/*
		There's no fucking reason to keep this server-side, but to speed up the
		app development, we'll keep it this way. for now.
	*/
	avatar := avatar.RandAvatar()
	raw := fmt.Sprintf(`
<svg width="80" height="80" xmlns="http://www.w3.org/2000/svg">
    <circle r="40" cx="40" cy="40" fill="#%s"/>
    <g transform="translate(%f, %f) scale(2)">
        <path d="M4 10c%s" stroke="#%s" fill="none" stroke-linecap="round"></path>
        <rect x="%f" y="%f" width="1.5" height="2" rx="1" stroke="none" fill="#%s"></rect>
        <rect x="%f" y="%f" width="1.5" height="2" rx="1" stroke="none" fill="#%s"></rect>
    </g>
</svg>
	`, avatar.BgColor, avatar.FaceX, avatar.FaceY, avatar.Bezier, avatar.FaceColor, avatar.LeX, avatar.LeY, avatar.FaceColor, avatar.ReX, avatar.ReY, avatar.FaceColor)

	c.Data(200, "image/svg+xml; charset=utf-8", []byte(raw))
}

func setAvatar(c *gin.Context) {
	/*
		A funny guy could just send handmade values to this endpoint to get some *interesting* avatars,
		if he does that, well, good for him, he earned it.

		( We tought about adding a digital signature to the avatar in generateAvatar() 
		to prevent this, but it's more funny this way )
	*/
	jwtdata, _ := c.Get("data")

	var avatar models.Avatar;
	err := c.ShouldBind(&avatar)
	if (err != nil) {
		c.JSON(400, gin.H{"error": models.DBError{
			Message: "Invalid JSON Data!",
			ErrorCode: models.JsonFormatError,
		}})
		return
	}

	if err = checks.Check("color", avatar.BgColor); err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	if err = checks.Check("bezier", avatar.Bezier); err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	err = db.SetAvatar(jwtdata.(jwt.MapClaims)["uid"].(string), avatar)
	if (err != nil) {
		c.JSON(400, gin.H{"error": err})
		return
	}

	c.JSON(200, gin.H{});
}

