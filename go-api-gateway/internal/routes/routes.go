package routes

import (
	"net/http"

	"go-api-gateway/internal/config"
	"go-api-gateway/internal/middleware"
	"go-api-gateway/internal/proxy"

	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App, cfg config.Config) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Go Fiber API Gateway is running",
		})
	})

	api := app.Group("/api")

	// Auth routes
	api.Post("/user/register", proxy.Public(cfg.AuthServiceURL+"/user/register", http.MethodPost))
	api.Post("/auth/login", proxy.Public(cfg.AuthServiceURL+"/auth/login", http.MethodPost))

	// Public product routes
	api.Get("/products", proxy.WithGateway(cfg.ProductServiceURL+"/products", http.MethodGet, cfg.GatewaySecret))
	api.Get("/product/:id", func(c *fiber.Ctx) error {
		return proxy.WithGateway(cfg.ProductServiceURL+"/product/"+c.Params("id"), http.MethodGet, cfg.GatewaySecret)(c)
	})

	protected := api.Group("/", middleware.Protected(cfg.AuthServiceURL))

	// Protected auth routes
	protected.Get("/user/profile", proxy.Public(cfg.AuthServiceURL+"/user/profile", http.MethodGet))
	protected.Put("/user/update", proxy.Public(cfg.AuthServiceURL+"/user/update", http.MethodPut))
	protected.Put("/auth/token-refresh", proxy.Public(cfg.AuthServiceURL+"/auth/token-refresh", http.MethodPut))
	protected.Delete("/auth/token-revoke", proxy.Public(cfg.AuthServiceURL+"/auth/token-revoke", http.MethodDelete))

	// Protected product routes
	product := protected.Group("/product")
	product.Post("/create", proxy.WithGateway(cfg.ProductServiceURL+"/product/create", http.MethodPost, cfg.GatewaySecret))
	product.Put("/update/:id", func(c *fiber.Ctx) error {
		return proxy.WithGateway(cfg.ProductServiceURL+"/product/update/"+c.Params("id"), http.MethodPut, cfg.GatewaySecret)(c)
	})
	product.Delete("/delete/:id", func(c *fiber.Ctx) error {
		return proxy.WithGateway(cfg.ProductServiceURL+"/product/delete/"+c.Params("id"), http.MethodDelete, cfg.GatewaySecret)(c)
	})

	// Protected order routes
	protected.Get("/orders", proxy.WithGateway(cfg.OrderServiceURL+"/orders", http.MethodGet, cfg.GatewaySecret))
	order := protected.Group("/order")
	order.Post("/create", proxy.WithGateway(cfg.OrderServiceURL+"/order/create", http.MethodPost, cfg.GatewaySecret))
	order.Get("/:id", func(c *fiber.Ctx) error {
		return proxy.WithGateway(cfg.OrderServiceURL+"/order/"+c.Params("id"), http.MethodGet, cfg.GatewaySecret)(c)
	})
	order.Put("/update/:id", func(c *fiber.Ctx) error {
		return proxy.WithGateway(cfg.OrderServiceURL+"/order/update/"+c.Params("id"), http.MethodPut, cfg.GatewaySecret)(c)
	})
	order.Delete("/delete/:id", func(c *fiber.Ctx) error {
		return proxy.WithGateway(cfg.OrderServiceURL+"/order/delete/"+c.Params("id"), http.MethodDelete, cfg.GatewaySecret)(c)
	})
}
