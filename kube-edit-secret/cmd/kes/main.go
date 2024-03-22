package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/urfave/cli/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/yaml"
)

func main() {
	namespaceFlag := cli.StringFlag{
		Name:    "namespace",
		Aliases: []string{"n"},
		Usage:   "the kubernetes namespace containing the secret",
	}

	var name string

	app := cli.Command{
		Name:  "kes",
		Usage: "edit a k8s secret without base64",
		Flags: []cli.Flag{&namespaceFlag},
		Arguments: []cli.Argument{&cli.StringArg{
			Name:        "SECRET",
			UsageText:   "the name of the kubernetes secret",
			Destination: &name,
			Min:         1,
			Max:         1,
		}},
		Action: func(ctx context.Context, c *cli.Command) error {
			return EditSecret(ctx, c.String(namespaceFlag.Name), name)
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func secretEqual(old, new *corev1.Secret) bool {
	// look for labels that have been removed or changed
	for label, value := range old.Labels {
		changeValue, found := new.Labels[label]
		if !found || value != changeValue {
			return false
		}
	}

	// look for labels that have been added
	for label := range new.Labels {
		if _, found := old.Labels[label]; !found {
			return false
		}
	}

	// look for fields that have been removed or changed
	for field, value := range old.Data {
		changeValue, found := new.Data[field]
		if !found || !bytes.Equal(value, changeValue) {
			return false
		}
	}

	// look for fields that have been added
	for field := range new.Data {
		if _, found := old.Data[field]; !found {
			return false
		}
	}

	return true
}

func loadClientset() (*kubernetes.Clientset, string, error) {
	home := homedir.HomeDir()
	if home == "" {
		return nil, "", errors.New(
			"loading kubernetes clientset: missing required environment " +
				"variable: HOME",
		)
	}

	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{
			ExplicitPath: filepath.Join(home, ".kube", "config"),
		},
		&clientcmd.ConfigOverrides{
			ClusterInfo: clientcmdapi.Cluster{Server: ""},
		},
	)

	namespace, _, err := config.Namespace()
	if err != nil {
		return nil, "", fmt.Errorf(
			"loading kubernetes clientset: fetching namespace: %w",
			err,
		)
	}

	clientConfig, err := config.ClientConfig()
	if err != nil {
		return nil, "", fmt.Errorf("loading kubernetes clientset: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, "", fmt.Errorf("loading kubernetes clientset: %w", err)
	}

	return clientset, namespace, nil
}

func marshalSecret(secret *corev1.Secret) ([]byte, error) {
	type Metadata struct {
		Namespace string            `json:"namespace"`
		Name      string            `json:"name"`
		Labels    map[string]string `json:"labels"`
	}

	stringData := make(map[string]string, len(secret.Data))
	for key, encoded := range secret.Data {
		stringData[key] = *(*string)(unsafe.Pointer(&encoded))
	}

	return yaml.Marshal(struct {
		Kind       string            `json:"kind"`
		Metadata   Metadata          `json:"metadata,omitempty"`
		StringData map[string]string `json:"stringData"`
	}{
		Kind: "Secret",
		Metadata: Metadata{
			Namespace: secret.Namespace,
			Name:      secret.Name,
			Labels:    secret.Labels,
		},
		StringData: stringData,
	})
}
