# Release process in this repository

The release process in this repository differs from the standard release process for other repositories in the 
GitHub organization since this repository is used to publish Terraform providers to the Terraform Registry.
The release process is automated using GitHub Actions and is triggered by pushing a new tag to the repository.
The release process includes the following steps:

1. Ensure that you have pulled the latest changes from the main branch and that your local branch is up to date.
 
2. Create a new tag for the release using the following command:
```bash
git tag -a vX.Y.Z -m "Release version X.Y.Z"
```
Replace `X.Y.Z` with the appropriate version number for the release. You should check the existing tags and review the
merged pull requests to determine the appropriate version number for the new release
(see checkboxes in the pull request template for guidance on versioning).
For the patch and minor releases you should bump the corresponding version number
(e.g. from v1.2.3 to v1.2.4 for a patch release or from v1.2.3 to v1.3.0 for a minor release).

If any of the merged pull requests contain breaking changes, please ensure that you have followed the
[breaking change process](https://developer.hashicorp.com/terraform/plugin/sdkv2/best-practices/deprecations) 

3. Push the new tag to the remote repository using the following command:
```bash
git push origin vX.Y.Z
```

4. The GitHub Actions workflow will be triggered by the new tag and will automatically build and publish the Terraform
   provider to the Terraform Registry.
 
5. Once the release is published, you can verify that the new version is available on the Terraform Registry and
   update any relevant documentation or release notes as needed. Please not that there is a delay between the time
   the release is published and when it becomes available on the Terraform Registry.