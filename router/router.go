package router

import (
	"context"

	"git.blauwelle.com/go/crate/log"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
	"gorm.io/gorm"

	"git.blauwelle.com/go/crate/cmd/sso/handler"
	"git.blauwelle.com/go/crate/cmd/sso/middleware"
	"git.blauwelle.com/go/crate/cmd/sso/util"
)

func NewRouter(db *gorm.DB, redisDB *redis.Client, jwt *util.JWT) *bunrouter.Router {
	log.Info(context.TODO(), "Loading routes...")
	router := bunrouter.New(bunrouter.Use(
		reqlog.NewMiddleware(),
	))

	handlers := handler.NewHandler(db, redisDB, jwt)
	registerRoutes(router, handlers, jwt, db)

	return router
}

func registerRoutes(router *bunrouter.Router, handlers *handler.Handler, jwt *util.JWT, db *gorm.DB) {
	router.POST("/api/v1/login", handlers.Login())
	router.POST("/api/v1/verify", handlers.SSOVerify())

	routerJWTGroup := router.Use(middleware.HTTPMiddlewareJWT(jwt))
	routerJWTGroup.WithGroup("/api/v1", func(g *bunrouter.Group) {
		g.POST("/auth", handlers.SSOLogin())
		g.PUT("/me/username", handlers.UpdateUsername())
		g.PUT("/me/password", handlers.UpdatePassword())
	})

	routerPermissionGroup := routerJWTGroup.Use(middleware.CheckPermission(db))
	routerPermissionGroup.WithGroup("/api/v1", func(g *bunrouter.Group) {
		g.POST("/user/", handlers.CreateUser())
		g.GET("/user/", handlers.SearchUser())
		g.DELETE("/user/", handlers.DeleteUser())
		g.POST("/user/admin", handlers.CreateAdmin())
		g.DELETE("/user/admin", handlers.ConcelAdmin())
		g.POST("/app/", handlers.CreateApp())
		g.GET("/app/", handlers.SearchApp())
		g.DELETE("/app/", handlers.DeleteApp())
		g.PUT("/app/", handlers.UpdateApp())
	})
}
