# go-blob

This sample uses the `AZURE_STORAGE_ACCOUNT_NAME` environment variable with `azidentity.NewDefaultAzureCredential`.

Run locally with a pre-existing Azure Blob Storage account.

```bash
RESOURCE_GROUP='220600-keda'
# export AZURE_STORAGE_CONTAINER_NAME='mycontainer' (optional)
export AZURE_STORAGE_ACCOUNT_NAME="$(az storage account list -g $RESOURCE_GROUP -o tsv --query '[0].name')"

go run .
```

If you are testing locally you will need to ensure that the user signed in to the Azure CLI has the `Storage Blob Data Contributor` role to manage data within the Storage Account. You may need to wait a few minutes after running this commant for the permissions to propagate.

```bash
RESOURCE_GROUP='220600-keda'
AD_OBJECT_ID="$(az ad signed-in-user show --out tsv --query objectId)"
STORAGE_ACCOUNT_ID="$(az storage account list -g $RESOURCE_GROUP -o tsv --query '[0].id')"

az role assignment create \
    --role "Storage Blob Data Contributor" \
    --assignee "$AD_OBJECT_ID" \
    --scope "$STORAGE_ACCOUNT_ID"
```

Run docker image on your local machine with an account key provided by `az storage account keys list` and [azblob.NewSharedKeyCredential](https://pkg.go.dev/github.com/Azure/azure-storage-blob-go/azblob#hdr-Credentials). This is because the Azure CLI is not available within the container. Alternatively, you can also use an Azure Service Principal and [EnvironmentCredential](https://docs.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication?tabs=bash#-option-1-define-environment-variables). 

```bash
RESOURCE_GROUP='220600-keda'
export AZURE_STORAGE_ACCOUNT_NAME="$(az storage account list \
    -g $RESOURCE_GROUP 
    --out tsv 
    --query '[0].name')"

export AZURE_STORAGE_PRIMARY_ACCOUNT_KEY=$(az storage account keys list \
    --account-name "$AZURE_STORAGE_ACCOUNT_NAME" \
    --out tsv \
    --query '[0].value')

docker run --rm \
    --env AZURE_STORAGE_ACCOUNT_NAME \
    --env AZURE_STORAGE_PRIMARY_ACCOUNT_KEY \
    -it ghcr.io/asw101/go-blob:latest
```
