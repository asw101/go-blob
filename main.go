package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {

	client, err := containerClientFromEnv()
	if err != nil {
		return err
	}

	// container
	ctx := context.Background()
	_, err = client.Create(ctx, nil)
	var storageErr *azblob.StorageError
	if err != nil && errors.As(err, &storageErr) {
		if storageErr.ErrorCode == azblob.StorageErrorCodeContainerAlreadyExists {
			log.Printf("container already exists")
		} else {
			return err
		}
	}

	// upload
	blobName := "hello.txt"
	data := []byte("hello, world")
	err = uploadBlob(client, blobName, data)
	if err != nil {
		return err
	}

	// list
	err = listDownloadAndDeleteBlobs(client)
	if err != nil {
		return err
	}

	return nil
}

func listDownloadAndDeleteBlobs(client *azblob.ContainerClient) error {

	ctx := context.Background()
	pager := client.ListBlobsFlat(nil)
	for pager.NextPage(ctx) {
		resp := pager.PageResponse()
		for _, v := range resp.ListBlobsFlatSegmentResponse.Segment.BlobItems {
			// remove properties to make output less verbose
			v.Properties = nil
			b, err := json.Marshal(v)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", b)
			downloadBlob(client, *v.Name)
			deleteBlob(client, *v.Name)
		}
	}
	if err := pager.Err(); err != nil {
		return err
	}
	return nil
}

func uploadBlob(client *azblob.ContainerClient, blobName string, data []byte) error {

	blobClient, err := client.NewBlockBlobClient(blobName)
	if err != nil {
		return err
	}

	ctx := context.Background()
	option := azblob.UploadOption{}
	option.HTTPHeaders = &azblob.BlobHTTPHeaders{
		BlobContentType: to.Ptr("text/plain; charset=utf-8"),
	}
	_, err = blobClient.UploadBuffer(ctx, data, option)
	if err != nil {
		return err
	}

	return nil
}

func downloadBlob(client *azblob.ContainerClient, blobName string) error {

	blobClient, err := client.NewBlockBlobClient(blobName)
	if err != nil {
		return err
	}

	ctx := context.Background()
	resp, err := blobClient.Download(ctx, nil)
	if err != nil {
		return err
	}

	reader := resp.Body(nil)
	defer reader.Close()
	b, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	map1 := map[string]interface{}{
		"Name": blobName,
		"Body": string(b),
	}
	b, err = json.Marshal(map1)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", b)

	return nil
}

func deleteBlob(client *azblob.ContainerClient, blobName string) error {

	blobClient, err := client.NewBlockBlobClient(blobName)
	if err != nil {
		return err
	}

	ctx := context.Background()
	_, err = blobClient.Delete(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}

func containerClientFromEnv() (*azblob.ContainerClient, error) {
	storageAccountName := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	if storageAccountName == "" {
		return nil, errors.New("AZURE_STORAGE_ACCOUNT_NAME not set")
	}

	storageAccountKey := os.Getenv("AZURE_STORAGE_PRIMARY_ACCOUNT_KEY")

	containerName := os.Getenv("AZURE_STORAGE_CONTAINER_NAME")
	if containerName == "" {
		containerName = "mycontainer"
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", storageAccountName)

	getServiceClient := func() (*azblob.ServiceClient, error) {
		if storageAccountKey != "" {

			cred, err := azblob.NewSharedKeyCredential(storageAccountName, storageAccountKey)
			if err != nil {
				return nil, err
			}

			serviceClient, err := azblob.NewServiceClientWithSharedKey(serviceURL, cred, nil)
			if err != nil {
				return nil, err
			}

			return serviceClient, nil
		}

		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, err
		}

		serviceClient, err := azblob.NewServiceClient(serviceURL, cred, nil)
		if err != nil {
			return nil, err
		}

		return serviceClient, nil
	}

	serviceClient, err := getServiceClient()
	if err != nil {
		return nil, err
	}

	containerClient, err := serviceClient.NewContainerClient(containerName)
	if err != nil {
		return nil, err
	}

	return containerClient, nil
}
