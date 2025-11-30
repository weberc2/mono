package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) (err error) {
	var (
		app = application{
			spreadsheet: os.Getenv("SHEETS_SPREADSHEET_ID"),
		}
		sheetsClient *sheets.Service
		gsaFile      = os.Getenv("SHEETS_GSA_FILE")
	)

	if app.spreadsheet == "" {
		return errors.New(
			"initializing sheets client: " +
				"missing required env var: SHEETS_SPREADSHEET_ID",
		)
	}

	if gsaFile == "" {
		return errors.New(
			"initializing sheets client: " +
				"missing required env var: SHEETS_GSA_FILE",
		)
	}

	if sheetsClient, err = sheets.NewService(
		ctx,
		option.WithCredentialsFile(gsaFile),
	); err != nil {
		return fmt.Errorf("initializing sheets client: %w", err)
	}
	app.sheets = sheetsClient.Spreadsheets

	if app.clientset, err = clientsetFromEnv(); err != nil {
		return fmt.Errorf("initializing k8s client: %w", err)
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for i := true; true; i = !i {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if i {
				err = app.handleNodes(ctx)
			} else {
				err = app.handleNamespaces(ctx)
			}
			if err != nil {
				log.Printf("ERROR %v", err)
			}
		}
	}
	return
}

type application struct {
	clientset      *kubernetes.Clientset
	sheets         *sheets.SpreadsheetsService
	spreadsheet    string
	nodeState      sheetUpdateState
	namespaceState sheetUpdateState
}

func (app *application) handleNamespaces(ctx context.Context) (err error) {
	var (
		namespaces *v1.NamespaceList
		opts       metav1.ListOptions
	)
	if namespaces, err = app.clientset.CoreV1().Namespaces().List(
		ctx,
		opts,
	); err != nil {
		return fmt.Errorf("fetching namespaces: %w", err)
	}

	return updateSheet(
		app.sheets,
		ctx,
		app.spreadsheet,
		&app.namespaceState,
		&namespacesSheet,
		namespaces.Items,
	)
}

func (app *application) handleNodes(ctx context.Context) (err error) {
	var (
		nodes *v1.NodeList
		opts  metav1.ListOptions
	)
	if nodes, err = app.clientset.CoreV1().Nodes().List(ctx, opts); err != nil {
		return fmt.Errorf("listing nodes: %w", err)
	}

	return updateSheet(
		app.sheets,
		ctx,
		app.spreadsheet,
		&app.nodeState,
		&nodeSheet,
		nodes.Items,
	)
}

// clientsetFromEnv creates and returns a new Kubernetes clientset.
// It tries to use an in-cluster configuration first. If that fails,
// it falls back to a local kubeconfig file.
func clientsetFromEnv() (*kubernetes.Clientset, error) {
	// Attempt to create an in-cluster config
	config, err := rest.InClusterConfig()
	if err == nil {
		// In-cluster config successful, return clientset
		fmt.Println("Using in-cluster configuration")
		return kubernetes.NewForConfig(config)
	}

	// In-cluster config failed, so assume out-of-cluster
	// Use the default kubeconfig file location
	var kubeconfigPath string
	if home := homedir.HomeDir(); home != "" {
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	// Use BuildConfigFromFlags to create a config from the kubeconfig file
	// The first parameter is an optional context, which we leave empty
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	fmt.Println("Using local kubeconfig file:", kubeconfigPath)
	return kubernetes.NewForConfig(config)
}

const (
	sheetsTimeFormat = "2006-01-02 15:04:05"

	sheetNodes = "_Nodes"
)
