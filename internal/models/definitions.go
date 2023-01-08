package models

// YAMLParameter is a parameter in a YAML file
type YAMLDefinition struct {
	Description string
	Type        string
	Properties  []YAMLProperty
}
