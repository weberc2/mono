package main

import (
	"log/slog"
	"os"

	v1 "iamcontroller/pkg/api/v1"
	"iamcontroller/pkg/controller"
	"iamcontroller/pkg/log"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func main() {
	var level slog.Level
	level.UnmarshalText([]byte(os.Getenv("LOG_LEVEL")))
	logger := slog.New(
		slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level}),
	)
	slog.SetDefault(logger)
	logger.Debug("starting...")

	ctrl.SetLogger(logr.FromSlogHandler(logger.Handler()))
	ctx := ctrl.SetupSignalHandler()

	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		fatal(logger, "adding scheme", "err", err.Error())
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		fatal(logger, "adding corev1 scheme", "err", err.Error())
	}

	manager, err := ctrl.NewManager(
		ctrl.GetConfigOrDie(),
		manager.Options{Scheme: scheme},
	)
	if err != nil {
		fatal(logger, "creating controller manager", "err", err.Error())
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		fatal(logger, "missing required environment variable: AWS_REGION")
	}

	awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		fatal(logger, "loading aws sdk config", "err", err.Error())
	}

	userReconciler := &controller.UserReconciler{
		Client: manager.GetClient(),
		Users:  controller.UserClient{IAM: iam.NewFromConfig(awsConfig)},
	}
	accessKeyReconciler := controller.AccessKeyReconciler{
		Client: userReconciler.Client,
		Users:  userReconciler.Users,
	}
	if err := userReconciler.Configure(manager); err != nil {
		fatal(logger, "configuring user reconciler", "err", err.Error())
	}
	if err := accessKeyReconciler.Configure(manager); err != nil {
		fatal(logger, "configuring access key reconciler", "err", err.Error())
	}

	if err = manager.Start(log.Context(ctx, logger)); err != nil {
		fatal(logger, "starting controller manager", "err", err.Error())
	}

	logger.Info("exiting")
}

func fatal(logger *slog.Logger, msg string, vs ...any) {
	logger.Error(msg, vs...)
	os.Exit(1)
}
