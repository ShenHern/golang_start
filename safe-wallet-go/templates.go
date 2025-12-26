package main

// EntryTemplate defines a template for creating new entries.
type EntryTemplate struct {
	Name   string
	Fields []EntryField
}

// Defines the preset templates for creating new entries.
var entryTemplates = []EntryTemplate{
	{
		Name: "Credit Card",
		Fields: []EntryField{
			{Name: "Cardholder Name", Type: FieldTypeGeneral},
			{Name: "Card Number", Type: FieldTypeGeneral},
			{Name: "Expiration Date", Type: FieldTypeGeneral},
			{Name: "CVV", Type: FieldTypePIN},
			{Name: "PIN", Type: FieldTypePIN},
		},
	},
	{
		Name: "Password",
		Fields: []EntryField{
			{Name: "Username", Type: FieldTypeGeneral},
			{Name: "Password", Type: FieldTypePassword},
			{Name: "URL", Type: FieldTypeGeneral},
			{Name: "Notes", Type: FieldTypeGeneral},
		},
	},
	{
		Name: "Note",
		Fields: []EntryField{
			{Name: "Note", Type: FieldTypeGeneral},
		},
	},
	{
		Name: "Bank Account",
		Fields: []EntryField{
			{Name: "Bank Name", Type: FieldTypeGeneral},
			{Name: "Account Type", Type: FieldTypeGeneral},
			{Name: "Account Holder Name", Type: FieldTypeGeneral},
			{Name: "Account Number", Type: FieldTypeGeneral},
			{Name: "Password", Type: FieldTypePassword},
			{Name: "PIN", Type: FieldTypePIN},
		},
	},
}
