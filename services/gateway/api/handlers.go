package api

import (
	"net/http"

	authv1 "erp/platform/genproto/erp/auth/v1"
	finv1 "erp/platform/genproto/erp/finance/v1"
	hrv1 "erp/platform/genproto/erp/hr/v1"
	procv1 "erp/platform/genproto/erp/procurement/v1"
	whv1 "erp/platform/genproto/erp/warehouse/v1"
	"erp/platform/gateway/middleware"
	"erp/platform/gateway/services"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Cl *services.Clients
}

func (h *Handlers) Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.Cl.Auth.Login(c.Request.Context(), &authv1.LoginRequest{Email: body.Email, Password: body.Password})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": resp.GetAccessToken(), "user_id": resp.GetUserId(), "role": resp.GetRole()})
}

func (h *Handlers) Register(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.Cl.Auth.Register(c.Request.Context(), &authv1.RegisterRequest{Email: body.Email, Password: body.Password, Role: body.Role})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user_id": resp.GetUserId()})
}

func (h *Handlers) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"user_id": c.GetString(middleware.CtxUserID), "role": c.GetString(middleware.CtxRole)})
}

// --- HR ---

func (h *Handlers) HRCreateDept(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.Cl.HR.CreateDepartment(c.Request.Context(), &hrv1.CreateDepartmentRequest{Name: body.Name})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handlers) HRListDept(c *gin.Context) {
	resp, err := h.Cl.HR.ListDepartments(c.Request.Context(), &hrv1.ListDepartmentsRequest{Limit: 50, Offset: 0})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handlers) HRCreateEmployee(c *gin.Context) {
	var body struct {
		DepartmentID string `json:"department_id" binding:"required"`
		UserID       string `json:"user_id"`
		FullName     string `json:"full_name" binding:"required"`
		JobTitle     string `json:"job_title"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.Cl.HR.CreateEmployee(c.Request.Context(), &hrv1.CreateEmployeeRequest{
		DepartmentId: body.DepartmentID, UserId: body.UserID, FullName: body.FullName, JobTitle: body.JobTitle,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handlers) HRListEmployees(c *gin.Context) {
	resp, err := h.Cl.HR.ListEmployees(c.Request.Context(), &hrv1.ListEmployeesRequest{Limit: 50, Offset: 0})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handlers) HRGetEmployee(c *gin.Context) {
	resp, err := h.Cl.HR.GetEmployee(c.Request.Context(), &hrv1.GetEmployeeRequest{Id: c.Param("id")})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Procurement ---

func (h *Handlers) ProcCreateSupplier(c *gin.Context) {
	var body struct {
		Name          string `json:"name" binding:"required"`
		ContactEmail  string `json:"contact_email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.Cl.Procurement.CreateSupplier(c.Request.Context(), &procv1.CreateSupplierRequest{Name: body.Name, ContactEmail: body.ContactEmail})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handlers) ProcListSuppliers(c *gin.Context) {
	resp, err := h.Cl.Procurement.ListSuppliers(c.Request.Context(), &procv1.ListSuppliersRequest{Limit: 50, Offset: 0})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handlers) ProcCreatePO(c *gin.Context) {
	var body struct {
		SupplierID string `json:"supplier_id" binding:"required"`
		Lines      []struct {
			ProductID string `json:"product_id" binding:"required"`
			Quantity  int32  `json:"quantity" binding:"required"`
			UnitPrice string `json:"unit_price" binding:"required"`
		} `json:"lines" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var lines []*procv1.PurchaseOrderLine
	for _, ln := range body.Lines {
		lines = append(lines, &procv1.PurchaseOrderLine{ProductId: ln.ProductID, Quantity: ln.Quantity, UnitPrice: ln.UnitPrice})
	}
	resp, err := h.Cl.Procurement.CreatePurchaseOrder(c.Request.Context(), &procv1.CreatePurchaseOrderRequest{
		SupplierId: body.SupplierID, Lines: lines,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handlers) ProcListPO(c *gin.Context) {
	resp, err := h.Cl.Procurement.ListPurchaseOrders(c.Request.Context(), &procv1.ListPurchaseOrdersRequest{Limit: 50, Offset: 0})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handlers) ProcGetPO(c *gin.Context) {
	resp, err := h.Cl.Procurement.GetPurchaseOrder(c.Request.Context(), &procv1.GetPurchaseOrderRequest{Id: c.Param("id")})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Warehouse ---

func (h *Handlers) WHCreateProduct(c *gin.Context) {
	var body struct {
		SKU         string `json:"sku" binding:"required"`
		Name        string `json:"name" binding:"required"`
		InitialQty  int64  `json:"initial_qty"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.Cl.Warehouse.CreateProduct(c.Request.Context(), &whv1.CreateProductRequest{Sku: body.SKU, Name: body.Name, InitialQty: body.InitialQty})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handlers) WHListProducts(c *gin.Context) {
	resp, err := h.Cl.Warehouse.ListProducts(c.Request.Context(), &whv1.ListProductsRequest{Limit: 50, Offset: 0})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handlers) WHGetProduct(c *gin.Context) {
	resp, err := h.Cl.Warehouse.GetProduct(c.Request.Context(), &whv1.GetProductRequest{Id: c.Param("id")})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handlers) WHAdjustStock(c *gin.Context) {
	var body struct {
		Delta  int64  `json:"delta" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.Cl.Warehouse.AdjustStock(c.Request.Context(), &whv1.AdjustStockRequest{
		ProductId: c.Param("id"), Delta: body.Delta, Reason: body.Reason,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Finance ---

func (h *Handlers) FinCreateInvoice(c *gin.Context) {
	var body struct {
		PurchaseOrderID string `json:"purchase_order_id"`
		Amount          string `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.Cl.Finance.CreateInvoice(c.Request.Context(), &finv1.CreateInvoiceRequest{
		PurchaseOrderId: body.PurchaseOrderID, Amount: body.Amount,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handlers) FinListInvoices(c *gin.Context) {
	resp, err := h.Cl.Finance.ListInvoices(c.Request.Context(), &finv1.ListInvoicesRequest{Limit: 50, Offset: 0})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handlers) FinGetInvoice(c *gin.Context) {
	resp, err := h.Cl.Finance.GetInvoice(c.Request.Context(), &finv1.GetInvoiceRequest{Id: c.Param("id")})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
