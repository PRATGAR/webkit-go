package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/example/sample-app/config"
	"github.com/example/sample-app/internal/handlers"
	"github.com/example/sample-app/internal/middleware"
	"github.com/example/sample-app/internal/services"
)

func main() {
	cfg := config.Get()

	if err := services.InitDatabase(cfg.DatabaseURL); err != nil {
		log.Printf("Database connection failed: %v", err)
	}

	services.InitRedis(cfg.RedisURL)

	r := gin.Default()

	setupMiddleware(r)
	setupRoutes(r)

	log.Printf("Server starting on :8080")
	r.Run(":8080")
}

func setupMiddleware(r *gin.Engine) {
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS())
	r.Use(middleware.SecurityHeaders())
}

func setupRoutes(r *gin.Engine) {
	authHandler := handlers.NewAuthHandler()
	userHandler := handlers.NewUserHandler()
	fileHandler := handlers.NewFileHandler()
	adminHandler := handlers.NewAdminHandler()
	searchHandler := handlers.NewSearchHandler()
	reportHandler := handlers.NewReportHandler()
	uploadHandler := handlers.NewUploadHandler()
	webhookHandler := handlers.NewWebhookHandler()
	integrationHandler := handlers.NewIntegrationHandler()
	exportHandler := handlers.NewExportHandler()

	authMiddleware := middleware.NewAuthMiddleware()

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.GET("/redirect", authHandler.Redirect)
			auth.GET("/verify", authHandler.VerifyToken)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.GET("/reset-token", authHandler.GenerateResetToken)
			auth.POST("/session", authHandler.CreateSession)

			auth.POST("/v2/login", authHandler.LoginV2)
			auth.GET("/v2/redirect", authHandler.RedirectV2)
			auth.GET("/v2/verify", authHandler.VerifyTokenV2)
			auth.POST("/v2/session", authHandler.CreateSessionV2)
		}

		users := api.Group("/users")
		{
			users.GET("/:id", userHandler.GetUser)
			users.GET("/profile", userHandler.RenderProfile)
			users.POST("/transfer", userHandler.TransferFunds)
			users.GET("/ldap", userHandler.LDAPSearch)
			users.PUT("/:id", userHandler.UpdateProfile)
			users.GET("/list", userHandler.ListUsers)

			users.GET("/v2/:id", userHandler.GetUserV2)
			users.GET("/v2/profile", userHandler.RenderProfileV2)
			users.POST("/v2/transfer", userHandler.TransferFundsV2)
			users.GET("/v2/ldap", userHandler.LDAPSearchV2)
			users.PUT("/v2/:id", userHandler.UpdateProfileV2)
		}

		files := api.Group("/files")
		{
			files.GET("/download", fileHandler.DownloadFile)
			files.GET("/read", fileHandler.ReadDocument)
			files.POST("/archive", fileHandler.ArchiveFiles)
			files.GET("/document/:id", fileHandler.GetDocument)
			files.GET("/dynamic", fileHandler.DynamicRead)
			files.POST("/resolve", fileHandler.ResolvePath)
			files.GET("/list", fileHandler.ListFiles)
			files.GET("/attachment/:id", fileHandler.AttachReport)
			files.GET("/inline/:id", fileHandler.InlineReport)
			files.GET("/download/safe-archive", fileHandler.DownloadSafeArchive)
			files.GET("/confined-read", fileHandler.ConfinedRead)

			files.GET("/v2/download", fileHandler.DownloadFileV2)
			files.GET("/v2/read", fileHandler.ReadDocumentV2)
			files.POST("/v2/archive", fileHandler.ArchiveFilesV2)
			files.GET("/v2/document/:id", fileHandler.GetDocumentV2)
			files.GET("/v2/list", fileHandler.ListFilesV2)
		}

		admin := api.Group("/admin")
		admin.Use(authMiddleware.Authenticate())
		admin.Use(authMiddleware.AdminOnly())
		{
			admin.GET("/diagnostic", adminHandler.RunDiagnostic)
			admin.POST("/task", adminHandler.ExecuteTask)
			admin.POST("/batch", adminHandler.BatchProcess)
			admin.POST("/async", adminHandler.AsyncExecute)
			admin.POST("/reflect", adminHandler.ReflectionExec)
			admin.GET("/log", adminHandler.LogAction)
			admin.POST("/elevate", adminHandler.ElevateAccess)
			admin.POST("/context", adminHandler.ProcessWithContext)
			admin.POST("/dynamic", adminHandler.DynamicHandler)
			admin.POST("/coerce", adminHandler.TypeCoerce)
			admin.POST("/maintenance/exec", adminHandler.ExecuteMaintenanceTask)

			admin.GET("/v2/diagnostic", adminHandler.RunDiagnosticV2)
			admin.POST("/v2/task", adminHandler.ExecuteTaskV2)
			admin.GET("/v2/log", adminHandler.LogActionV2)
			admin.POST("/v2/elevate", adminHandler.ElevateAccessV2)
			admin.POST("/v2/coerce", adminHandler.TypeCoerceV2)
		}

		search := api.Group("/search")
		{
			search.GET("/users", searchHandler.SearchUsers)
			search.POST("/advanced", searchHandler.AdvancedSearch)
			search.GET("/stream", searchHandler.StreamQuery)
			search.POST("/dynamic", searchHandler.DynamicQuery)
			search.POST("/hof", searchHandler.HigherOrderQuery)
			search.GET("/pattern", searchHandler.ValidatePattern)
			search.POST("/filter", searchHandler.FilteredSearch)
			search.GET("/text", searchHandler.TextSearch)

			search.GET("/v2/users", searchHandler.SearchUsersV2)
			search.POST("/v2/advanced", searchHandler.AdvancedSearchV2)
			search.GET("/v2/pattern", searchHandler.ValidatePatternV2)
			search.POST("/v2/filter", searchHandler.FilteredSearchV2)
			search.GET("/v2/text", searchHandler.TextSearchV2)
		}

		reports := api.Group("/reports")
		{
			reports.GET("/message", reportHandler.DisplayMessage)
			reports.POST("/render", reportHandler.RenderReport)
			reports.POST("/template", reportHandler.RenderTemplate)
			reports.GET("/build", reportHandler.BuildResponse)
			reports.POST("/dynamic", reportHandler.DynamicRender)
			reports.POST("/pdf", reportHandler.GeneratePDF)
			reports.GET("/catalog", reportHandler.ListCatalog)
			reports.GET("/query/sorted", reportHandler.QuerySortedReports)

			reports.GET("/v2/message", reportHandler.DisplayMessageV2)
			reports.POST("/v2/render", reportHandler.RenderReportV2)
			reports.POST("/v2/template", reportHandler.RenderTemplateV2)
			reports.GET("/v2/build", reportHandler.BuildResponseV2)
		}

		upload := api.Group("/upload")
		{
			upload.POST("/config", uploadHandler.ImportConfig)
			upload.POST("/xml", uploadHandler.ParseXML)
			upload.POST("/generic", uploadHandler.HandleGeneric)
			upload.POST("/archive", uploadHandler.ExtractArchive)
			upload.POST("/multi", uploadHandler.ProcessMultipleFormats)
			upload.POST("/file", uploadHandler.UploadFile)

			upload.POST("/v2/config", uploadHandler.ImportConfigV2)
			upload.POST("/v2/xml", uploadHandler.ParseXMLV2)
			upload.POST("/v2/archive", uploadHandler.ExtractArchiveV2)
			upload.POST("/v2/file", uploadHandler.UploadFileV2)
		}

		webhook := api.Group("/webhook")
		{
			webhook.GET("/header", webhookHandler.SetHeader)
			webhook.POST("/forward", webhookHandler.Forward)
			webhook.POST("/callback", webhookHandler.WebhookForward)
			webhook.POST("/event", webhookHandler.ProcessEvent)
			webhook.GET("/response", webhookHandler.DynamicResponse)
			webhook.POST("/register", webhookHandler.RegisterWebhook)

			webhook.GET("/v2/header", webhookHandler.SetHeaderV2)
			webhook.POST("/v2/forward", webhookHandler.ForwardV2)
			webhook.POST("/v2/event", webhookHandler.ProcessEventV2)
			webhook.POST("/v2/register", webhookHandler.RegisterWebhookV2)
		}

		integrations := api.Group("/integrations")
		{
			integrations.GET("/fetch", integrationHandler.FetchExternal)
			integrations.GET("/proxy", integrationHandler.ProxyRequest)
			integrations.POST("/route", integrationHandler.RouteRequest)
			integrations.GET("/build", integrationHandler.URLBuilder)
			integrations.POST("/configure", integrationHandler.ConfigureIntegration)
			integrations.GET("/fasthttp", integrationHandler.FastHTTPProxy)
			integrations.POST("/sync", integrationHandler.SyncData)
			integrations.POST("/custom-webhook", integrationHandler.CustomWebhookDispatch)

			integrations.GET("/v2/fetch", integrationHandler.FetchExternalV2)
			integrations.GET("/v2/proxy", integrationHandler.ProxyRequestV2)
			integrations.POST("/v2/route", integrationHandler.RouteRequestV2)
			integrations.POST("/v2/sync", integrationHandler.SyncDataV2)
		}

		export := api.Group("/export")
		{
			export.GET("/data", exportHandler.ExportData)
			export.GET("/goroutine", exportHandler.GoroutineFlow)
			export.GET("/aliased", exportHandler.AliasedFlow)
			export.POST("/batch", exportHandler.BatchExport)
			export.GET("/stream", exportHandler.StreamExport)

			export.GET("/v2/data", exportHandler.ExportDataV2)
			export.GET("/v2/goroutine", exportHandler.GoroutineFlowV2)
			export.GET("/v2/stream", exportHandler.StreamExportV2)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/config", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"debug":   config.Get().Debug,
			"version": "1.0.0",
		})
	})
}
