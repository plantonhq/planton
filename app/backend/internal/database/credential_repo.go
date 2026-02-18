package database

import (
	"context"
	"fmt"
	"time"

	"github.com/plantonhq/openmcf/app/backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// CredentialCollectionName is the unified collection for all provider credentials
	CredentialCollectionName = "credentials"
)

// CredentialRepository provides unified data access for all provider credentials.
type CredentialRepository struct {
	collection *mongo.Collection
}

// NewCredentialRepository creates a new unified credential repository instance.
func NewCredentialRepository(db *MongoDB) *CredentialRepository {
	return &CredentialRepository{
		collection: db.Database.Collection(CredentialCollectionName),
	}
}

// CreateGcp creates a new GCP credential.
func (r *CredentialRepository) CreateGcp(ctx context.Context, name, serviceAccountKeyBase64 string) (*models.GcpCredential, error) {
	// Check if a credential for this provider already exists
	exists, err := r.ExistsByProvider(ctx, "gcp")
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing GCP credential: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("credential for provider 'gcp' already exists")
	}

	now := time.Now()
	credential := &models.GcpCredential{
		ID:                      primitive.NewObjectID(),
		Name:                    name,
		ServiceAccountKeyBase64: serviceAccountKeyBase64,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	// Store as document with provider field
	doc := bson.M{
		"_id":                        credential.ID,
		"name":                       credential.Name,
		"provider":                   "gcp",
		"service_account_key_base64": credential.ServiceAccountKeyBase64,
		"created_at":                 credential.CreatedAt,
		"updated_at":                 credential.UpdatedAt,
	}

	_, err = r.collection.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP credential: %w", err)
	}

	return credential, nil
}

// CreateAws creates a new AWS credential.
func (r *CredentialRepository) CreateAws(ctx context.Context, name, accountID, accessKeyID, secretAccessKey, region, sessionToken string) (*models.AwsCredential, error) {
	// Check if a credential for this provider already exists
	exists, err := r.ExistsByProvider(ctx, "aws")
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing AWS credential: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("credential for provider 'aws' already exists")
	}

	now := time.Now()
	credential := &models.AwsCredential{
		ID:              primitive.NewObjectID(),
		Name:            name,
		AccountID:       accountID,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		Region:          region,
		SessionToken:    sessionToken,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	doc := bson.M{
		"_id":               credential.ID,
		"name":              credential.Name,
		"provider":          "aws",
		"account_id":        credential.AccountID,
		"access_key_id":     credential.AccessKeyID,
		"secret_access_key": credential.SecretAccessKey,
		"created_at":        credential.CreatedAt,
		"updated_at":        credential.UpdatedAt,
	}
	if region != "" {
		doc["region"] = region
	}
	if sessionToken != "" {
		doc["session_token"] = sessionToken
	}

	_, err = r.collection.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS credential: %w", err)
	}

	return credential, nil
}

// CreateAzure creates a new Azure credential.
func (r *CredentialRepository) CreateAzure(ctx context.Context, name, clientID, clientSecret, tenantID, subscriptionID string) (*models.AzureCredential, error) {
	// Check if a credential for this provider already exists
	exists, err := r.ExistsByProvider(ctx, "azure")
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing Azure credential: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("credential for provider 'azure' already exists")
	}

	now := time.Now()
	credential := &models.AzureCredential{
		ID:             primitive.NewObjectID(),
		Name:           name,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		TenantID:       tenantID,
		SubscriptionID: subscriptionID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	doc := bson.M{
		"_id":             credential.ID,
		"name":            credential.Name,
		"provider":        "azure",
		"client_id":       credential.ClientID,
		"client_secret":   credential.ClientSecret,
		"tenant_id":       credential.TenantID,
		"subscription_id": credential.SubscriptionID,
		"created_at":      credential.CreatedAt,
		"updated_at":      credential.UpdatedAt,
	}

	_, err = r.collection.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	return credential, nil
}

// UpdateGcp updates an existing GCP credential.
func (r *CredentialRepository) UpdateGcp(ctx context.Context, id string, name, serviceAccountKeyBase64 string) (*models.GcpCredential, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"name":                       name,
			"service_account_key_base64": serviceAccountKeyBase64,
			"updated_at":                 now,
		},
	}

	result := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID, "provider": "gcp"}, update)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("GCP credential with ID '%s' not found", id)
	}
	if result.Err() != nil {
		return nil, fmt.Errorf("failed to update GCP credential: %w", result.Err())
	}

	// Fetch updated document
	doc, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("credential not found after update")
	}

	return convertToGcpCredential(doc)
}

// UpdateAws updates an existing AWS credential.
func (r *CredentialRepository) UpdateAws(ctx context.Context, id string, name, accountID, accessKeyID, secretAccessKey, region, sessionToken string) (*models.AwsCredential, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	now := time.Now()
	setFields := bson.M{
		"name":              name,
		"account_id":        accountID,
		"access_key_id":     accessKeyID,
		"secret_access_key": secretAccessKey,
		"updated_at":        now,
	}
	unsetFields := bson.M{}

	if region != "" {
		setFields["region"] = region
	} else {
		unsetFields["region"] = ""
	}
	if sessionToken != "" {
		setFields["session_token"] = sessionToken
	} else {
		unsetFields["session_token"] = ""
	}

	update := bson.M{"$set": setFields}
	if len(unsetFields) > 0 {
		update["$unset"] = unsetFields
	}

	result := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID, "provider": "aws"}, update)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("AWS credential with ID '%s' not found", id)
	}
	if result.Err() != nil {
		return nil, fmt.Errorf("failed to update AWS credential: %w", result.Err())
	}

	// Fetch updated document
	doc, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("credential not found after update")
	}

	return convertToAwsCredential(doc)
}

// CreateAuth0 creates a new Auth0 credential.
func (r *CredentialRepository) CreateAuth0(ctx context.Context, name, domain, clientID, clientSecret string) (*models.Auth0Credential, error) {
	// Check if a credential for this provider already exists
	exists, err := r.ExistsByProvider(ctx, "auth0")
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing Auth0 credential: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("credential for provider 'auth0' already exists")
	}

	now := time.Now()
	credential := &models.Auth0Credential{
		ID:           primitive.NewObjectID(),
		Name:         name,
		Domain:       domain,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	doc := bson.M{
		"_id":           credential.ID,
		"name":          credential.Name,
		"provider":      "auth0",
		"domain":        credential.Domain,
		"client_id":     credential.ClientID,
		"client_secret": credential.ClientSecret,
		"created_at":    credential.CreatedAt,
		"updated_at":    credential.UpdatedAt,
	}

	_, err = r.collection.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to create Auth0 credential: %w", err)
	}

	return credential, nil
}

// UpdateAuth0 updates an existing Auth0 credential.
func (r *CredentialRepository) UpdateAuth0(ctx context.Context, id string, name, domain, clientID, clientSecret string) (*models.Auth0Credential, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"name":          name,
			"domain":        domain,
			"client_id":     clientID,
			"client_secret": clientSecret,
			"updated_at":    now,
		},
	}

	result := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID, "provider": "auth0"}, update)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("Auth0 credential with ID '%s' not found", id)
	}
	if result.Err() != nil {
		return nil, fmt.Errorf("failed to update Auth0 credential: %w", result.Err())
	}

	// Fetch updated document
	doc, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("credential not found after update")
	}

	return convertToAuth0Credential(doc)
}

// UpdateAzure updates an existing Azure credential.
func (r *CredentialRepository) UpdateAzure(ctx context.Context, id string, name, clientID, clientSecret, tenantID, subscriptionID string) (*models.AzureCredential, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"name":            name,
			"client_id":       clientID,
			"client_secret":   clientSecret,
			"tenant_id":       tenantID,
			"subscription_id": subscriptionID,
			"updated_at":      now,
		},
	}

	result := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID, "provider": "azure"}, update)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("azure credential with ID '%s' not found", id)
	}
	if result.Err() != nil {
		return nil, fmt.Errorf("failed to update Azure credential: %w", result.Err())
	}

	// Fetch updated document
	doc, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("credential not found after update")
	}

	return convertToAzureCredential(doc)
}

// Delete deletes a credential by ID.
func (r *CredentialRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("credential with ID '%s' not found", id)
	}

	return nil
}

// FindFirstByProvider retrieves the first credential for a given provider.
func (r *CredentialRepository) FindFirstByProvider(ctx context.Context, provider string) (interface{}, error) {
	filter := bson.M{"provider": provider}

	var result bson.M
	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find %s credential: %w", provider, err)
	}

	// Convert to appropriate model based on provider
	switch provider {
	case "gcp":
		return convertToGcpCredential(result)
	case "aws":
		return convertToAwsCredential(result)
	case "azure":
		return convertToAzureCredential(result)
	case "auth0":
		return convertToAuth0Credential(result)
	case "openstack":
		return convertToOpenStackCredential(result)
	case "scaleway":
		return convertToScalewayCredential(result)
	case "alicloud":
		return convertToAlicloudCredential(result)
	case "oci":
		return convertToOciCredential(result)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// ExistsByProvider checks if a credential exists for the given provider.
func (r *CredentialRepository) ExistsByProvider(ctx context.Context, provider string) (bool, error) {
	filter := bson.M{"provider": provider}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check credential existence: %w", err)
	}
	return count > 0, nil
}

// FindByID retrieves a credential by ID.
func (r *CredentialRepository) FindByID(ctx context.Context, id string) (bson.M, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	var result bson.M
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil // Not found, but not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query credential by ID: %w", err)
	}
	return result, nil
}

// List retrieves all credentials with optional provider filter.
// Returns credential summaries (without sensitive data like keys/secrets).
func (r *CredentialRepository) List(ctx context.Context, provider *string) ([]bson.M, error) {
	filter := bson.M{}
	if provider != nil && *provider != "" {
		filter["provider"] = *provider
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode credentials: %w", err)
	}

	return results, nil
}

// Helper functions to convert bson.M to typed credentials
func convertToGcpCredential(doc bson.M) (*models.GcpCredential, error) {
	id, ok := doc["_id"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid _id field")
	}

	// Convert primitive.DateTime to time.Time
	var createdAt, updatedAt time.Time
	if dt, ok := doc["created_at"].(primitive.DateTime); ok {
		createdAt = dt.Time()
	}
	if dt, ok := doc["updated_at"].(primitive.DateTime); ok {
		updatedAt = dt.Time()
	}

	return &models.GcpCredential{
		ID:                      id,
		Name:                    doc["name"].(string),
		ServiceAccountKeyBase64: doc["service_account_key_base64"].(string),
		CreatedAt:               createdAt,
		UpdatedAt:               updatedAt,
	}, nil
}

func convertToAwsCredential(doc bson.M) (*models.AwsCredential, error) {
	id, ok := doc["_id"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid _id field")
	}

	// Convert primitive.DateTime to time.Time
	var createdAt, updatedAt time.Time
	if dt, ok := doc["created_at"].(primitive.DateTime); ok {
		createdAt = dt.Time()
	}
	if dt, ok := doc["updated_at"].(primitive.DateTime); ok {
		updatedAt = dt.Time()
	}

	cred := &models.AwsCredential{
		ID:              id,
		Name:            doc["name"].(string),
		AccountID:       doc["account_id"].(string),
		AccessKeyID:     doc["access_key_id"].(string),
		SecretAccessKey: doc["secret_access_key"].(string),
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}

	if region, ok := doc["region"].(string); ok {
		cred.Region = region
	}
	if sessionToken, ok := doc["session_token"].(string); ok {
		cred.SessionToken = sessionToken
	}

	return cred, nil
}

func convertToAzureCredential(doc bson.M) (*models.AzureCredential, error) {
	id, ok := doc["_id"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid _id field")
	}

	// Convert primitive.DateTime to time.Time
	var createdAt, updatedAt time.Time
	if dt, ok := doc["created_at"].(primitive.DateTime); ok {
		createdAt = dt.Time()
	}
	if dt, ok := doc["updated_at"].(primitive.DateTime); ok {
		updatedAt = dt.Time()
	}

	return &models.AzureCredential{
		ID:             id,
		Name:           doc["name"].(string),
		ClientID:       doc["client_id"].(string),
		ClientSecret:   doc["client_secret"].(string),
		TenantID:       doc["tenant_id"].(string),
		SubscriptionID: doc["subscription_id"].(string),
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}, nil
}

func convertToAuth0Credential(doc bson.M) (*models.Auth0Credential, error) {
	id, ok := doc["_id"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid _id field")
	}

	// Convert primitive.DateTime to time.Time
	var createdAt, updatedAt time.Time
	if dt, ok := doc["created_at"].(primitive.DateTime); ok {
		createdAt = dt.Time()
	}
	if dt, ok := doc["updated_at"].(primitive.DateTime); ok {
		updatedAt = dt.Time()
	}

	return &models.Auth0Credential{
		ID:           id,
		Name:         doc["name"].(string),
		Domain:       doc["domain"].(string),
		ClientID:     doc["client_id"].(string),
		ClientSecret: doc["client_secret"].(string),
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}

// CreateOpenStack creates a new OpenStack credential.
func (r *CredentialRepository) CreateOpenStack(ctx context.Context, cred *models.OpenStackCredential) (*models.OpenStackCredential, error) {
	exists, err := r.ExistsByProvider(ctx, "openstack")
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing OpenStack credential: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("credential for provider 'openstack' already exists")
	}

	now := time.Now()
	cred.ID = primitive.NewObjectID()
	cred.CreatedAt = now
	cred.UpdatedAt = now

	doc := bson.M{
		"_id":         cred.ID,
		"name":        cred.Name,
		"provider":    "openstack",
		"auth_url":    cred.AuthURL,
		"auth_method": cred.AuthMethod,
		"created_at":  cred.CreatedAt,
		"updated_at":  cred.UpdatedAt,
	}

	// Optional fields
	setOptionalString(doc, "region", cred.Region)
	setOptionalString(doc, "user_name", cred.UserName)
	setOptionalString(doc, "password", cred.Password)
	setOptionalString(doc, "application_credential_id", cred.ApplicationCredentialID)
	setOptionalString(doc, "application_credential_name", cred.ApplicationCredentialName)
	setOptionalString(doc, "application_credential_secret", cred.ApplicationCredentialSecret)
	setOptionalString(doc, "token", cred.Token)
	setOptionalString(doc, "tenant_name", cred.TenantName)
	setOptionalString(doc, "tenant_id", cred.TenantID)
	setOptionalString(doc, "user_domain_name", cred.UserDomainName)
	setOptionalString(doc, "user_domain_id", cred.UserDomainID)
	setOptionalString(doc, "project_domain_name", cred.ProjectDomainName)
	setOptionalString(doc, "project_domain_id", cred.ProjectDomainID)
	setOptionalString(doc, "cacert_file", cred.CACertFile)
	setOptionalString(doc, "endpoint_type", cred.EndpointType)
	if cred.Insecure {
		doc["insecure"] = cred.Insecure
	}

	_, err = r.collection.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenStack credential: %w", err)
	}

	return cred, nil
}

// UpdateOpenStack updates an existing OpenStack credential.
func (r *CredentialRepository) UpdateOpenStack(ctx context.Context, id string, cred *models.OpenStackCredential) (*models.OpenStackCredential, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	now := time.Now()
	setFields := bson.M{
		"name":        cred.Name,
		"auth_url":    cred.AuthURL,
		"auth_method": cred.AuthMethod,
		"updated_at":  now,
	}

	// Set all optional fields (overwrite with new values or empty)
	setFields["region"] = cred.Region
	setFields["user_name"] = cred.UserName
	setFields["password"] = cred.Password
	setFields["application_credential_id"] = cred.ApplicationCredentialID
	setFields["application_credential_name"] = cred.ApplicationCredentialName
	setFields["application_credential_secret"] = cred.ApplicationCredentialSecret
	setFields["token"] = cred.Token
	setFields["tenant_name"] = cred.TenantName
	setFields["tenant_id"] = cred.TenantID
	setFields["user_domain_name"] = cred.UserDomainName
	setFields["user_domain_id"] = cred.UserDomainID
	setFields["project_domain_name"] = cred.ProjectDomainName
	setFields["project_domain_id"] = cred.ProjectDomainID
	setFields["insecure"] = cred.Insecure
	setFields["cacert_file"] = cred.CACertFile
	setFields["endpoint_type"] = cred.EndpointType

	update := bson.M{"$set": setFields}

	result := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID, "provider": "openstack"}, update)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("OpenStack credential with ID '%s' not found", id)
	}
	if result.Err() != nil {
		return nil, fmt.Errorf("failed to update OpenStack credential: %w", result.Err())
	}

	doc, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("credential not found after update")
	}

	return convertToOpenStackCredential(doc)
}

func convertToOpenStackCredential(doc bson.M) (*models.OpenStackCredential, error) {
	id, ok := doc["_id"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid _id field")
	}

	var createdAt, updatedAt time.Time
	if dt, ok := doc["created_at"].(primitive.DateTime); ok {
		createdAt = dt.Time()
	}
	if dt, ok := doc["updated_at"].(primitive.DateTime); ok {
		updatedAt = dt.Time()
	}

	cred := &models.OpenStackCredential{
		ID:        id,
		Name:      doc["name"].(string),
		AuthURL:   doc["auth_url"].(string),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if v, ok := doc["auth_method"].(string); ok {
		cred.AuthMethod = v
	}
	if v, ok := doc["region"].(string); ok {
		cred.Region = v
	}
	if v, ok := doc["user_name"].(string); ok {
		cred.UserName = v
	}
	if v, ok := doc["password"].(string); ok {
		cred.Password = v
	}
	if v, ok := doc["application_credential_id"].(string); ok {
		cred.ApplicationCredentialID = v
	}
	if v, ok := doc["application_credential_name"].(string); ok {
		cred.ApplicationCredentialName = v
	}
	if v, ok := doc["application_credential_secret"].(string); ok {
		cred.ApplicationCredentialSecret = v
	}
	if v, ok := doc["token"].(string); ok {
		cred.Token = v
	}
	if v, ok := doc["tenant_name"].(string); ok {
		cred.TenantName = v
	}
	if v, ok := doc["tenant_id"].(string); ok {
		cred.TenantID = v
	}
	if v, ok := doc["user_domain_name"].(string); ok {
		cred.UserDomainName = v
	}
	if v, ok := doc["user_domain_id"].(string); ok {
		cred.UserDomainID = v
	}
	if v, ok := doc["project_domain_name"].(string); ok {
		cred.ProjectDomainName = v
	}
	if v, ok := doc["project_domain_id"].(string); ok {
		cred.ProjectDomainID = v
	}
	if v, ok := doc["insecure"].(bool); ok {
		cred.Insecure = v
	}
	if v, ok := doc["cacert_file"].(string); ok {
		cred.CACertFile = v
	}
	if v, ok := doc["endpoint_type"].(string); ok {
		cred.EndpointType = v
	}

	return cred, nil
}

// CreateScaleway creates a new Scaleway credential.
func (r *CredentialRepository) CreateScaleway(ctx context.Context, cred *models.ScalewayCredential) (*models.ScalewayCredential, error) {
	exists, err := r.ExistsByProvider(ctx, "scaleway")
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing Scaleway credential: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("credential for provider 'scaleway' already exists")
	}

	now := time.Now()
	cred.ID = primitive.NewObjectID()
	cred.CreatedAt = now
	cred.UpdatedAt = now

	doc := bson.M{
		"_id":        cred.ID,
		"name":       cred.Name,
		"provider":   "scaleway",
		"access_key": cred.AccessKey,
		"secret_key": cred.SecretKey,
		"created_at": cred.CreatedAt,
		"updated_at": cred.UpdatedAt,
	}

	setOptionalString(doc, "project_id", cred.ProjectID)
	setOptionalString(doc, "organization_id", cred.OrganizationID)
	setOptionalString(doc, "region", cred.Region)
	setOptionalString(doc, "zone", cred.Zone)

	_, err = r.collection.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to create Scaleway credential: %w", err)
	}

	return cred, nil
}

// UpdateScaleway updates an existing Scaleway credential.
func (r *CredentialRepository) UpdateScaleway(ctx context.Context, id string, cred *models.ScalewayCredential) (*models.ScalewayCredential, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	now := time.Now()
	setFields := bson.M{
		"name":       cred.Name,
		"access_key": cred.AccessKey,
		"secret_key": cred.SecretKey,
		"updated_at": now,
	}

	// Set all optional fields (overwrite with new values or empty)
	setFields["project_id"] = cred.ProjectID
	setFields["organization_id"] = cred.OrganizationID
	setFields["region"] = cred.Region
	setFields["zone"] = cred.Zone

	update := bson.M{"$set": setFields}

	result := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID, "provider": "scaleway"}, update)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("Scaleway credential with ID '%s' not found", id)
	}
	if result.Err() != nil {
		return nil, fmt.Errorf("failed to update Scaleway credential: %w", result.Err())
	}

	doc, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("credential not found after update")
	}

	return convertToScalewayCredential(doc)
}

func convertToScalewayCredential(doc bson.M) (*models.ScalewayCredential, error) {
	id, ok := doc["_id"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid _id field")
	}

	var createdAt, updatedAt time.Time
	if dt, ok := doc["created_at"].(primitive.DateTime); ok {
		createdAt = dt.Time()
	}
	if dt, ok := doc["updated_at"].(primitive.DateTime); ok {
		updatedAt = dt.Time()
	}

	cred := &models.ScalewayCredential{
		ID:        id,
		Name:      doc["name"].(string),
		AccessKey: doc["access_key"].(string),
		SecretKey: doc["secret_key"].(string),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if v, ok := doc["project_id"].(string); ok {
		cred.ProjectID = v
	}
	if v, ok := doc["organization_id"].(string); ok {
		cred.OrganizationID = v
	}
	if v, ok := doc["region"].(string); ok {
		cred.Region = v
	}
	if v, ok := doc["zone"].(string); ok {
		cred.Zone = v
	}

	return cred, nil
}

// CreateAlicloud creates a new Alibaba Cloud credential.
func (r *CredentialRepository) CreateAlicloud(ctx context.Context, cred *models.AlicloudCredential) (*models.AlicloudCredential, error) {
	exists, err := r.ExistsByProvider(ctx, "alicloud")
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing Alibaba Cloud credential: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("credential for provider 'alicloud' already exists")
	}

	now := time.Now()
	cred.ID = primitive.NewObjectID()
	cred.CreatedAt = now
	cred.UpdatedAt = now

	doc := bson.M{
		"_id":         cred.ID,
		"name":        cred.Name,
		"provider":    "alicloud",
		"auth_method": cred.AuthMethod,
		"created_at":  cred.CreatedAt,
		"updated_at":  cred.UpdatedAt,
	}

	setOptionalString(doc, "region", cred.Region)
	setOptionalString(doc, "account_id", cred.AccountId)
	setOptionalString(doc, "account_type", cred.AccountType)
	setOptionalString(doc, "access_key", cred.AccessKey)
	setOptionalString(doc, "secret_key", cred.SecretKey)
	setOptionalString(doc, "security_token", cred.SecurityToken)
	setOptionalString(doc, "ecs_role_name", cred.EcsRoleName)
	setOptionalString(doc, "role_arn", cred.RoleArn)
	setOptionalString(doc, "session_name", cred.SessionName)
	setOptionalString(doc, "policy", cred.Policy)
	if cred.SessionExpiration != 0 {
		doc["session_expiration"] = cred.SessionExpiration
	}
	setOptionalString(doc, "external_id", cred.ExternalId)
	setOptionalString(doc, "oidc_provider_arn", cred.OidcProviderArn)
	setOptionalString(doc, "oidc_token", cred.OidcToken)
	setOptionalString(doc, "oidc_token_file", cred.OidcTokenFile)
	setOptionalString(doc, "credentials_file", cred.CredentialsFile)
	setOptionalString(doc, "profile", cred.Profile)
	setOptionalString(doc, "credentials_uri", cred.CredentialsUri)

	_, err = r.collection.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to create Alibaba Cloud credential: %w", err)
	}

	return cred, nil
}

// UpdateAlicloud updates an existing Alibaba Cloud credential.
func (r *CredentialRepository) UpdateAlicloud(ctx context.Context, id string, cred *models.AlicloudCredential) (*models.AlicloudCredential, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	now := time.Now()
	setFields := bson.M{
		"name":               cred.Name,
		"auth_method":        cred.AuthMethod,
		"updated_at":         now,
		"region":             cred.Region,
		"account_id":         cred.AccountId,
		"account_type":       cred.AccountType,
		"access_key":         cred.AccessKey,
		"secret_key":         cred.SecretKey,
		"security_token":     cred.SecurityToken,
		"ecs_role_name":      cred.EcsRoleName,
		"role_arn":           cred.RoleArn,
		"session_name":       cred.SessionName,
		"policy":             cred.Policy,
		"session_expiration": cred.SessionExpiration,
		"external_id":        cred.ExternalId,
		"oidc_provider_arn":  cred.OidcProviderArn,
		"oidc_token":         cred.OidcToken,
		"oidc_token_file":    cred.OidcTokenFile,
		"credentials_file":   cred.CredentialsFile,
		"profile":            cred.Profile,
		"credentials_uri":    cred.CredentialsUri,
	}

	update := bson.M{"$set": setFields}

	result := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID, "provider": "alicloud"}, update)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("Alibaba Cloud credential with ID '%s' not found", id)
	}
	if result.Err() != nil {
		return nil, fmt.Errorf("failed to update Alibaba Cloud credential: %w", result.Err())
	}

	doc, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("credential not found after update")
	}

	return convertToAlicloudCredential(doc)
}

func convertToAlicloudCredential(doc bson.M) (*models.AlicloudCredential, error) {
	id, ok := doc["_id"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid _id field")
	}

	var createdAt, updatedAt time.Time
	if dt, ok := doc["created_at"].(primitive.DateTime); ok {
		createdAt = dt.Time()
	}
	if dt, ok := doc["updated_at"].(primitive.DateTime); ok {
		updatedAt = dt.Time()
	}

	cred := &models.AlicloudCredential{
		ID:        id,
		Name:      doc["name"].(string),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if v, ok := doc["auth_method"].(string); ok {
		cred.AuthMethod = v
	}
	if v, ok := doc["region"].(string); ok {
		cred.Region = v
	}
	if v, ok := doc["account_id"].(string); ok {
		cred.AccountId = v
	}
	if v, ok := doc["account_type"].(string); ok {
		cred.AccountType = v
	}
	if v, ok := doc["access_key"].(string); ok {
		cred.AccessKey = v
	}
	if v, ok := doc["secret_key"].(string); ok {
		cred.SecretKey = v
	}
	if v, ok := doc["security_token"].(string); ok {
		cred.SecurityToken = v
	}
	if v, ok := doc["ecs_role_name"].(string); ok {
		cred.EcsRoleName = v
	}
	if v, ok := doc["role_arn"].(string); ok {
		cred.RoleArn = v
	}
	if v, ok := doc["session_name"].(string); ok {
		cred.SessionName = v
	}
	if v, ok := doc["policy"].(string); ok {
		cred.Policy = v
	}
	if v, ok := doc["session_expiration"].(int32); ok {
		cred.SessionExpiration = v
	}
	if v, ok := doc["external_id"].(string); ok {
		cred.ExternalId = v
	}
	if v, ok := doc["oidc_provider_arn"].(string); ok {
		cred.OidcProviderArn = v
	}
	if v, ok := doc["oidc_token"].(string); ok {
		cred.OidcToken = v
	}
	if v, ok := doc["oidc_token_file"].(string); ok {
		cred.OidcTokenFile = v
	}
	if v, ok := doc["credentials_file"].(string); ok {
		cred.CredentialsFile = v
	}
	if v, ok := doc["profile"].(string); ok {
		cred.Profile = v
	}
	if v, ok := doc["credentials_uri"].(string); ok {
		cred.CredentialsUri = v
	}

	return cred, nil
}

// CreateOci creates a new OCI credential.
func (r *CredentialRepository) CreateOci(ctx context.Context, cred *models.OciCredential) (*models.OciCredential, error) {
	exists, err := r.ExistsByProvider(ctx, "oci")
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing OCI credential: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("credential for provider 'oci' already exists")
	}

	now := time.Now()
	cred.ID = primitive.NewObjectID()
	cred.CreatedAt = now
	cred.UpdatedAt = now

	doc := bson.M{
		"_id":         cred.ID,
		"name":        cred.Name,
		"provider":    "oci",
		"auth_method": cred.AuthMethod,
		"created_at":  cred.CreatedAt,
		"updated_at":  cred.UpdatedAt,
	}

	setOptionalString(doc, "region", cred.Region)
	setOptionalString(doc, "tenancy_ocid", cred.TenancyOcid)
	setOptionalString(doc, "user_ocid", cred.UserOcid)
	setOptionalString(doc, "fingerprint", cred.Fingerprint)
	setOptionalString(doc, "private_key", cred.PrivateKey)
	setOptionalString(doc, "private_key_password", cred.PrivateKeyPassword)
	setOptionalString(doc, "config_file_profile", cred.ConfigFileProfile)

	_, err = r.collection.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to create OCI credential: %w", err)
	}

	return cred, nil
}

// UpdateOci updates an existing OCI credential.
func (r *CredentialRepository) UpdateOci(ctx context.Context, id string, cred *models.OciCredential) (*models.OciCredential, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	now := time.Now()
	setFields := bson.M{
		"name":                 cred.Name,
		"auth_method":          cred.AuthMethod,
		"updated_at":           now,
		"region":               cred.Region,
		"tenancy_ocid":         cred.TenancyOcid,
		"user_ocid":            cred.UserOcid,
		"fingerprint":          cred.Fingerprint,
		"private_key":          cred.PrivateKey,
		"private_key_password": cred.PrivateKeyPassword,
		"config_file_profile":  cred.ConfigFileProfile,
	}

	update := bson.M{"$set": setFields}

	result := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID, "provider": "oci"}, update)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("OCI credential with ID '%s' not found", id)
	}
	if result.Err() != nil {
		return nil, fmt.Errorf("failed to update OCI credential: %w", result.Err())
	}

	doc, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("credential not found after update")
	}

	return convertToOciCredential(doc)
}

func convertToOciCredential(doc bson.M) (*models.OciCredential, error) {
	id, ok := doc["_id"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid _id field")
	}

	var createdAt, updatedAt time.Time
	if dt, ok := doc["created_at"].(primitive.DateTime); ok {
		createdAt = dt.Time()
	}
	if dt, ok := doc["updated_at"].(primitive.DateTime); ok {
		updatedAt = dt.Time()
	}

	cred := &models.OciCredential{
		ID:        id,
		Name:      doc["name"].(string),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if v, ok := doc["auth_method"].(string); ok {
		cred.AuthMethod = v
	}
	if v, ok := doc["region"].(string); ok {
		cred.Region = v
	}
	if v, ok := doc["tenancy_ocid"].(string); ok {
		cred.TenancyOcid = v
	}
	if v, ok := doc["user_ocid"].(string); ok {
		cred.UserOcid = v
	}
	if v, ok := doc["fingerprint"].(string); ok {
		cred.Fingerprint = v
	}
	if v, ok := doc["private_key"].(string); ok {
		cred.PrivateKey = v
	}
	if v, ok := doc["private_key_password"].(string); ok {
		cred.PrivateKeyPassword = v
	}
	if v, ok := doc["config_file_profile"].(string); ok {
		cred.ConfigFileProfile = v
	}

	return cred, nil
}

// setOptionalString sets a key in the document only if the value is non-empty.
func setOptionalString(doc bson.M, key, value string) {
	if value != "" {
		doc[key] = value
	}
}
