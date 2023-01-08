package models

type YAMLParameter struct {
	Name        string
	Description string
	Required    bool
	Type        string
	Format      string
	In          string
}
