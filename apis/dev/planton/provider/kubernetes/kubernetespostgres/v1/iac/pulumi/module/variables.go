package module

var vars = struct {
	TeamId                  string
	PostgresVersion         string
	PostgresPodSectorLabels map[string]string
}{
	// TeamId required by zalando operator
	// a random word has been chosen here. This could be any valid string.
	TeamId:          "db",
	PostgresVersion: "14",
	//all the pods created for the postgres database have this label.
	PostgresPodSectorLabels: map[string]string{
		"application": "spilo",
	},
}
