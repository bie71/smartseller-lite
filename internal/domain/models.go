package domain

import "time"

type CustomerType string

const (
	CustomerTypeCustomer CustomerType = "customer"
	CustomerTypeMarketer CustomerType = "marketer"
	CustomerTypeReseller CustomerType = "reseller"
)

type Product struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	SKU               string     `json:"sku"`
	CostPrice         float64    `json:"costPrice"`
	SalePrice         float64    `json:"salePrice"`
	Stock             int        `json:"stock"`
	Category          string     `json:"category"`
	LowStockThreshold int        `json:"lowStockThreshold"`
	Description       string     `json:"description"`
	ImagePath         string     `json:"imagePath"`
	ThumbPath         string     `json:"thumbPath"`
	ImageURL          string     `json:"imageUrl"`
	ThumbURL          string     `json:"thumbUrl"`
	ImageHash         string     `json:"imageHash"`
	ImageWidth        int        `json:"imageWidth"`
	ImageHeight       int        `json:"imageHeight"`
	ImageSizeBytes    int64      `json:"imageSizeBytes"`
	ThumbWidth        int        `json:"thumbWidth"`
	ThumbHeight       int        `json:"thumbHeight"`
	ThumbSizeBytes    int64      `json:"thumbSizeBytes"`
	ImageData         string     `json:"imageData,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	DeletedAt         *time.Time `json:"deletedAt"`
}

type Customer struct {
	ID        string       `json:"id"`
	Type      CustomerType `json:"type"`
	Name      string       `json:"name"`
	Phone     string       `json:"phone"`
	Email     string       `json:"email"`
	Address   string       `json:"address"`
	City      string       `json:"city"`
	Province  string       `json:"province"`
	Postal    string       `json:"postal"`
	Notes     string       `json:"notes"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

// Order aggregates order items and shipment metadata.
type Order struct {
	ID          string      `json:"id"`
	Code        string      `json:"code"`
	BuyerID     string      `json:"buyerId"`
	RecipientID string      `json:"recipientId"`
	Shipment    Shipment    `json:"shipment"`
	Items       []OrderItem `json:"items"`
	Discount    float64     `json:"discount"`
	Total       float64     `json:"total"`
	Profit      float64     `json:"profit"`
	Notes       string      `json:"notes"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type OrderItem struct {
	ID        string  `json:"id"`
	OrderID   string  `json:"orderId"`
	ProductID string  `json:"productId"`
	SKU       string  `json:"sku"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
	Discount  float64 `json:"discount"`
	CostPrice float64 `json:"costPrice"`
	Profit    float64 `json:"profit"`
}

type Shipment struct {
	Courier      string  `json:"courier"`
	TrackingCode string  `json:"trackingCode"`
	ServiceLevel string  `json:"serviceLevel"`
	ShippingCost float64 `json:"shippingCost"`
}

// LabelData is a flattened view to render PDF labels.
type LabelData struct {
	OrderCode       string    `json:"orderCode"`
	Courier         string    `json:"courier"`
	RecipientName   string    `json:"recipientName"`
	RecipientAddr   string    `json:"recipientAddr"`
	RecipientCity   string    `json:"recipientCity"`
	RecipientProv   string    `json:"recipientProv"`
	RecipientPostal string    `json:"recipientPostal"`
	RecipientPhone  string    `json:"recipientPhone"`
	CreatedAt       time.Time `json:"createdAt"`
}

// StockMutation tracks manual stock adjustments for audit.
type StockMutation struct {
	ID        string    `json:"id"`
	ProductID string    `json:"productId"`
	Delta     int       `json:"delta"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"createdAt"`
}

// StockOpname captures the result of a stock take session.
type StockOpname struct {
	ID          string            `json:"id"`
	Note        string            `json:"note"`
	PerformedBy string            `json:"performedBy"`
	PerformedAt time.Time         `json:"performedAt"`
	CreatedAt   time.Time         `json:"createdAt"`
	Items       []StockOpnameItem `json:"items"`
}

// StockOpnameItem records the difference for a single product during stock take.
type StockOpnameItem struct {
	ID            string `json:"id"`
	StockOpnameID string `json:"stockOpnameId"`
	ProductID     string `json:"productId"`
	ProductName   string `json:"productName"`
	ProductSKU    string `json:"productSku"`
	Counted       int    `json:"counted"`
	PreviousStock int    `json:"previousStock"`
	Difference    int    `json:"difference"`
}

// AppSettings represents lightweight configuration persisted for the desktop app.
type AppSettings struct {
	BrandName     string `json:"brandName"`
	LogoPath      string `json:"logoPath"`
	LogoURL       string `json:"logoUrl"`
	LogoHash      string `json:"logoHash"`
	LogoWidth     int    `json:"logoWidth"`
	LogoHeight    int    `json:"logoHeight"`
	LogoSizeBytes int64  `json:"logoSizeBytes"`
	LogoMime      string `json:"logoMime"`
	LogoData      string `json:"logoData,omitempty"`
}

// Courier represents an expedition/shipping partner.
type Courier struct {
	ID            string    `json:"id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Services      string    `json:"services"`
	TrackingURL   string    `json:"trackingUrl"`
	Contact       string    `json:"contact"`
	Notes         string    `json:"notes"`
	LogoPath      string    `json:"logoPath"`
	LogoURL       string    `json:"logoUrl"`
	LogoHash      string    `json:"logoHash"`
	LogoWidth     int       `json:"logoWidth"`
	LogoHeight    int       `json:"logoHeight"`
	LogoSizeBytes int64     `json:"logoSizeBytes"`
	LogoMime      string    `json:"logoMime"`
	LogoData      string    `json:"logoData,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
