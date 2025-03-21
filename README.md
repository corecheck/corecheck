<h1 align="center">
  <br>
  <a href="https://corecheck.dev"><img src="https://github.com/bitcoin-coverage/core/raw/master/docs/assets/logo.png" alt="Bitcoin Coverage" width="200"></a>
  <br>
    Bitcoin Coverage's Infra as Code
  <br>
</h1>

<h4 align="center">Bitcoin Coverage's infrastructure as code</h4>

## 📖 Introduction
This repository contains the infrastructure as code for Bitcoin Coverage. It is used to deploy all the components of the project.

## 🚀 CI/CD
The CI/CD is handled by GitHub Actions and is located in the `.github/workflows` folder. It is used to deploy the infrastructure on AWS on every push to the `master` branch.

## 🤝 Contributing
Contributions are welcome! To set up a local working environment, provision a copy of infrastructure to your own AWS account using a Terraform "namespace" with the following steps.

Ensure your AWS environment credentials are properly configured. Provision the S3 buckets for the remote state bucket and buckets containing the Lambda function artifacts:
```
cd deploy/terraform/remote-state
terraform init
terraform workspace new [developer-namespace]
terraform deploy
```

Initialize your local copy to use your namespaced remote state bucket.
```
cd deploy/terraform
terraform init -backend-config="bucket=bitcoin-coverage-state-[developer-namespace]"
terraform workspace new [developer-namespace]
```

Build the Lambda artifacts (requires `docker` installed on the local machine).
```
make build-lambdas

export CORECHECK_S3_API_BUCKET=corecheck-api-lambdas-[developer-namespace]
export CORECHECK_S3_COMPUTE_BUCKET=corecheck-compute-lambdas-[developer-namespace]
make deploy-lambdas
```

Deploy the infrastructure with `terraform apply`

## 📝 License

MIT - [Aurèle Oulès](https://github.com/aureleoules)
