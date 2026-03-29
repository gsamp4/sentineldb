package handlers

type CreateAssetRequest struct {
	Type  string `json:"type" validate:"required,oneof=ip domain email"`
	Value string `json:"value" validate:"required"`
	Label string `json:"label"`
}

type AssetResponse struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Value     string `json:"value"`
	Label     string `json:"label"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
}