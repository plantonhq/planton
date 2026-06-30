package module

import (
	"fmt"

	pulumigcp "github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/artifactregistry"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// repo creates a repository and configures IAM bindings based on public/private access
func repo(ctx *pulumi.Context, locals *Locals, gcpProvider *pulumigcp.Provider, serviceAccounts *ServiceAccounts) error {
	//create a variable with descriptive name for the api-resource in the input
	gcpArtifactRegistryRepo := locals.GcpArtifactRegistryRepo

	//todo: might be better to add a random suffix to the repo name to avoid conflicts
	repoName := gcpArtifactRegistryRepo.Metadata.Name

	// Get project ID from StringValueOrRef (currently only supports literal value)
	// TODO: Implement reference resolution in a shared library
	projectId := gcpArtifactRegistryRepo.Spec.ProjectId.GetValue()

	//create repository
	createdRepo, err := artifactregistry.NewRepository(ctx,
		repoName,
		&artifactregistry.RepositoryArgs{
			Project:      pulumi.String(projectId),
			Location:     pulumi.String(gcpArtifactRegistryRepo.Spec.Region),
			RepositoryId: pulumi.String(gcpArtifactRegistryRepo.Metadata.Name),
			Format:       pulumi.String(gcpArtifactRegistryRepo.Spec.RepoFormat.String()),
			Labels:       pulumi.ToStringMap(locals.GcpLabels),
		}, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create repo")
	}

	if gcpArtifactRegistryRepo.Spec.EnablePublicAccess {
		//grant "reader" role for all users on the repo to make it public
		_, err = artifactregistry.NewRepositoryIamMember(ctx,
			fmt.Sprintf("%s-public-reader", repoName),
			&artifactregistry.RepositoryIamMemberArgs{
				Project:    pulumi.String(projectId),
				Location:   pulumi.String(gcpArtifactRegistryRepo.Spec.Region),
				Repository: createdRepo.RepositoryId,
				Role:       pulumi.String("roles/artifactregistry.reader"),
				//"allUsers" is a special identifier on google identity system which is used
				//for granting permissions to for everyone.
				Member: pulumi.Sprintf("allUsers"),
			}, pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdRepo}))
		if err != nil {
			return errors.Wrap(err, "failed to grant reader role on repo for all users")
		}
	} else {
		//grant "reader" role to the reader service account for private repos
		_, err = artifactregistry.NewRepositoryIamMember(ctx,
			fmt.Sprintf("%s-reader-sa", repoName),
			&artifactregistry.RepositoryIamMemberArgs{
				Project:    pulumi.String(projectId),
				Location:   pulumi.String(gcpArtifactRegistryRepo.Spec.Region),
				Repository: createdRepo.RepositoryId,
				Role:       pulumi.String("roles/artifactregistry.reader"),
				Member:     pulumi.Sprintf("serviceAccount:%s", serviceAccounts.ReaderAccount.Email),
			}, pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdRepo, serviceAccounts.ReaderAccount}))
		if err != nil {
			return errors.Wrap(err, "failed to grant reader role on repo for reader service account")
		}
	}

	//grant "writer" role to the writer service account (always, regardless of public/private)
	_, err = artifactregistry.NewRepositoryIamMember(ctx,
		fmt.Sprintf("%s-writer-sa", repoName),
		&artifactregistry.RepositoryIamMemberArgs{
			Project:    pulumi.String(projectId),
			Location:   pulumi.String(gcpArtifactRegistryRepo.Spec.Region),
			Repository: createdRepo.RepositoryId,
			Role:       pulumi.String("roles/artifactregistry.writer"),
			Member:     pulumi.Sprintf("serviceAccount:%s", serviceAccounts.WriterAccount.Email),
		}, pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdRepo, serviceAccounts.WriterAccount}))
	if err != nil {
		return errors.Wrap(err, "failed to grant writer role on repo for writer service account")
	}

	//grant "repoAdmin" role to the writer service account (for full repository management)
	_, err = artifactregistry.NewRepositoryIamMember(ctx,
		fmt.Sprintf("%s-admin-sa", repoName),
		&artifactregistry.RepositoryIamMemberArgs{
			Project:    pulumi.String(projectId),
			Location:   pulumi.String(gcpArtifactRegistryRepo.Spec.Region),
			Repository: createdRepo.RepositoryId,
			Role:       pulumi.String("roles/artifactregistry.repoAdmin"),
			Member:     pulumi.Sprintf("serviceAccount:%s", serviceAccounts.WriterAccount.Email),
		}, pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdRepo, serviceAccounts.WriterAccount}))
	if err != nil {
		return errors.Wrap(err, "failed to grant repoAdmin role on repo for writer service account")
	}

	//export important attributes of the repository as outputs
	ctx.Export(OpRepoName, createdRepo.RepositoryId)
	//export docker-repo hostname
	ctx.Export(OpRepoHostname, pulumi.Sprintf(
		"%s-docker.pkg.dev", createdRepo.Location))
	//export complete repo url based on the attributes of the created docker-repo
	// ex: artifactregistry://us-central1-docker.pkg.dev/my-gcp-project-id/my-company-docker-repo
	ctx.Export(OpRepoUrl, pulumi.Sprintf(
		"%s-docker.pkg.dev/%s/%s",
		createdRepo.Location, createdRepo.Project, createdRepo.Name))

	return nil
}
