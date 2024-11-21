package serve

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"log"
	"time"
	"video-translation-status/server/pkg/controllers"
)

func NewServeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve video status backend",
		Long:  "Serve video status backend",
		Run:   runServe,
	}

	return cmd
}

func runServe(cmd *cobra.Command, args []string) {
	router := SetupRouter()

	err := router.Run("0.0.0.0:8080")
	if err != nil {
		log.Fatal(err)
	}
}

func SetupRouter() *gin.Engine {
	router := gin.New()
	corsMiddleware := cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST"},
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	})

	router.Use(corsMiddleware)
	router.Use(gin.Recovery())

	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "Not found", "message": "Page not found"})
	})

	translationJobRouter := controllers.NewTranslationJobHandler(5*time.Second, 15*time.Second)

	router.POST("/", translationJobRouter.CreateJob)
	router.GET("/status/:job_id", translationJobRouter.GetJobStatus)

	return router
}
