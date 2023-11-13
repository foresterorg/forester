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
	RootPwSnippetKind   = iota
	SecuritySnippetKind = iota
	LocaleSnippetKind   = iota
)

var AllSnippetKinds = []SnippetKind{
	DiskSnippetKind,
	PostSnippetKind,
	RootPwSnippetKind,
	SecuritySnippetKind,
	LocaleSnippetKind,
}

func ParseSnippetKind(i int16) SnippetKind {
	switch i {
	case 0:
		return ReservedSnippetKind
	case 1:
		return DiskSnippetKind
	case 2:
		return PostSnippetKind
	case 3:
		return RootPwSnippetKind
	case 4:
		return SecuritySnippetKind
	case 5:
		return LocaleSnippetKind
	default:
		return -1
	}
}

func (sk SnippetKind) String() string {
	switch sk {
	case DiskSnippetKind:
		return "disk"
	case PostSnippetKind:
		return "post"
	case RootPwSnippetKind:
		return "rootpw"
	case SecuritySnippetKind:
		return "security"
	case LocaleSnippetKind:
		return "locale"
	}
	return ""
}
