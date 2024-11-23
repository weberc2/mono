package controller

import (
	"context"
	"fmt"
	v1 "iamcontroller/pkg/api/v1"
	"iamcontroller/pkg/log"
	"log/slog"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type AccessKeyReconciler struct {
	Client         client.Client
	Users          UserClient
	ResyncInterval time.Duration
}

func (reconciler *AccessKeyReconciler) Configure(
	manager manager.Manager,
) error {
	if reconciler.ResyncInterval == 0 {
		reconciler.ResyncInterval = 5 * time.Minute
	}
	return ctrl.NewControllerManagedBy(manager).
		For(&corev1.Secret{}).
		WithEventFilter(
			predicate.NewPredicateFuncs(func(object client.Object) bool {
				if secret, ok := object.(*corev1.Secret); ok {
					for i := range secret.OwnerReferences {
						ref := secret.OwnerReferences[i]
						if ref.APIVersion == v1.SchemeGroupVersion.String() &&
							ref.Kind == kindUser {

							return true
						}
					}
				}
				return false
			})).
		Complete(reconciler)
}

func (reconciler *AccessKeyReconciler) Reconcile(
	ctx context.Context,
	r reconcile.Request,
) (res reconcile.Result, err error) {
	logger := log.FromContext(ctx)

	defer func() {
		if err == nil {
			res = reconcile.Result{RequeueAfter: reconciler.ResyncInterval}
		} else {
			err = fmt.Errorf(
				"reconciling secret `%s`: %w",
				r.NamespacedName,
				err,
			)
		}
	}()

	var secret corev1.Secret
	if err = reconciler.Client.Get(
		ctx,
		r.NamespacedName,
		&secret,
	); err != nil {
		err = fmt.Errorf("fetching secret: %w", err)
		return
	}
	if secret.StringData == nil {
		secret.StringData = map[string]string{}
	}

	accessKeyID := string(secret.Data[envAccessKeyID])
	userName := string(secret.Data[envUserName])

	if !secret.DeletionTimestamp.IsZero() {
		user := findUser(&secret)
		if user == "" {
			logger.Error(
				"owning user not found for secret--"+
					"reconciler should not be invoked for this secret"+
					"--aborting",
				"ownerReferences", secret.OwnerReferences,
			)
			return
		}

		logger.Info(
			"deleting access key",
			"user", user,
			"awsUser", userName,
			"accessKey", accessKeyID,
		)
		if err = reconciler.Users.DeleteAccessKey(
			ctx,
			userName,
			accessKeyID,
		); err != nil {
			if as[*types.NoSuchEntityException](err) == nil {
				return
			}

			// the access key has already been deleted
			err = nil
		}
		if i := slices.Index(secret.Finalizers, finalizerAccessKey); i >= 0 {
			logger.Debug("removing finalizer", "finalizer", finalizerAccessKey)
			secret.Finalizers = slices.Delete(secret.Finalizers, i, i+1)
			if err = reconciler.Client.Update(ctx, &secret); err != nil {
				err = fmt.Errorf(
					"deleting secret: "+
						"cleaning up associated access key: "+
						"removing finalizer `%s`: %w",
					finalizerAccessKey,
					err,
				)
			}
		}
		return
	}

	if !slices.Contains(secret.Finalizers, finalizerAccessKey) {
		logger.Debug(
			"finalizer not found; attaching finalizer",
			"finalizer", finalizerAccessKey,
		)
		secret.Finalizers = append(secret.Finalizers, finalizerAccessKey)
		if err = reconciler.Client.Update(ctx, &secret); err != nil {
			err = fmt.Errorf("attaching finalizer: updating secret: %w", err)
			return
		}
	}

	if userName == "" {
		logger.Warn(
			"secret is missing required field `%s`: " +
				"repairing based on owning user resource",
		)
		key := client.ObjectKey{
			Namespace: secret.Namespace,
			Name:      findUser(&secret),
		}
		if key.Name == "" {
			logger.Error(
				"owning user not found for secret--"+
					"reconciler should not be invoked for this secret"+
					"--aborting",
				"ownerReferences", secret.OwnerReferences,
			)
			return
		}

		var user v1.User
		if err = reconciler.Client.Get(ctx, key, &user); err != nil {
			if errors.IsNotFound(err) {
				logger.Error(
					"owning kubernetes resource doesn't exist! "+
						"deleting access key secret!",
					"owningUser", key.Name,
				)
				if err = reconciler.Client.Delete(ctx, &secret); err != nil {
					err = fmt.Errorf("deleting orphaned secret: %w", err)
					return
				}
			}
			err = fmt.Errorf("fetching owning user: %w", err)
			return
		}

		userName = user.UserName
		secret.StringData[envUserName] = userName
		if err = reconciler.Client.Update(ctx, &secret); err != nil {
			err = fmt.Errorf(
				"repairing secret: restoring aws user name `%s`: %w",
				userName,
				err,
			)
			return
		}
	}

	if accessKeyID == "" {
		logger.Info(
			"access key and/or user name not found in secret; " +
				"creating new access key",
		)
		goto RECREATE
	}
	if _, err = reconciler.Users.FetchAccessKey(
		ctx,
		userName,
		accessKeyID,
	); err != nil {
		if as[*UserNotFoundErr](err) != nil {
			logger.Info(
				"aws user deleted; deleting corresponding secret",
				"awsUser", userName,
			)

			// user no longer exists for the secret; delete the secret and let
			// the user controller recreate it as necessary
			if err = reconciler.Client.Delete(ctx, &secret); err != nil {
				err = fmt.Errorf(
					"deleting secret for non-existent AWS user `%s`: %w",
					userName,
					err,
				)
				return
			}

			logger.Debug(
				"deleted secret for non-existant AWS user; nothing left to do!",
			)
			return
		}

		// if the access key no longer exists we need to create a new one and
		// update this secret
		if as[*AccessKeyNotFoundErr](err) != nil {
			logger.Info("access key no longer exists; creating new one")
			goto RECREATE
		}

		// if we got here; we encountered some other error fetching the access
		// key--return so the reconciliation can be retried
		return
	}

	// everything is already in the desired state; nothing left to do.
	logger.Debug("access key already exists; nothing left to do!")
	return

RECREATE:
	slog.Debug(
		"creating new access key: cleaning up extraneous access keys for user",
		"awsUser", userName,
	)
	// delete all of the access keys for the user just to ensure any
	// orphaned keys are tidied up
	var accessKeys []types.AccessKeyMetadata
	if accessKeys, err = reconciler.Users.ListAccessKeys(
		ctx,
		userName,
	); err != nil {
		err = fmt.Errorf("deleting access keys for user: %w", err)
		return
	}

	for i := range accessKeys {
		logger.Info(
			"cleaning up unknown access key before creating new one",
			"awsUser", userName,
			"accessKey", *accessKeys[i].AccessKeyId,
		)
		if err = reconciler.Users.DeleteAccessKey(
			ctx,
			userName,
			*accessKeys[i].AccessKeyId,
		); err != nil {
			err = fmt.Errorf("deleting access keys for user: %w", err)
			return
		}
	}

	// create a new access key for the user
	logger.Info("creating new access key", "awsUser", userName)
	var accessKey *types.AccessKey
	if accessKey, err = reconciler.Users.CreateAccessKey(
		ctx,
		userName,
	); err != nil {
		return
	}

	// update the secret with the new access key information
	logger.Info(
		"updating secret for new access key",
		"awsUser", userName,
		"accessKey", *accessKey.AccessKeyId,
	)
	secret.StringData[envAccessKeyID] = *accessKey.AccessKeyId
	secret.StringData[envAccessKeySecret] = *accessKey.SecretAccessKey
	if err = reconciler.Client.Update(ctx, &secret); err != nil {
		err = fmt.Errorf(
			"creating new access key `%s`: updating secret: %w",
			*accessKey.AccessKeyId,
			err,
		)
		return
	}

	// we've put everything into the desired state; nothing left to do.
	logger.Debug("access key created and secret updated; nothing left to do!")
	return
}

func findUser(secret *corev1.Secret) string {
	for i := range secret.OwnerReferences {
		if secret.OwnerReferences[i].Kind == kindUser {
			return secret.OwnerReferences[i].Name
		}
	}
	return ""
}

const (
	finalizerAccessKey = "accesskeys.iam.aws.weberc2.com/cleanup"
	labelAccessKeyName = "accesskeys.iam.aws.weberc2.com/name"
	envUserName        = "AWS_USER"
	envAccessKeyID     = "AWS_ACCESS_KEY_ID"
	envAccessKeySecret = "AWS_SECRET_ACCESS_KEY"
)
