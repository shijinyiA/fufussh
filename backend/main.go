package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"web-ssh/handlers"
	"web-ssh/store"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mysqlDSN := os.Getenv("MYSQL_DSN")
	if mysqlDSN == "" {
		log.Fatal("MYSQL_DSN 环境变量未设置，请通过安装脚本配置数据库连接信息，或在 .env 文件中设置 MYSQL_DSN")
	}

	dbStore, err := store.NewMySQLStore(mysqlDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\nPlease ensure MySQL is running and database 'ssh-web' exists.", err)
	}
	defer dbStore.Close()

	authHandler := handlers.NewAuthHandler(dbStore)
	serverHandler := handlers.NewServerHandler(dbStore)
	sshHandler := handlers.NewSSHHandler(dbStore)
	sftpHandler := handlers.NewSFTPHandler(dbStore)
	settingsHandler := handlers.NewSettingsHandler(dbStore)
	systemSettingsHandler := handlers.NewSystemSettingsHandler(dbStore)
	directHandler := handlers.NewDirectConnectHandler(dbStore)
	directSFTPHandler := handlers.NewDirectSFTPHandler()
	userHandler := handlers.NewUserHandler(dbStore)
	shareHandler := handlers.NewShareHandler(dbStore)
	adminHandler := handlers.NewAdminHandler(dbStore)

	go authHandler.CleanupSessions()
	go handlers.CleanupDirectSessions()
	go handlers.CleanupDirectSFTPSessions()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Cookie"},
		ExposeHeaders:    []string{"Content-Length", "Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	}))

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/check", authHandler.CheckAuth)
		}

		api.GET("/system-settings", systemSettingsHandler.GetPublicSettings)
		api.GET("/quick-commands", systemSettingsHandler.GetPublicCommands)

		admin := api.Group("/admin")
		{
			admin.GET("/status", adminHandler.Status)
			admin.POST("/init", adminHandler.Init)
			admin.POST("/login", adminHandler.Login)

			adminProtected := admin.Group("")
			adminProtected.Use(adminHandler.RequireAdmin())
			{
				adminProtected.GET("/check", adminHandler.Check)
				adminProtected.POST("/change-password", adminHandler.ChangePassword)
				adminProtected.GET("/stats", adminHandler.Stats)
			adminProtected.GET("/system-info", adminHandler.SystemInfo)

				adminProtected.GET("/users", adminHandler.ListUsers)
				adminProtected.POST("/users", adminHandler.CreateUser)
				adminProtected.PUT("/users/:id", adminHandler.UpdateUser)
				adminProtected.PUT("/users/:id/reset-password", adminHandler.ResetUserPassword)
				adminProtected.DELETE("/users/:id", adminHandler.DeleteUser)

				adminProtected.GET("/servers", adminHandler.ListServers)
				adminProtected.DELETE("/servers/:id", adminHandler.DeleteServer)

				adminProtected.GET("/settings", adminHandler.GetSettings)
				adminProtected.POST("/settings", adminHandler.SaveSettings)
				adminProtected.GET("/commands", adminHandler.GetCommands)
				adminProtected.POST("/commands", adminHandler.SaveCommands)
				adminProtected.POST("/upload", adminHandler.Upload)
				adminProtected.GET("/logs", adminHandler.ListLogs)

				adminProtected.GET("/database/tables", adminHandler.ListDatabaseTables)
			adminProtected.GET("/database/status", adminHandler.DatabaseStatus)
			adminProtected.GET("/database/backup", adminHandler.BackupDatabase)
				adminProtected.GET("/database/tables/:table", adminHandler.BrowseDatabaseTable)
				adminProtected.DELETE("/database/tables/:table/rows", adminHandler.DeleteDatabaseRow)
				adminProtected.POST("/database/query", adminHandler.QueryDatabase)
			}
		}

		hardwareMonitorHandler := handlers.NewHardwareMonitorHandler(dbStore)
		api.GET("/hardware-stats", hardwareMonitorHandler.GetRemoteStats)

		direct := api.Group("/direct")
		{
			direct.POST("/connect", directHandler.Connect)
			direct.GET("/session/:token", directHandler.GetSession)
			direct.GET("/ws/terminal", sshHandler.HandleDirectTerminal)
			direct.GET("/ws/sftp", directSFTPHandler.HandleDirectSFTP)
			direct.GET("/sftp/download", directSFTPHandler.HandleDirectSFTPDownload)
		}

		share := api.Group("/share")
		{
			share.GET("/:token", shareHandler.GetSharedServer)
		}

		protected := api.Group("")
		protected.Use(authHandler.RequireAuth())
		{
			user := protected.Group("/user")
			{
				user.GET("/profile", userHandler.GetProfile)
				user.PUT("/profile", userHandler.UpdateProfile)
				user.POST("/password", userHandler.ChangePassword)
			}

			servers := protected.Group("/servers")
			{
				servers.GET("", serverHandler.ListServers)
				servers.GET("/groups", serverHandler.GetGroups)
				servers.GET("/:id", serverHandler.GetServer)
				servers.POST("", serverHandler.CreateServer)
				servers.PUT("/:id", serverHandler.UpdateServer)
				servers.DELETE("/:id", serverHandler.DeleteServer)
				servers.POST("/:id/test", serverHandler.TestConnection)
				servers.POST("/:id/share", shareHandler.CreateShareToken)
			}

			settings := protected.Group("/settings")
			{
				settings.GET("/:id", settingsHandler.GetSettings)
				settings.POST("/:id", settingsHandler.SaveSettings)
			}

			protected.GET("/ws/terminal", sshHandler.HandleTerminal)
			protected.GET("/ws/sftp", sftpHandler.HandleSFTPWebSocket)
			protected.POST("/sftp/upload", sftpHandler.HandleFileUpload)
			protected.GET("/sftp/download", sftpHandler.HandleFileDownload)
			protected.GET("/sessions", sshHandler.GetActiveSessions)
		}
	}

	r.Static("/uploads", "./uploads")

	distDir := "./frontend_dist"
	if _, err := os.Stat(distDir); err == nil {
		r.Static("/assets", filepath.Join(distDir, "assets"))
		r.StaticFile("/favicon.svg", filepath.Join(distDir, "favicon.svg"))
		r.NoRoute(func(c *gin.Context) { c.File(filepath.Join(distDir, "index.html")) })
	} else {
		log.Printf("Warning: Frontend dist directory not found at %s", distDir)
		r.NoRoute(func(c *gin.Context) { c.String(http.StatusNotFound, "Frontend not built") })
	}

	adminDistDir := "./admin_dist"
	if _, err := os.Stat(adminDistDir); err == nil {
		r.Static("/admin/assets", filepath.Join(adminDistDir, "assets"))
		r.StaticFile("/admin/favicon.svg", filepath.Join(adminDistDir, "favicon.svg"))

		adminRoutes := []string{
			"/admin", "/admin/", "/admin/login", "/admin/init", "/admin/database",
			"/admin/users", "/admin/servers", "/admin/settings", "/admin/commands", "/admin/logs",
		}
		for _, route := range adminRoutes {
			r.GET(route, func(c *gin.Context) {
				if c.Request.URL.Path == "/admin" {
					c.Redirect(302, "/admin/")
					return
				}
				c.File(filepath.Join(adminDistDir, "index.html"))
			})
		}

		log.Printf("Admin panel available at /admin")
	}

	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║      芙芙云 SSH Terminal Server 2.0       ║")
	fmt.Println("╠══════════════════════════════════════════╣")
	fmt.Printf("║  Server:  http://0.0.0.0:%s              ║\n", port)
	fmt.Printf("║  Admin:   http://0.0.0.0:%s/admin        ║\n", port)
	fmt.Printf("║  Database: MySQL (ssh-web)               ║\n")
	fmt.Println("╚══════════════════════════════════════════╝")

	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down server...")
}
