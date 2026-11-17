package presentation

type EntitlementResponse struct {
	WorkspaceID string                `json:"workspace_id"`
	Plan        PlanDTO               `json:"plan"`
	Limits      map[string]int64      `json:"limits"`
	Features    map[string]FeatureDTO `json:"features"`
	Usage       UsageDTO              `json:"usage"`
}

type PlanDTO struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type FeatureDTO struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type UsageDTO struct {
	Pages UsageItemDTO `json:"pages"`
}

type UsageItemDTO struct {
	Used      int64  `json:"used"`
	Limit     int64  `json:"limit"`
	Remaining *int64 `json:"remaining"`
	Unlimited bool   `json:"unlimited"`
}

