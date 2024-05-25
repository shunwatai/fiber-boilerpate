package server

import (
	"fmt"
	"golang-api-starter/internal/auth"
	"golang-api-starter/internal/config"
	zlog "golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/middleware"
	"golang-api-starter/internal/modules/document"
	"golang-api-starter/internal/modules/group"
	"golang-api-starter/internal/modules/groupResourceAcl"
	"golang-api-starter/internal/modules/groupUser"
	"golang-api-starter/internal/modules/log"
	"golang-api-starter/internal/modules/oauth"
	"golang-api-starter/internal/modules/passwordReset"
	"golang-api-starter/internal/modules/permissionType"
	"golang-api-starter/internal/modules/qrcode"
	"golang-api-starter/internal/modules/resource"
	"golang-api-starter/internal/modules/sample"
	"golang-api-starter/internal/modules/todo"
	"golang-api-starter/internal/modules/todoDocument"
	"golang-api-starter/internal/modules/user"
	"golang-api-starter/internal/modules/web"
	"golang-api-starter/web/static"
	lg "log"
	"net/http"
	"strings"

	_ "golang-api-starter/docs" // docs is generated by Swag CLI, you have to import it.

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger" // swagger handler
)

type Fiber struct {
	App *fiber.App
}

var cfg = config.Cfg

func (f *Fiber) GetApp() {
	cfg.LoadEnvVariables()
	zlog.NewZlog()
	f.App = fiber.New(fiber.Config{
		// Prefork:       true,
		CaseSensitive:                true,
		StrictRouting:                false,
		ServerHeader:                 "Fiber",
		BodyLimit:                    500 << 20, // 500Mb
		DisablePreParseMultipartForm: true,      // ref:https://github.com/gofiber/fiber/issues/1838#issuecomment-1086214017
		StreamRequestBody:            true,
	})
}

func (f *Fiber) LoadMiddlewares() {
	f.App.Use(logger.New())
	f.App.Use(recover.New())
	f.App.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowCredentials: true,
	}))
	auth.NewOAuth()
}

func (f *Fiber) LoadSwagger() {
	/* for swagger web */
	serverUrl := fmt.Sprintf("%s/swagger/doc.json", cfg.GetServerUrl())
	f.App.Get("/swagger/*", swagger.HandlerDefault)
	f.App.Get("/swagger/*", swagger.New(swagger.Config{ // custom
		URL:         fmt.Sprintf("http://%s/doc.json", serverUrl),
		DeepLinking: false,
		// Expand ("list") or Collapse ("none") tag groups by default
		DocExpansion: "none",
		// Prefill OAuth ClientId on Authorize popup
		// OAuth: &swagger.OAuthConfig{
		// 	AppName:  "OAuth Provider",
		// 	ClientId: "21bb4edc-05a7-4afc-86f1-2e151e4ba6e2",
		// },
		// Ability to change OAuth2 redirect uri location
		// OAuth2RedirectUrl: fmt.Sprintf("http://%s:8080/swagger/oauth2-redirect.html", serverUrl),
	}))
}

func (f *Fiber) LoadAllRoutes() {
	// embed static files for frontend template
	// ref: https://docs.gofiber.io/api/middleware/filesystem/#embed
	f.App.Use("/static", filesystem.New(filesystem.Config{
		Root:   http.FS(static.StaticDir),
		Browse: false,
	}))

	// initiate custom middleware
	custMiddlewares := middleware.NewCustMiddlewares()

	router := f.App.Group("", custMiddlewares.Log()) // add logging to all routes
	sample.GetRoutes(router, custMiddlewares)        // sample routes for testing
	user.GetRoutes(router, custMiddlewares, group.Repo)
	group.GetRoutes(router, custMiddlewares, user.Repo)
	groupUser.GetRoutes(router, custMiddlewares, group.Repo, user.Repo)
	document.GetRoutes(router, custMiddlewares)
	groupResourceAcl.GetRoutes(router, custMiddlewares)
	log.GetRoutes(router, custMiddlewares)
	oauth.GetRoutes(router, custMiddlewares)
	passwordReset.GetRoutes(router, custMiddlewares)
	permissionType.GetRoutes(router, custMiddlewares)
	qrcode.GetRoutes(router, custMiddlewares)
	resource.GetRoutes(router, custMiddlewares)
	todo.GetRoutes(router, custMiddlewares)
	todoDocument.GetRoutes(router, custMiddlewares)
	web.GetRoutes(router, custMiddlewares)

	// a custom 404 handler instead of default "Cannot GET /page-not-found"
	// ref: https://github.com/gofiber/fiber/issues/748#issuecomment-687503079
	f.App.Use(func(ctx *fiber.Ctx) error {
		return ctx.Status(404).JSON(fiber.Map{
			"code":    404,
			"message": "Resource Not Found",
		})
	})
}

func (f *Fiber) Start() {
	cfg.WatchConfig()

	fmt.Println(strings.Repeat("*", 50))
	fmt.Printf("server env: %+v\n", cfg.ServerConf.Env)
	fmt.Printf("using DB: %+v\n", cfg.DbConf.Driver)
	fmt.Println(strings.Repeat("*", 50))

	lg.Fatal(f.App.Listen(fmt.Sprintf(":%s", cfg.ServerConf.Port)))
}

var Api = &Fiber{}
