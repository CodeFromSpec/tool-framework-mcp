package parsing

import "errors"

var (
	ErrFileUnreadable = errors.New("file unreadable")
	ErrMalformedYAML  = errors.New("malformed YAML")

	ErrNotASpecReference                   = errors.New("not a SPEC/ reference")
	ErrHasQualifier                        = errors.New("logical name has qualifier")
	ErrUnexpectedContentBeforeFirstHeading = errors.New("unexpected content before first heading")
	ErrNodeNameDoesNotMatch                = errors.New("node name does not match")
	ErrDuplicatePublicSection              = errors.New("duplicate # Public section")
	ErrDuplicateAgentSection               = errors.New("duplicate # Agent section")
	ErrDuplicatePrivateSection             = errors.New("duplicate # Private section")
	ErrUnrecognizedSection                 = errors.New("unrecognized section")
	ErrDuplicateSubsection                 = errors.New("duplicate subsection")

	ErrUnrecognizedPrefix = errors.New("unrecognized prefix")
	ErrInvalidName        = errors.New("invalid name")
	ErrNoOutput           = errors.New("no output declared")
	ErrInvalidPath        = errors.New("invalid path")
)
</content>
</invoke>
