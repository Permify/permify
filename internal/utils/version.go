package utils

// Version -
type Version string

// String -
func (v Version) String() string {
	return string(v)
}

// IsSet -
func (v Version) IsSet() bool {
	return v != ""
}
