package models

type StoreLinkRequest struct {
	LinkTypeID string  `json:"link_type_id"`
	InwardID   string  `json:"inward_id"`
	OutwardID  string  `json:"outward_id"`
	Notes      *string `json:"notes"`
}
