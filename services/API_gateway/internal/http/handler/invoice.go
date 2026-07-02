package handler

import (
	"api_gateway/internal/proto/adminpb"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

type InvoiceHandle struct {
	adminClient adminpb.AppUserServiceClient
}

func NewInvoiceHandler(client adminpb.AppUserServiceClient) *InvoiceHandle {
	return &InvoiceHandle{
		adminClient: client,
	}
}

func (h *InvoiceHandle) DownloadInvoice(c *gin.Context) {

	parts := strings.Split(c.Request.URL.Path, "/")
	if len(parts) < 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice path"})
		return
	}

	invoiceID := parts[4]

	// Get Authorization header
	auth := c.GetHeader("Authorization")

	// Create outgoing gRPC metadata
	md := metadata.New(map[string]string{
		"authorization": auth,
	})

	ctx := metadata.NewOutgoingContext(c.Request.Context(), md)

	resp, err := h.adminClient.GetInvoice(ctx, &adminpb.GetInvoiceRequest{InvoiceId: invoiceID})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := resp.Filename
	if filename == "" {
		filename = fmt.Sprintf("invoice-%s.pdf", invoiceID)
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Length", strconv.Itoa(len(resp.Pdf)))
	c.Data(http.StatusOK, "application/pdf", resp.Pdf)
}
