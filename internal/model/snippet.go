package model

type Snippet struct {
	// Required auto-generated PK.
	ID int64 `db:"id"`

	// User-facing name. Required.
	Name string `db:"name"`

	// Kind is appliance type
	Kind SnippetKind `db:"kind"`

	// Snippet contents
	Body string `db:"body"`
}

type SnippetKind int16

const (
	ReservedSnippetKind = iota
	DiskSnippetKind     = iota
	PostSnippetKind     = iota
)

func ParseSnippetKind(i int16) SnippetKind {
	switch i {
	case 0:
		return ReservedSnippetKind
	case 1:
		return DiskSnippetKind
	case 2:
		return PostSnippetKind
	default:
		return -1
	}
}
