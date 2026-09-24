package keycloak

import "testing"

// The operator OWNS the user-attribute mappers on the LDAP component -- it
// never relies on which ones the identity server's vendor default happens to
// create. The Active Directory default has no first-name mapper at all, so an
// operator that only overrode what it found left the manifest's
// firstNameAttribute a silent no-op on every AD directory (caught live: a
// person imported with no first name). These pins hold the owned set and the
// shape of a mapper the operator creates.
func TestOwnedAttributeMappers_CoverEveryNameTheManifestDeclares(t *testing.T) {
	fed := &OwnedLDAPFederation{
		UsernameAttribute:  "sAMAccountName",
		EmailAttribute:     "mail",
		FirstNameAttribute: "givenName",
		LastNameAttribute:  "sn",
	}
	owned := fed.ownedAttributeMappers()

	want := map[string]ownedAttributeMapper{
		"username":   {userModelAttribute: "username", ldapAttribute: "sAMAccountName"},
		"email":      {userModelAttribute: "email", ldapAttribute: "mail"},
		"first name": {userModelAttribute: "firstName", ldapAttribute: "givenName"},
		"last name":  {userModelAttribute: "lastName", ldapAttribute: "sn"},
	}
	if len(owned) != len(want) {
		t.Fatalf("owned mapper set = %v, want exactly %d mappers", owned, len(want))
	}
	for name, expected := range want {
		got, ok := owned[name]
		if !ok {
			t.Errorf("owned set lacks %q", name)
			continue
		}
		if got != expected {
			t.Errorf("%q = %+v, want %+v", name, got, expected)
		}
	}
}

func TestLDAPAttributeMapperConfig_ReadOnlyNeverMandatory(t *testing.T) {
	cfg := ldapAttributeMapperConfig(ownedAttributeMapper{userModelAttribute: "firstName", ldapAttribute: "givenName"})

	expect := map[string]string{
		"user.model.attribute":        "firstName",
		"ldap.attribute":              "givenName",
		"read.only":                   "true",
		"always.read.value.from.ldap": "true",
		// A directory entry without the attribute (a service account with no
		// given name) must still import; whether an account may exist without
		// an email is the product's decision at sign-in, never the mapper's.
		"is.mandatory.in.ldap": "false",
		"is.binary.attribute":  "false",
	}
	for key, value := range expect {
		got := cfg[key]
		if len(got) != 1 || got[0] != value {
			t.Errorf("%s = %v, want [%s]", key, got, value)
		}
	}
}
