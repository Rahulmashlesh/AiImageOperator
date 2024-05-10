package controller

import (
	aiexamplecomv1 "AiImageOperator/api/v1"
	config2 "AiImageOperator/internal/config"
	"AiImageOperator/internal/image"
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-logr/logr"
	"github.com/go-redis/redis/v8"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"math"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"
	"strings"
)

// ImageReconciler reconciles a Image object
type ImageReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	//Cache  image.Cache
	Generator image.Generator
}

//+kubebuilder:rbac:groups=ai.example.com,resources=images,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ai.example.com,resources=images/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ai.example.com,resources=images/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Image object against the actual cluster state, an d then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *ImageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := log.FromContext(ctx)
	logger.Info("New reconcile loop")
	img := &aiexamplecomv1.Image{}
	const imageFinalizer = "finalizer.images.ai.example.com"

	if err := r.Get(ctx, req.NamespacedName, img); err != nil {
		if errors.IsNotFound(err) {

			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Check if the Image resource is marked for deletion
	if img.DeletionTimestamp != nil {
		if containsString(img.ObjectMeta.Finalizers, imageFinalizer) {
			// Run finalization logic for imageFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeImage(ctx, logger, img); err != nil {
				return ctrl.Result{}, err
			}

			// Remove imageFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			img.ObjectMeta.Finalizers = removeString(img.ObjectMeta.Finalizers, imageFinalizer)
			err := r.Update(ctx, img)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR

	if !controllerutil.ContainsFinalizer(img, imageFinalizer) {
		controllerutil.AddFinalizer(img, imageFinalizer)
		if err := r.Update(ctx, img); err != nil {
			return ctrl.Result{}, err
		}
	}

	input := image.Input{
		Prompt: img.Spec.Prompt,
		Seed:   img.Status.Seed,
	}
	output := image.Output{}
	var err error

	if img.Status.Seed == "" {

		// It's a new prompt,
		// generate the seed and make the API call to get the image.
		input.Seed = strconv.FormatInt(rand.Int63nRange(0, math.MaxUint32), 10)

		logger.Info("New Image CR", "seed: ", input.Seed, "prompt: ", input.Prompt)

		output, err = r.Generator.Generate(ctx, input, logger)
		if err != nil {
			logger.Error(err, "Error Generating Image", nil)
			return ctrl.Result{}, err
		}

		img.Status.Seed = output.Seed
		if err := r.Status().Update(ctx, img); err != nil {
			return ctrl.Result{}, err
		}

		// TODO: component, make an interface called storage and pass
		logger.Info("Storing the Generated Image", "Image will be saved in: ", config2.AppConfig.SaveTo)
		switch strings.ToLower(config2.AppConfig.SaveTo) {
		case "s3":
			err := UploadToS3(ctx, output.Data, config2.AppConfig.S3BucketName, output.Seed+"/"+input.Prompt, logger)
			if err != nil {
				return ctrl.Result{}, err
			}
		case "redis":
			err := writeToRedis(ctx, output)
			if err != nil {
				return ctrl.Result{}, err
			}
		case "volume":
			_, err := storeImageInObjectStorage(output)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	} else {
		logger.Info("Nothing to do", "Seed exists: ", input.Seed)
	}
	return ctrl.Result{}, nil
}

func writeToRedis(ctx context.Context, output image.Output) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "my-redis-master:6379", // Redis server address
		Password: "",                     // no password set
		DB:       0,                      // use default DB

	})

	// Test connection to Redis
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return err
	}
	_, err = rdb.Set(ctx, output.Seed, output.Data, 0).Result()
	if err != nil {
		return err
	}
	return nil
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	result := []string{}
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}

func (r *ImageReconciler) finalizeImage(ctx context.Context, logger logr.Logger, img *aiexamplecomv1.Image) error {
	// Execute cleanup logic, e.g., delete images from S3
	logger.Info("Deleting external resources for", "Image", img.Name)
	// Assume we use the image's UID to identify the resource in external systems
	//resourceKey := fmt.Sprintf("%s/%s", img.UID, img.Spec.Prompt)

	/*	if config2.AppConfig.SaveToS3 == "true" {
		if err := deleteFromS3(ctx, config2.AppConfig.S3BucketName, resourceKey, logger); err != nil {
			return err
		}
	}*/
	return nil
}

func deleteFromS3(ctx context.Context, bucket, objectKey string, logger logr.Logger) error {
	// TODO delete from S3

	return nil
}

// TODO crete an interface with just the methods being used ie putObject() helps testing
func UploadToS3(ctx context.Context, data []byte, bucket, objectKey string, logger logr.Logger) error {
	// TODO read config from ./aws/config
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-2"),
		config.WithCredentialsProvider(aws.NewCredentialsCache(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     config2.AppConfig.AccessKeyID,
				SecretAccessKey: config2.AppConfig.SecretAccessKey,
				SessionToken:    config2.AppConfig.SessionToken, // Optional,
			}, nil
		})),
		))
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	reader := bytes.NewReader(data)
	logger.Info("Writing to S3")
	// Upload the data to S3
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:          aws.String(bucket),
		Key:             aws.String(objectKey),
		Body:            reader,
		ContentEncoding: aws.String("png"),
	})

	if err != nil {
		logger.Info("tracker 6")
		logger.Error(err, "unable to upload to s3", 0)
	}

	logger.Info("Successfully uploaded to bucket", "bucket", bucket, "objectKey", objectKey)
	return nil
}

func storeImageInObjectStorage(output image.Output) (string, error) {
	// Construct the file path where the image will be saved
	// "/mnt/my-storage" is the mount path of the PVC in your pod
	filePath := filepath.Join("/mnt/my-storage", output.Seed)

	// Write the image data to the file
	err := os.WriteFile(filePath, output.Data, 0644)
	if err != nil {
		return "", err
	}
	// Return the file path as a reference to where the image is stored
	return filePath, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ImageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aiexamplecomv1.Image{}).
		Complete(r)
}

/* write how we want to store the returned image. what the best way.
0. Log:
	Bad ideas first :P .... Just log it for now to see if the API call works.
1. config map in k8s
	limitations: max size is 1MB
2. PVC
	large images can be stored for a long duration. can mount a fs to the cluster
3. Cloud storage like S3
	we can use the aws sdk and push the images to S3 (once we go to prod with this app) :P
4. pocketbase
	or some other DB to store the image file.
*/

// TODO next time
//controller to manage the deployment

// put aws to secrets,

// make the PVC approach work to save the imge.

//TODO april 11

//finalizer to remove images when the CR is deleted.

// TODO
// explore geneating the seed. approach.
