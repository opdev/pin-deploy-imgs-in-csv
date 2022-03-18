package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	parser "github.com/novln/docker-parser"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"sigs.k8s.io/yaml"
)

var (
	usage = `This tool will check your ClusterServiceVersion's
deployment containers for images referenced using a tag, and replaces
the tag with the digest of the image at that point in time`
	execution = fmt.Sprintf("%s /path/to/clusterserviceversion.yaml", os.Args[0])
	version   = "unknown" // replace with -ldflags at build time.
)

func Run() int {
	// Expect only a single positional arg - the csv file path
	if len(os.Args) != 2 {
		fmt.Fprintf(
			os.Stderr,
			"ERR accepts only a single positional arg: the path"+
				"to the CSV file to modify. Received %d",
			len(os.Args)-1)
		return 10
	}

	// Handle requests for help or the version
	switch strings.ToLower(os.Args[1]) {
	case "help":
		printUsage()
		return 0
	case "version":
		fmt.Println(version)
		return 0
	}

	// Read in CSV file
	csvFile := os.Args[1]
	bts, err := os.ReadFile(csvFile)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Unable to read the file on disk at path %s with error\n%s",
			csvFile,
			err,
		)
		return 20
	}

	var csv operatorsv1alpha1.ClusterServiceVersion

	// Decode data as  yaml
	err = yaml.Unmarshal(bts, &csv)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"The CSV file provided %s did not cleanly marshal to a CSV struct with error\n%s",
			csvFile,
			err,
		)
		return 30
	}

	// Scan containers in deployments for image values,
	// and replace them with digests. Store a mapping id imageWithDigest: tag
	// in replacedImageStrings to add the tags as comments in post processing.
	replacedImageStrings := map[string]string{}
	for di, deploymentSpec := range csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
		for ci, container := range deploymentSpec.Spec.Template.Spec.Containers {
			image := container.Image
			ref, err := parser.Parse(image)
			if err != nil {
				fmt.Fprintf(
					os.Stderr,
					"Unable to parse the image reference %s with the following error\n%s",
					image,
					err,
				)
				return 40
			}
			isTag := true
			tagOrDigest := ref.Tag() // returns the tag or digest value

			// check to see if it's a digest by matching the first 6 characters)
			if len(tagOrDigest) > 6 && tagOrDigest[0:6] == "sha256" {
				isTag = false
			}

			if isTag {
				// get the image digest and replace the image reference that
				// uses a tag.
				desc, err := crane.Head(image)
				if err != nil {
					fmt.Fprintf(
						os.Stderr,
						"Unable to query the registry for image %s with the following error\n%s",
						image,
						err,
					)
					return 50
				}

				imageBreakdown := strings.Split(image, ":")
				if len(imageBreakdown) == 1 {
					// NOTE(komish):
					//
					// The image did not contain a colon followed
					// by a tag. Image semantics indicates that this
					// means the user wanted the "latest" tag.
					//
					// We expect the tag to be in the 1 index of
					// imageBreakdown so we can reliably replace it with
					// the digest. If there's no colon, we won't have
					// a value in the 1 index.
					//
					// In this case, we'll add "latest" to the
					// breakdown so we can then replace it with
					// the digest.
					imageBreakdown = append(imageBreakdown, "latest")
				}

				imageBreakdown[1] = desc.Digest.String()

				newImageString := imageBreakdown[0] + "@" + desc.Digest.String()
				// set the container.Image value to the digest.
				updateContainerImage(&csv, di, ci, newImageString)
				replacedImageStrings[newImageString] = ref.Tag()
			}
		}
	}

	newCSV, err := yaml.Marshal(csv)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Unable to render the ClusterServiceVersion struct to bytes."+
				"This is an internal error. File a bug at on GitHub for this project"+
				"https://github.com/opdev/pin-deploy-imgs-in-csv"+
				"The error was:\n%s",
			err,
		)
		return 60
	}

	// Post-process the rendered CSV after serialization.
	// The tasks here cannot be done while working with the CSV
	// struct.

	// Strip the status block. This is rendered with empty fields otherwise.
	statusBytes := []byte("status:")
	newCSV = bytes.Split(newCSV, statusBytes)[0]

	// Add comments to image lines with the tag that was replaced.
	for imgStr, tag := range replacedImageStrings {
		imgEntryAsBytes := []byte("image: " + imgStr)
		tagCommentAsBytes := []byte(" # " + tag)

		insertionPoint := bytes.Index(newCSV, imgEntryAsBytes) + len(imgEntryAsBytes)
		newCSV = bytes.Join([][]byte{newCSV[:insertionPoint], tagCommentAsBytes, newCSV[insertionPoint:]}, nil)
	}

	// write the file in-place
	err = os.WriteFile(csvFile, newCSV, 0644)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Unable to overwrite the file at path %s with error\n%s",
			csvFile,
			err,
		)

		// TODO(): should we consider sending the newCSV to stdout?
		return 70
	}

	return 0
}

// updateContainerImage will update the deployment's container image value with newValue.
func updateContainerImage(csv *operatorsv1alpha1.ClusterServiceVersion, deploymentIndex, containerIndex int, newValue string) {
	csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[deploymentIndex].Spec.Template.Spec.Containers[containerIndex].Image = newValue
}

func printUsage() {
	fmt.Printf(`%s

%s`, execution, usage,
	)
}
