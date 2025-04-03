package controllers

import (
	"context"
	"fileSystem/models"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	//"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	//"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/gin-gonic/gin"
)

func GetFile(c *gin.Context){
	c.JSON(http.StatusOK, gin.H{"message": "file!"})
}

// Function to convert multipart.File to []byte
func fileToBytes(c *gin.Context, multipartFile multipart.File) ([]byte, error) {
    data, err := io.ReadAll(multipartFile)
    handleError(c, err)
	
    return data, nil // Return the data and nil for no error
}

func uploadBlobBuffer(c *gin.Context,client *azblob.Client, containerName string, blobName string, file []byte) {
    // Upload the data to a block blob
    _, err := client.UploadBuffer(context.TODO(), containerName, blobName, file, nil)
    handleError(c, err)
}

func getServiceClientSAS(c *gin.Context, accountURL string, sasToken string) *azblob.Client {

	
    // Create a new service client with an existing SAS token

    // Append the SAS to the account URL with a "?" delimiter
    accountURLWithSAS := fmt.Sprintf("%s?%s", accountURL, sasToken)

    client, err := azblob.NewClientWithNoCredential(accountURLWithSAS, nil)
    handleError(c, err)

    return client
}

// Upload handler
func UploadHandler(c *gin.Context) {

	containerName := os.Getenv("azureBlobContainerName")
	accountURL := os.Getenv("azureBlobAccountURL")
	blobSASToken := os.Getenv("blobSASToken")

	// Get the file from the request
	multipartFile, header, err := c.Request.FormFile("file")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file"})
		return
	}

	defer multipartFile.Close()

	// Convert file to byte slice
	fileBytes, err := fileToBytes(c, multipartFile)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	fmt.Println("File size:", header.Size)
	fmt.Println("File name:", header.Filename)
	fmt.Println("File type:", header.Header.Get("Content-Type"))

	fileName := header.Filename

	client := getServiceClientSAS(c, accountURL, blobSASToken)
	uploadBlobBuffer(c, client, containerName, fileName, fileBytes)

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
	})
}

func listBlobsFlat(c *gin.Context, client *azblob.Client, containerName string) ([]*models.File, error) {
    var blobs []*models.File

    // List the blobs in the container
    pager := client.NewListBlobsFlatPager(containerName, &azblob.ListBlobsFlatOptions{
        Include: azblob.ListBlobsInclude{Snapshots: true, Versions: true},
    })

    // Iterate through the pager to get all blob items
    for pager.More() {
        resp, err := pager.NextPage(context.TODO())
        if err != nil {
            return nil, err // Return error if any
        }
		
        // Collect blob names
        for _, blob := range resp.Segment.BlobItems {

			file := &models.File{ // Declare a new instance inside the loop
				BlobName: *blob.Name,
				FileDate: blob.Properties.LastModified.String(),
			}
			blobs = append(blobs, file)
        }
    }

    return blobs, nil // Return the list of blob names
}


func ListBlobs(c *gin.Context){
	containerName := os.Getenv("azureBlobContainerName")
	accountURL := os.Getenv("azureBlobAccountURL")
	blobSASToken := os.Getenv("blobSASToken")

	client := getServiceClientSAS(c, accountURL, blobSASToken)
	

	listedBlobs, err := listBlobsFlat(c, client, containerName)
	handleError(c, err)

	c.JSON(http.StatusOK, gin.H{
		"Blobs": listedBlobs,
	})
}

func deleteBlob(c *gin.Context, client *azblob.Client, containerName string, blobName string) {
    // Delete the blob
    _, err := client.DeleteBlob(context.TODO(), containerName, blobName, nil)
    handleError(c, err)

	
}

// downloadBlob downloads an Azure blob and saves it locally
func downloadBlob(c *gin.Context, client *azblob.Client, containerName string, blobName string, downloadPath string) {
	// Ensure the download directory exists
	err := os.MkdirAll(downloadPath, os.ModePerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create download directory"})
		return
	}

	// Construct the full file path correctly
	filePath := filepath.Join(downloadPath, blobName)

	// Create or open the file
	file, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return
	}
	defer file.Close()

	// Download the blob
	_, err = client.DownloadFile(context.TODO(), containerName, blobName, file, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download blob"})
		return
	}

	fmt.Println("Downloaded file:", filePath)

	c.JSON(http.StatusOK, gin.H{"message": "Blob downloaded", "filePath": filePath})
}

// DownloadFile handles the API request to download a blob
func DownloadFile(c *gin.Context) {
	containerName := os.Getenv("azureBlobContainerName")
	accountURL := os.Getenv("azureBlobAccountURL")
	blobSASToken := os.Getenv("blobSASToken")
	homeDir, _ := os.UserHomeDir()
	downloadPath := filepath.Join(homeDir, "Downloads")

	client := getServiceClientSAS(c, accountURL, blobSASToken)

	var reqBody models.File

	// Bind incoming JSON request
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body", "error": err.Error()})
		return
	}

	// Ensure BlobName is provided
	if reqBody.BlobName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "BlobName is required"})
		return
	}

	// Download the requested blob
	downloadBlob(c, client, containerName, reqBody.BlobName, downloadPath)
}


/*
func deleteBlobWithSnapshots(c *gin.Context, client *azblob.Client, containerName string, blobName string) {
    // Delete the blob and its snapshots
    _, err := client.DeleteBlob(context.TODO(), containerName, blobName, &blob.DeleteOptions{
        DeleteSnapshots: to.Ptr(blob.DeleteSnapshotsOptionTypeInclude),
    })
	
    handleError(c, err)
}
*/

func DeleteBlob(c *gin.Context){
	
	containerName := os.Getenv("azureBlobContainerName")
	accountURL := os.Getenv("azureBlobAccountURL")
	blobSASToken := os.Getenv("blobSASToken")

	client := getServiceClientSAS(c, accountURL, blobSASToken)

	var reqBody models.File

	// Bind the incoming JSON data to the struct
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	fmt.Println(reqBody.BlobName)

	deleteBlob(c, client, containerName, reqBody.BlobName)

	// For now, we'll just return a success message
	c.JSON(http.StatusOK, gin.H{
		"message": "Blob deleted",
	})

}