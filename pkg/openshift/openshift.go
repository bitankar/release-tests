package openshift

import (
	"fmt"
	"log"

	"github.com/getgauge-contrib/gauge-go/testsuit"
	"github.com/openshift-pipelines/release-tests/pkg/clients"
	"github.com/openshift-pipelines/release-tests/pkg/config"
	imageStream "github.com/openshift/client-go/image/clientset/versioned"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func GetImageStreamTags(c *clients.Clients, namespace, name string) []string {
	fmt.Printf("Getting imagestream from the namespace %s", namespace)
	is := imageStream.NewForConfigOrDie(c.KubeConfig)
	isRequired, err := is.ImageV1().ImageStreams(namespace).Get(c.Ctx, name, v1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}
	tags := isRequired.Spec.Tags
	var tagNames []string
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}
	return tagNames
}

func VerifyImageStreamExists(c *clients.Clients, name, namespace string) {
	log.Printf("Verify that image stream %q exists in namespace %q", name, namespace)
	is := imageStream.NewForConfigOrDie(c.KubeConfig)

	if err := wait.PollImmediate(config.APIRetry, config.APITimeout, func() (bool, error) {
		_, err := is.ImageV1().ImageStreams(namespace).Get(c.Ctx, name, v1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			return false, nil
		}
		return true, err
	}); err != nil {
		testsuit.T.Errorf("Failed to get image stream %q in namespace %q for tests: %s", name, namespace, err)
	}
}
