package routes

import (
	"go.uber.org/ratelimit"

	"github.com/gin-gonic/gin"
	"github.com/zibbadies/homies/internal/homies/middlewares"
)

var ( // requests per second
	authRLimit    = 2
	userRLimit    = 10
	houseRLimit   = 10
	listsRLimit   = 10
	debugRLimit   = 3
	datasetRLimit = 1
)

func SetupRoutes(router *gin.Engine) {
	authRoutes := router.Group("/auth")
	authRoutes.Use(middlewares.GetLimiter(ratelimit.New(authRLimit)))
	authRoutes.POST("/login", authLogin)
	authRoutes.POST("/register", authRegister)
	authRoutes.POST("/renew", middlewares.AuthMiddleware, authRenew)

	userRoutes := router.Group("/user")
	userRoutes.Use(middlewares.GetLimiter(ratelimit.New(userRLimit)))
	userRoutes.Use(middlewares.AuthMiddleware)
	userRoutes.GET("/:id", userInfo)
	userRoutes.GET("/:id/house", userHouse)
	userRoutes.DELETE("/:id/house", leaveHouse)
	userRoutes.GET("/:id/overview", homeOverview)
	userRoutes.GET("/:id/avatar", generateAvatar)
	userRoutes.PATCH("/:id/avatar", setAvatar)

	houseRoutes := router.Group("/house")
	houseRoutes.Use(middlewares.GetLimiter(ratelimit.New(houseRLimit)))
	houseRoutes.Use(middlewares.AuthMiddleware)
	houseRoutes.POST("/create", createHouse)
	houseRoutes.POST("/:invite", joinHouse)
	houseRoutes.GET("/:invite", inviteInfo)

	cartRoutes := router.Group("/lists")
	cartRoutes.Use(middlewares.GetLimiter(ratelimit.New(listsRLimit)))
	cartRoutes.Use(middlewares.AuthMiddleware)
	cartRoutes.GET("/", getLists)
	cartRoutes.GET("/:id/", getItems)
	cartRoutes.PUT("/:id/", newItem)
	cartRoutes.PATCH("/:id/:item_id", updateItem)
	cartRoutes.DELETE("/:id/:item_id", deleteItem)

	debugRoutes := router.Group("/debug")
	debugRoutes.Use(middlewares.GetLimiter(ratelimit.New(debugRLimit)))
	debugRoutes.Use(middlewares.AuthMiddleware, middlewares.AdminMiddleware)
	debugRoutes.GET("/db/check", checkDatabase)

	datasetRoutes := router.Group("/data")
	datasetRoutes.Use(middlewares.GetLimiter(ratelimit.New(datasetRLimit)))
	datasetRoutes.GET("/checks-dataset", checksDataset)
}
