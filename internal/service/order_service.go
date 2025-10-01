package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/chai2010/webp"
	"github.com/jung-kurt/gofpdf"

	_ "image/gif"
	_ "image/jpeg"

	"smartseller-lite-starter/internal/domain"
	"smartseller-lite-starter/internal/repo"
)

type OrderItemInput struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
	Discount  float64 `json:"discount"`
}

type CreateOrderInput struct {
	BuyerID      string           `json:"buyerId"`
	RecipientID  string           `json:"recipientId"`
	Items        []OrderItemInput `json:"items"`
	Discount     float64          `json:"discount"`
	Notes        string           `json:"notes"`
	Courier      string           `json:"courier"`
	ServiceLevel string           `json:"serviceLevel"`
	TrackingCode string           `json:"trackingCode"`
	ShippingCost float64          `json:"shippingCost"`
}

// OrderService coordinates order lifecycle and profit calculation.
type OrderService struct {
	repo      *repo.OrderRepository
	products  *ProductService
	customers *CustomerService
	settings  *SettingsService
}

type OrderListOptions struct {
	Query     string     `json:"query"`
	Courier   string     `json:"courier"`
	DateStart *time.Time `json:"dateStart,omitempty"`
	DateEnd   *time.Time `json:"dateEnd,omitempty"`
	Page      int        `json:"page"`
	PageSize  int        `json:"pageSize"`
}

type OrderListSummary struct {
	Count          int     `json:"count"`
	Revenue        float64 `json:"revenue"`
	Profit         float64 `json:"profit"`
	TopCourier     string  `json:"topCourier"`
	TopCourierHits int     `json:"topCourierHits"`
	TopProductID   string  `json:"topProductId"`
	TopProductName string  `json:"topProductName"`
	TopProductQty  int     `json:"topProductQty"`
}

type OrderListResult struct {
	Items    []domain.Order  `json:"items"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
	Summary  OrderListSummary `json:"summary"`
	Couriers []string        `json:"couriers"`
}

func NewOrderService(repo *repo.OrderRepository, products *ProductService, customers *CustomerService, settings *SettingsService) *OrderService {
	return &OrderService{repo: repo, products: products, customers: customers, settings: settings}
}

func (s *OrderService) Warm(ctx context.Context) {
	_, _ = s.repo.List(ctx, 5)
}

func (s *OrderService) Create(ctx context.Context, input CreateOrderInput) (*domain.Order, error) {
	if input.BuyerID == "" || input.RecipientID == "" {
		return nil, errors.New("buyer and recipient are required")
	}
	if len(input.Items) == 0 {
		return nil, errors.New("order requires at least one item")
	}
	input.Courier = strings.TrimSpace(input.Courier)
	input.ServiceLevel = strings.TrimSpace(input.ServiceLevel)
	input.TrackingCode = strings.TrimSpace(input.TrackingCode)
	input.Notes = strings.TrimSpace(input.Notes)
	if input.Courier == "" {
		input.Courier = "JNE"
	}
	if input.Discount < 0 {
		input.Discount = 0
	}
	if input.ShippingCost < 0 {
		input.ShippingCost = 0
	}

	var (
		subtotal  float64
		totalCost float64
		items     []domain.OrderItem
	)

	for _, line := range input.Items {
		if line.ProductID == "" {
			return nil, errors.New("item missing product")
		}
		if line.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity for product %s", line.ProductID)
		}
		prod, err := s.products.Get(ctx, line.ProductID)
		if err != nil {
			return nil, err
		}
		if prod.Stock < line.Quantity {
			return nil, fmt.Errorf("insufficient stock for %s", prod.Name)
		}
		unitPrice := line.UnitPrice
		if unitPrice <= 0 {
			unitPrice = prod.SalePrice
		}
		if unitPrice < 0 {
			unitPrice = 0
		}
		itemDiscount := line.Discount
		if itemDiscount < 0 {
			itemDiscount = 0
		}
		lineRevenue := unitPrice*float64(line.Quantity) - itemDiscount
		if lineRevenue < 0 {
			lineRevenue = 0
		}
		cost := prod.CostPrice * float64(line.Quantity)
		lineProfit := lineRevenue - cost

		subtotal += lineRevenue
		totalCost += cost

		items = append(items, domain.OrderItem{
			ProductID: prod.ID,
			SKU:       prod.SKU,
			Quantity:  line.Quantity,
			UnitPrice: unitPrice,
			Discount:  itemDiscount,
			CostPrice: prod.CostPrice,
			Profit:    lineProfit,
		})
	}

	profit := subtotal - input.Discount - totalCost - input.ShippingCost
	if profit < 0 {
		profit = 0
	}

	order := &domain.Order{
		BuyerID:     input.BuyerID,
		RecipientID: input.RecipientID,
		Items:       items,
		Discount:    input.Discount,
		Notes:       input.Notes,
		Shipment: domain.Shipment{
			Courier:      input.Courier,
			ServiceLevel: input.ServiceLevel,
			TrackingCode: input.TrackingCode,
			ShippingCost: input.ShippingCost,
		},
		Total:  subtotal - input.Discount + input.ShippingCost,
		Profit: profit,
	}

	saved, err := s.repo.Create(ctx, order)
	if err != nil {
		return nil, err
	}

	// Reduce stock after persisting order to minimise lost updates.
	for _, item := range saved.Items {
		reason := fmt.Sprintf("order:%s", saved.Code)
		if err := s.products.AdjustStock(ctx, item.ProductID, -item.Quantity, reason); err != nil {
			return nil, err
		}
	}

	return saved, nil
}

func (s *OrderService) List(ctx context.Context, limit int) ([]domain.Order, error) {
	return s.repo.List(ctx, limit)
}

func (s *OrderService) ListPaged(ctx context.Context, opts OrderListOptions) (OrderListResult, error) {
	repoResult, err := s.repo.ListPaged(ctx, repo.OrderListOptions{
		Query:     opts.Query,
		Courier:   opts.Courier,
		DateStart: opts.DateStart,
		DateEnd:   opts.DateEnd,
		Page:      opts.Page,
		PageSize:  opts.PageSize,
	})
	if err != nil {
		return OrderListResult{}, err
	}

	result := OrderListResult{
		Items:    repoResult.Items,
		Total:    repoResult.Total,
		Page:     repoResult.Page,
		PageSize: repoResult.PageSize,
		Couriers: repoResult.Couriers,
		Summary: OrderListSummary{
			Count:          repoResult.Summary.Count,
			Revenue:        repoResult.Summary.Revenue,
			Profit:         repoResult.Summary.Profit,
			TopCourier:     repoResult.Summary.TopCourier,
			TopCourierHits: repoResult.Summary.TopCourierHits,
			TopProductID:   repoResult.Summary.TopProductID,
			TopProductName: repoResult.Summary.TopProductName,
			TopProductQty:  repoResult.Summary.TopProductQty,
		},
	}
	return result, nil
}

func (s *OrderService) ListAll(ctx context.Context) ([]domain.Order, error) {
	return s.repo.ListAll(ctx)
}

func (s *OrderService) Get(ctx context.Context, id string) (*domain.Order, error) {
	if id == "" {
		return nil, errors.New("order id required")
	}
	return s.repo.Get(ctx, id)
}

func (s *OrderService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("order id required")
	}
	return s.repo.Delete(ctx, id)
}

func (s *OrderService) ReplaceAll(ctx context.Context, orders []domain.Order) error {
	return s.repo.ReplaceAll(ctx, orders)
}

func (s *OrderService) GenerateLabelPDF(ctx context.Context, orderID string) ([]byte, error) {
	order, err := s.Get(ctx, orderID)
	if err != nil {
		return nil, err
	}

	var recipient *domain.Customer
	if strings.TrimSpace(order.RecipientID) == "" {
		recipient = &domain.Customer{
			Name:    "Penerima belum ditentukan",
			Address: "Lengkapi data penerima pada order sebelum pengiriman.",
		}
	} else {
		if fetched, err := s.customers.Get(ctx, order.RecipientID); err == nil {
			recipient = fetched
		} else {
			recipient = &domain.Customer{
				Name:    "Kontak penerima tidak ditemukan",
				Address: fmt.Sprintf("Data penerima dengan ID %s tidak tersedia.", order.RecipientID),
			}
		}
	}

	var settings domain.AppSettings
	var logoBytes []byte
	var logoMime string
	if s.settings != nil {
		if fetched, err := s.settings.Get(ctx); err == nil {
			settings = fetched
			if data, mime, err := s.settings.LoadLogoBytes(ctx, fetched); err == nil {
				logoBytes = data
				logoMime = mime
			}
		}
	}
	brandName := strings.TrimSpace(settings.BrandName)
	if brandName == "" {
		brandName = "SmartSeller Lite"
	}

	pdf := gofpdf.New("P", "mm", "A6", "")
	pdf.SetMargins(12, 14, 12)
	pdf.SetAutoPageBreak(true, 14)
	pdf.AddPage()
	pageW, _ := pdf.GetPageSize()

	if len(logoBytes) > 0 {
		if normalised, mime, err := normaliseLogoForPDF(logoBytes, logoMime); err == nil {
			logoBytes = normalised
			logoMime = mime
		} else {
			logoBytes = nil
			logoMime = ""
		}
	}

	leftMargin, topMargin, rightMargin, _ := pdf.GetMargins()
	logoEdge := 12.0
	logoX := leftMargin
	logoY := topMargin
	textX := logoX + logoEdge + 6

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 12)

	if len(logoBytes) > 0 {
		imgType := imageTypeFromMime(logoMime)
		opt := gofpdf.ImageOptions{ImageType: imgType, ReadDpi: true}
		if info := pdf.RegisterImageOptionsReader("brand_logo", opt, bytes.NewReader(logoBytes)); info != nil {
			pdf.ImageOptions("brand_logo", logoX, logoY, logoEdge, 0, false, opt, 0, "")
		}
	} else {
		radius := logoEdge / 2
		cx := logoX + radius
		cy := logoY + radius
		pdf.SetFillColor(220, 220, 220)
		pdf.Circle(cx, cy, radius, "F")
		pdf.SetFillColor(255, 255, 255)
		initial := brandInitial(brandName)
		pdf.SetFont("Arial", "B", 9)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetXY(logoX, cy-3)
		pdf.CellFormat(logoEdge, 6, initial, "", 0, "C", false, 0, "")
		pdf.SetFont("Arial", "B", 12)
	}

	pdf.SetXY(textX, logoY)
	pdf.CellFormat(pageW-textX-rightMargin, 5, brandName, "", 0, "L", false, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.SetXY(textX, logoY+6)
	pdf.CellFormat(pageW-textX-rightMargin, 4, fmt.Sprintf("Kode Order: %s", order.Code), "", 0, "L", false, 0, "")

	headerBottom := math.Max(logoY+logoEdge, logoY+10)
	pdf.SetY(headerBottom + 4)

	labelWidth := 28.0
	valueWidth := pageW - leftMargin - rightMargin - labelWidth
	writeRow := func(label, value string) {
		startX := pdf.GetX()
		startY := pdf.GetY()
		pdf.SetFont("Arial", "B", 9)
		pdf.CellFormat(labelWidth, 5, label, "", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "", 9)
		pdf.SetXY(startX+labelWidth, startY)
		pdf.MultiCell(valueWidth, 5, value, "", "L", false)
		rowHeight := pdf.GetY() - startY
		pdf.SetXY(startX, startY+rowHeight+1)
	}

	service := strings.TrimSpace(order.Shipment.ServiceLevel)
	courierLine := order.Shipment.Courier
	if courierLine == "" {
		courierLine = "Kurir belum ditentukan"
	}
	if service != "" {
		courierLine = fmt.Sprintf("%s · %s", courierLine, service)
	}
	writeRow("Kurir", courierLine)

	recipientName := strings.TrimSpace(recipient.Name)
	if recipientName == "" {
		recipientName = "Penerima belum ditentukan"
	}
	writeRow("Penerima", recipientName)

	phoneValue := strings.TrimSpace(recipient.Phone)
	if phoneValue == "" {
		phoneValue = "Tidak tersedia"
	}
	writeRow("Telp", phoneValue)

	addressLines := buildRecipientAddress(recipient)
	if len(addressLines) == 0 {
		addressLines = []string{"Alamat belum diisi."}
	}
	writeRow("Alamat", strings.Join(addressLines, "\n"))

	writeRow("Produk", describeOrderItems(order.Items))

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func imageTypeFromMime(mime string) string {
	switch strings.ToLower(mime) {
	case "image/jpeg", "image/jpg":
		return "JPG"
	case "image/png":
		return "PNG"
	case "image/gif":
		return "GIF"
	default:
		return "PNG"
	}
}

func formatCustomerLocation(c *domain.Customer) string {
	if c == nil {
		return ""
	}
	var segments []string
	if city := strings.TrimSpace(c.City); city != "" {
		segments = append(segments, city)
	}
	if province := strings.TrimSpace(c.Province); province != "" {
		segments = append(segments, province)
	}
	location := strings.Join(segments, ", ")
	if postal := strings.TrimSpace(c.Postal); postal != "" {
		if location != "" {
			location = fmt.Sprintf("%s %s", location, postal)
		} else {
			location = postal
		}
	}
	return strings.TrimSpace(location)
}

func normaliseLogoForPDF(data []byte, mime string) ([]byte, string, error) {
	trimmedMime := strings.TrimSpace(strings.ToLower(mime))
	var img image.Image
	var err error
	if trimmedMime == "image/webp" {
		img, err = webp.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, "", fmt.Errorf("decode webp logo: %w", err)
		}
	} else {
		img, _, err = image.Decode(bytes.NewReader(data))
		if err != nil {
			if alt, altErr := webp.Decode(bytes.NewReader(data)); altErr == nil {
				img = alt
			} else {
				return nil, "", fmt.Errorf("decode logo: %w", err)
			}
		}
	}

	square := cropSquare(img)
	circular := applyCircularMask(square)

	var buf bytes.Buffer
	if err := png.Encode(&buf, circular); err != nil {
		return nil, "", fmt.Errorf("encode png logo: %w", err)
	}
	return buf.Bytes(), "image/png", nil
}

func buildRecipientAddress(c *domain.Customer) []string {
	if c == nil {
		return nil
	}
	var lines []string
	if addr := strings.TrimSpace(c.Address); addr != "" {
		addr = strings.ReplaceAll(addr, "\r", "")
		for _, line := range strings.Split(addr, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				lines = append(lines, line)
			}
		}
	}
	if location := formatCustomerLocation(c); location != "" {
		lines = append(lines, location)
	}
	return lines
}

func cropSquare(img image.Image) image.Image {
	b := img.Bounds()
	size := b.Dx()
	if b.Dy() < size {
		size = b.Dy()
	}
	startX := b.Min.X + (b.Dx()-size)/2
	startY := b.Min.Y + (b.Dy()-size)/2
	square := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			square.Set(x, y, img.At(startX+x, startY+y))
		}
	}
	return square
}

func applyCircularMask(img image.Image) image.Image {
	b := img.Bounds()
	size := b.Dx()
	masked := image.NewRGBA(image.Rect(0, 0, size, size))
	radius := float64(size) / 2
	cx := radius - 0.5
	cy := radius - 0.5
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			if dx*dx+dy*dy <= radius*radius {
				masked.Set(x, y, img.At(b.Min.X+x, b.Min.Y+y))
			} else {
				masked.Set(x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}
	return masked
}

func brandInitial(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ""
	}
	r, _ := utf8.DecodeRuneInString(trimmed)
	if r == utf8.RuneError {
		return strings.ToUpper(trimmed[:1])
	}
	return strings.ToUpper(string(r))
}

func describeOrderItems(items []domain.OrderItem) string {
	if len(items) == 0 {
		return "Tidak ada produk"
	}
	var entries []string
	limit := 3
	for idx, item := range items {
		if idx == limit {
			break
		}
		label := strings.TrimSpace(item.SKU)
		if label == "" {
			label = strings.TrimSpace(item.ProductID)
			if len(label) > 8 {
				label = label[:8]
			}
		}
		if label == "" {
			label = fmt.Sprintf("Item %d", idx+1)
		}
		entries = append(entries, fmt.Sprintf("%s × %d", label, item.Quantity))
	}
	if len(items) > limit {
		entries = append(entries, fmt.Sprintf("+%d item lainnya", len(items)-limit))
	}
	return strings.Join(entries, ", ")
}
