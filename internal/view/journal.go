package view

import (
	"github.com/lxn/walk"
	"time"
)

type Journal struct {
	walk.TableModelBase
	entries []JournalEntry
}

type JournalEntry struct {
	Time     time.Time
	Text     string
	LogLevel LogLevel
}

type LogLevel int

// Log levels.
const (
	DBG LogLevel = iota
	INF
	WRN
	ERR
)

func (x *Journal) RowCount() int {
	return len(x.entries)
}

func (x *Journal) Value(row, col int) interface{} {

	entry := x.entries[row]
	switch col {
	case 0:
		return entry.Time.Format("15:04:05")
	case 1:
		return entry.Text
	}
	return ""
}

func (x *Journal) StyleCell(c *walk.CellStyle) {

	switch c.Col() {
	case 0:
		c.TextColor = walk.RGB(0, 128, 0)
	case 1:
		switch x.entries[c.Row()].LogLevel {
		case INF:
			c.TextColor = walk.RGB(0, 0, 128)
		case ERR:
			c.TextColor = walk.RGB(255, 0, 0)
		case WRN:
			c.TextColor = walk.RGB(128, 0, 0)
		case DBG:
			c.TextColor = walk.RGB(105, 105, 105)
		}
	}
}
