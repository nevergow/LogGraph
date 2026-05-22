package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"loggraph/internal/handler"
)

func New(
	bh *handler.BlockHandler,
	nh *handler.NodeHandler,
	wh *handler.WebhookHandler,
	lh *handler.LarkHandler,
	ah *handler.AttachmentHandler,
) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handler.Health())

		r.Route("/blocks", func(r chi.Router) {
			r.Get("/", bh.List)
			r.Post("/", bh.Create)
			r.Get("/{id}", bh.GetByID)
			r.Patch("/{id}", bh.Update)
			r.Delete("/{id}", bh.Delete)
			r.Get("/{id}/graph", bh.Graph)
		})

		r.Route("/nodes", func(r chi.Router) {
			r.Get("/", nh.List)
			r.Get("/suggest", nh.Suggest)
			r.Get("/{id}/graph", bh.NodeGraph)
		})

		r.Post("/attachments/presign", ah.Presign)

		// Webhook
		r.Route("/webhook", func(r chi.Router) {
			r.Post("/logs", wh.Receive)
			r.Post("/lark", lh.ServeHTTP)
			r.Get("/tokens", wh.ListTokens)
			r.Post("/tokens", wh.GenerateToken)
			r.Delete("/tokens/{id}", wh.DeleteToken)
		})
	})

	return r
}
