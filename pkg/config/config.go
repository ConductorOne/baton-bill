package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	Username = field.StringField(
		"username",
		field.WithDescription("The Bill username used to connect to the Bill API."),
		field.WithRequired(true),
	)

	Password = field.StringField(
		"password",
		field.WithDescription("The Bill password used to connect to the Bill API."),
		field.WithRequired(true),
		field.WithIsSecret(true),
	)

	OrganizationIDs = field.StringSliceField(
		"organizationIds",
		field.WithDescription("The Bill organizationIds used to connect to the Bill API."),
		field.WithRequired(true),
	)

	DeveloperKey = field.StringField(
		"developerKey",
		field.WithDescription("The Bill developerKey used to connect to the Bill API."),
		field.WithRequired(true),
		field.WithIsSecret(true),
	)
	BaseURLField = field.StringField(
		"base-url",
		field.WithDescription("Override the Bill API URL (for testing)"),
	)
)

//go:generate go run ./gen
var Config = field.NewConfiguration([]field.SchemaField{
	Username,
	Password,
	OrganizationIDs,
	DeveloperKey,
	BaseURLField,
})
