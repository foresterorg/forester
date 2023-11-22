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
	NetworkSnippetKind  = iota
	SourceSnippetKind   = iota
	DebugSnippetKind    = iota
)

var AllSnippetKinds = []SnippetKind{
	DiskSnippetKind,
	PostSnippetKind,
	RootPwSnippetKind,
	SecuritySnippetKind,
	LocaleSnippetKind,
	NetworkSnippetKind,
	SourceSnippetKind,
	DebugSnippetKind,
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
	case 6:
		return NetworkSnippetKind
	case 7:
		return SourceSnippetKind
	case 8:
		return DebugSnippetKind
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
	case NetworkSnippetKind:
		return "network"
	case SourceSnippetKind:
		return "source"
	}
	return ""
}
