package domain

import (
	"encoding/json"
	"time"
)

type Plan struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type Feature struct {
	Key   string          `json:"key"`
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value"`
}

type Usage struct {
	Pages int64 `json:"pages"`
}

type Assignment struct {
	StartsAt time.Time  `json:"starts_at"`
	EndsAt   *time.Time `json:"ends_at"`
}

type Entitlements struct {
	WorkspaceID string             `json:"workspace_id"`
	Plan        Plan               `json:"plan"`
	Limits      map[string]int64   `json:"limits"`
	Features    map[string]Feature `json:"features"`
	Usage       Usage              `json:"usage"`
	Assignment  Assignment         `json:"assignment"`
}

const Unlimited int64 = -1

