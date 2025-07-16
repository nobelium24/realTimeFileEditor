package repositories

type ResourceType string

const (
	ImageResource ResourceType = "image"
	VideoResource ResourceType = "video"
	RawResource   ResourceType = "raw"
)

type UploadedMedia struct {
	URL       string `json:"url"`
	PublicID  string `json:"public_id"`
	SecureURL string `json:"secure_url"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Format    string `json:"format"`
}
