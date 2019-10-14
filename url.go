package main

import (
	"time"
)

type Url struct {
	ShortURL string    `bson:"short_url" json:"short_url"`
	LongURL  string    `bson:"url" json:"url"`
	Expiry   time.Time `bson:"expires_at" json:"expires_at"`
	Validity int64     `json:"valid_for"`
}
