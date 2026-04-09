package api

import (
	"erp/platform/gateway/middleware"
	"erp/platform/gateway/services"

	"github.com/gin-gonic/gin"
)

func SetupRouter(cl *services.Clients) *gin.Engine {
	r := gin.Default()
	h := &Handlers{Cl: cl}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/v1/auth/login", h.Login)
	r.POST("/v1/auth/register", h.Register)

	authed := r.Group("/v1")
	authed.Use(middleware.RequireAuth(cl))
	{
		authed.GET("/auth/me", h.Me)

		authed.POST("/hr/departments", h.HRCreateDept)
		authed.GET("/hr/departments", h.HRListDept)
		authed.POST("/hr/employees", h.HRCreateEmployee)
		authed.GET("/hr/employees", h.HRListEmployees)
		authed.GET("/hr/employees/:id", h.HRGetEmployee)

		authed.POST("/procurement/suppliers", h.ProcCreateSupplier)
		authed.GET("/procurement/suppliers", h.ProcListSuppliers)
		authed.POST("/procurement/purchase-orders", h.ProcCreatePO)
		authed.GET("/procurement/purchase-orders", h.ProcListPO)
		authed.GET("/procurement/purchase-orders/:id", h.ProcGetPO)

		authed.POST("/warehouse/products", h.WHCreateProduct)
		authed.GET("/warehouse/products", h.WHListProducts)
		authed.GET("/warehouse/products/:id", h.WHGetProduct)
		authed.POST("/warehouse/products/:id/stock", h.WHAdjustStock)

		authed.POST("/finance/invoices", h.FinCreateInvoice)
		authed.GET("/finance/invoices", h.FinListInvoices)
		authed.GET("/finance/invoices/:id", h.FinGetInvoice)
	}

	return r
}
