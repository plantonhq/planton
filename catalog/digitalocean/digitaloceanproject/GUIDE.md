# DigitalOcean Project -- Operational Guide

What experience with this component teaches that the field reference cannot.

## Destroy relocates -- plan for where things land

Deleting a project never deletes its contents: DigitalOcean requires the project to be empty, so everything inside is moved to the account's DEFAULT project first, then the empty project is deleted (with retries while the asynchronous moves settle). If you tear down a project as part of an environment teardown, destroy the member resources first -- otherwise they keep running, and billing, from the default project.

## The relocation is slow sometimes -- the destroy waits ten minutes for it

DigitalOcean accepts the "move these members to the default project" request at once but applies it asynchronously, and the wait is not predictable: measured live, a one-member project read empty within seconds on one destroy, within about thirty seconds on another, and NOT within three minutes on a third. The provider retries the delete through `412 cannot delete a project with resources` only for its delete timeout, whose default is three minutes -- so that third destroy failed although the members had already been moved. Both provisioners therefore give the delete ten minutes (the tight default and the undocumented `timeouts` knob are reported upstream as [digitalocean/terraform-provider-digitalocean#1608](https://github.com/digitalocean/terraform-provider-digitalocean/issues/1608)). If a destroy still reports that error, nothing is wrong with the members: they are in the default project. Confirm the project reads empty (`GET /v2/projects/{id}/resources`) and run the destroy again; the empty project deletes in seconds. A project is a free object, so the retry costs nothing.

## One project per resource

A resource belongs to exactly one project. Declaring it in this project's `resources` list MOVES it -- including out of another project that also claims it. Two projects listing the same resource will fight forever. Give each resource one home, and prefer wiring membership by reference (`valueFrom` on the producing kind's `urn` output) so the graph is visible in code.

## Leave membership unmanaged when another system owns it

An empty `resources` list means "do not manage membership": resources assigned from the console or by their own project fields are left alone. Use this when teams assign resources to projects manually and the manifest should only own the container.

## The purpose field round-trips -- with one trap

Any free-text purpose works: DigitalOcean stores it as `Other: <your text>` and the provider strips the prefix when reading back, so manifests converge. The one value that can never converge is text that itself starts with `Other:` -- DigitalOcean keeps exactly one prefix and re-capitalizes the rest (`Other: probe` is stored as `Other: Probe`), the provider then strips the prefix, and the read-back (`Probe`) can never equal what you wrote -- so validation rejects it up front.

## is_default: almost always leave it alone

DigitalOcean documents that a Terraform-managed project should not be made the account's default, and its own acceptance test expects a never-settling plan when it is. The account can have only ONE default, so out-of-band changes show up here as drift, and a default project refuses deletion. Set it only if the account's default is genuinely meant to be managed as code, and expect the quirks.

## Environment casing

Declare `environment` lowercase (`production`); DigitalOcean reports it back capitalized (`Production`). The provisioners absorb the difference -- do not "fix" it by writing the capitalized form, which validation rejects to keep one canonical spelling.

## What is deliberately NOT here

DigitalOcean's standalone partial-ownership membership resource (`digitalocean_project_resources`) -- it conflicts by design with the project's own whole-list membership read-back; team member permissions (account-level, not a project property); and any per-member configuration, which belongs to each member's own kind.
