package model

import (
	"github.com/google/uuid"
	"time"
)

type Installation struct {
	// Required auto-generated PK.
	ID int64 `db:"id"`

	// An informative UUID, used for logging
	UUID uuid.UUID `db:"uuid"`

	// Installation state.
	State InstallState `db:"state"`

	// ValidUntil is time until installation is valid, otherwise it is expired.
	ValidUntil time.Time `db:"valid_until"`

	// QueuedAt is time when system was queued for installation.
	QueuedAt time.Time `db:"queued_at"`

	// The system.
	SystemID int64 `db:"system_id"`

	// The image.
	ImageID int64 `db:"image_id"`

	// SnippetText is a custom snippet, can be blank.
	SnippetText string `db:"snippet_text"`

	// KickstartOverride fully overrides kickstart, can be blank.
	KickstartOverride string `db:"kickstart_override"`

	// Comment, can be blank.
	Comment string `db:"comment"`
}

type InstallState int16

const (
	UnknownInstallState    InstallState = 0
	QueuedInstallState     InstallState = 100
	StartedInstallState    InstallState = 200
	BootingInstallState    InstallState = 300
	InstallingInstallState InstallState = 400
	FinishedInstallState   InstallState = 500
)

func ParseInstallState(i int16) InstallState {
	switch i {
	case 0:
		return UnknownInstallState
	case 100:
		return QueuedInstallState
	case 200:
		return StartedInstallState
	case 300:
		return BootingInstallState
	case 400:
		return InstallingInstallState
	case 500:
		return FinishedInstallState
	default:
		return -1
	}
}

func (is InstallState) String() string {
	switch is {
	case QueuedInstallState:
		return "queued"
	case StartedInstallState:
		return "started"
	case BootingInstallState:
		return "booting"
	case InstallingInstallState:
		return "installing"
	case FinishedInstallState:
		return "finished"
	}
	return ""
}
