package application

var defaults = map[string]any{
	"user_attributes": []string{
		"sub", "email", "name", "given_name", "family_name",
		"profile", "picture", "phone_number", "address", "created_at",
	},
	"scopes": []string{"openid", "profile", "email"},
}
