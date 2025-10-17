package models

import "time"

type Test struct {
  ID        int       `json:"id"`
  Code      string    `json:"code"`
  Name      string    `json:"name"`
  CreatedBy int       `json:"created_by"`
  CreatedAt time.Time `json:"created_at"`
}


