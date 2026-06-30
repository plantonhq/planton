package module

var vars = struct {
	MongodbRootPasswordKey string
	RootUsername           string
	MongoDbPort            int
	MongoDBVersion         string
	CRVersion              string
	ReplicaSetName         string
}{
	MongodbRootPasswordKey: "MONGODB_DATABASE_ADMIN_PASSWORD",
	RootUsername:           "databaseAdmin",
	MongoDbPort:            27017,
	MongoDBVersion:         "8.0.12-4",
	CRVersion:              "1.20.1",
	ReplicaSetName:         "rs0",
}
