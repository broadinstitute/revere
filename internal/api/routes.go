package api

import (
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/broadinstitute/revere/internal/version"
	"github.com/gin-gonic/gin"
	"net/http"
)

func getVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": version.BuildVersion})
}

func getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func NewRouter(config *configuration.Config) *gin.Engine {
	if config.Api.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	api := router.Group("/api/v1")

	// Routes available on both / and /api/v1/
	for _, g := range []*gin.RouterGroup{&router.RouterGroup, api} {
		g.GET("/version", getVersion)
		g.GET("/status", getStatus)
	}

	return router
}
