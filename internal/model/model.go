package model

import (
	"time"
	"gorm.io/gorm"
)

type PaymentAccount struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Name        string         `json:"name"`
	PromptpayID string         `json:"promptpay_id"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	
	// Analytics (Calculated or Virtual)
	TotalOrders  int64   `gorm:"-" json:"total_orders"`
	TotalRevenue float64 `gorm:"-" json:"total_revenue"`
}

type Brand struct {
	gorm.Model
	Name string `json:"name" gorm:"unique"`
}

type Item struct {
	gorm.Model
	Title                  string  `json:"title"`
	Price                  float32 `json:"price"`
	Discount               float32 `json:"discount"`
	Discount_type          string  `json:"discount_type"`
	Description            string  `json:"description"`
	Material               string  `json:"material"`
	Is_preorder            int     `json:"is_preorder"`
	Hidden                 int     `json:"hidden"`
	Last_edited_by_user_id uint    `json:"last_edited_by_user_id"`
	Images                 []Image `json:"images"`
	Colors                 []Color `json:"colors"`
	Sizes                  []Size  `json:"sizes"`
	Brand_id               uint    `json:"brand_id"`
	Brand                  Brand   `json:"brand"`
	Page_id                uint    `json:"page_id"`
	
	PaymentAccountID uint            `json:"payment_account_id"`
	PaymentAccount   *PaymentAccount `json:"payment_account,omitempty"`
}

type Image struct {
	gorm.Model
	Url     string `json:"url"`
	Item_id uint   `json:"item_id"`
}

type Size struct {
	gorm.Model
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
	Item_id  uint   `json:"item_id"`
	Color_id uint   `json:"color_id"`
}

type Color struct {
	gorm.Model
	Color   string `json:"color"`
	Item_id uint   `json:"item_id"`
}

type Customer struct {
	UUID      string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"uuid"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	GoogleID  string         `gorm:"unique;not null" json:"-"`
	Email     string         `gorm:"unique;not null" json:"email"`
	Name      string         `json:"name"`
	AvatarURL string         `json:"avatar_url"`
	Role      string         `gorm:"default:'customer'" json:"role"` // "customer" | "admin"
}

type Preorder struct {
	gorm.Model
	CustomerUUID   *string     `gorm:"type:uuid" json:"customer_uuid"`
	Customer       *Customer   `json:"customer,omitempty"`
	CustomerName   string      `json:"customer_name"`
	Social         string      `json:"social"`
	ContactNumber  string      `json:"contact_number"`
	ShippingMethod string      `json:"shipping_method"` // "pickup" or "postal"
	Address        string      `json:"address"`
	Items          []OrderItem `json:"items"`
	TotalPrice     float32     `json:"total_price"`
	ShippingCost   float32     `json:"shipping_cost"`
	Completed      int         `json:"completed" gorm:"default:0"`
	PaymentSlipURL string      `json:"payment_slip_url"`
	Status         string      `gorm:"default:'placed'" json:"status"` // placed|confirmed|packed|shipped|delivered
	TrackingNo     string      `json:"tracking_no"`
	Note           string      `json:"note"` // admin note to customer
}

type OrderItem struct {
	gorm.Model
	PreorderID uint    `json:"preorder_id"`
	ItemID     uint    `json:"item_id"`
	Item       Item    `json:"item"`
	Size       string  `json:"size"`
	Color      string  `json:"color"`
	Quantity   int     `json:"quantity"`
	Price      float32 `json:"price"` // Price at time of order
}

type Page struct {
	gorm.Model
	Slug         string `json:"slug" gorm:"unique"`
	Text         string `json:"text"`
	Order        int    `json:"order"`
	Is_Permanent int    `json:"is_permanent"`
}

type Site struct {
	gorm.Model
	Image_url string `json:"image_url"`
}
