// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package database

import (
	"time"

	"github.com/google/uuid"
)

type BatchStat struct {
	ID            uuid.UUID
	SymID         uuid.UUID
	StartTime     int64
	EndTime       int64
	Open          float64
	Close         float64
	High          float64
	Low           float64
	Volume        float64
	PeriodMinutes int32
}

type CryptoTick struct {
	ID          uuid.UUID
	SymID       uuid.UUID
	Price       string
	Time        int64
	Volume      string
	DailyChange string
	DailyDiff   string
	CreatedAt   time.Time
}

type Exchange struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Code      string
	Currency  string
	Country   string
	Iso2      string
	Iso3      string
	OpMic     string
}

type Symbol struct {
	ID           uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Code         string
	Name         string
	ExchangeName string
	ExchangeID   uuid.UUID
	Currency     string
	Type         string
}
