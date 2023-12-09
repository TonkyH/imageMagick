package imagemagick

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"cloud.google.com/go/storage"
	vision "cloud.google.com/go/vision/apiv1"
	"cloud.google.com/go/vision/v2/apiv1/visionpb"
)

var (
	storageClient *storage.Client
	visionClient  *vision.ImageAnnotatorClient
)

func init() {
	var err error

	storageClient, err = storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("storage.NewClient: %v", err)
	}

	visionClient, err = vision.NewImageAnnotatorClient(context.Background())
	if err != nil {
		log.Fatalf("vision.NewAnnotatorClient: %v", err)
	}
}

type GCSEvent struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

func BlurOffensiveImages(ctx context.Context, e GCSEvent) error {
	outputBucket := os.Getenv("BLURRED_BUCKET_NAME")
	if outputBucket == "" {
		return errors.New("BLURRED_BUCKET_NAME must be set")
	}

	img := vision.NewImageFromURI(fmt.Sprintf("gs://%s/%s", e.Bucket, e.Name))

	resp, err := visionClient.DetectSafeSearch(ctx, img, nil)
	if err != nil {
		return fmt.Errorf("AnnotateImage: %w", err)
	}

	if resp.GetAdult() == visionpb.Likelihood_VERY_LIKELY ||
		resp.GetViolence() == visionpb.Likelihood_VERY_LIKELY {
		return blur(ctx, e.Bucket, outputBucket, e.Name)
	}
	log.Printf("The image %q was detected as OK.", e.Name)
	return nil
}

func blur(ctx context.Context, inputBucket, outputBucket, name string) error {
	inputBlob := storageClient.Bucket(inputBucket).Object(name)
	r, err := inputBlob.NewReader(ctx)
	if err != nil {
		return fmt.Errorf("NewReader: %w", err)
	}

	outputBlob := storageClient.Bucket(outputBucket).Object(name)
	w := outputBlob.NewWriter(ctx)
	defer w.Close()

	// Use - as input and output to use stdin and stdout.
	cmd := exec.Command("convert", "-", "-blur", "0x8", "-")
	cmd.Stdin = r
	cmd.Stdout = w

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cmd.Run: %w", err)
	}

	log.Printf("Blurred image uploaded to gs://%s/%s", outputBlob.BucketName(), outputBlob.ObjectName())

	return nil
}
