package controller

import (
	"context"
	"fmt"
	v1 "iamcontroller/pkg/api/v1"
	"iamcontroller/pkg/log"
	"log/slog"
	"slices"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type UserReconciler struct {
	Client         client.Client
	Users          UserClient
	ResyncInterval time.Duration
}

func (reconciler *UserReconciler) Configure(manager manager.Manager) error {
	if reconciler.ResyncInterval == 0 {
		reconciler.ResyncInterval = 5 * time.Minute
	}
	return ctrl.NewControllerManagedBy(manager).
		For(&v1.User{}).
		Owns(&corev1.Secret{}).
		Complete(reconciler)
}

func (reconciler *UserReconciler) Reconcile(
	ctx context.Context,
	r reconcile.Request,
) (res reconcile.Result, err error) {
	defer func() {
		if err == nil {
			// ensure we retry reconciliation at least once every 5 minutes
			res.RequeueAfter = reconciler.ResyncInterval
		}
	}()

	logger := log.FromContext(ctx)
	logger.Debug("fetching User from API server")

	var user v1.User
	if err = reconciler.Client.Get(ctx, r.NamespacedName, &user); err != nil {
		err = fmt.Errorf("fetching user from api server: %w", err)
		return
	}

	// if the user resource exists, check to see if it is in a deleting state,
	// and clean up any associated resources.
	if !user.ObjectMeta.DeletionTimestamp.IsZero() {
		slog.Info("deleting user from aws", "awsUser", user.UserName)
		if err = reconciler.Users.DeleteUser(ctx, user.UserName); err != nil {
			// if the user is already deleted; continue removing the finalizer;
			// otherwise, return the error.
			if as[*UserNotFoundErr](err) == nil {
				logger.Error("deleting external user", "err", err.Error())
				err = fmt.Errorf("deleting external user: %w", err)
				return
			}
			logger.Warn(
				"attempted to delete external user, but user was already "+
					"deleted. continuing.",
				"awsUser", user.UserName,
			)
		}

		// remove the finalizer so the object can be deleted
		if i := slices.Index(user.Finalizers, finalizerUser); i >= 0 {
			slog.Debug("removing finalizer", "finalizer", finalizerUser)
			user.Finalizers = slices.Delete(user.Finalizers, i, i+1)
			if err = reconciler.Client.Update(ctx, &user); err != nil {
				err = fmt.Errorf("removing finalizer: updating user: %w", err)
			}
			return
		}
	}

	if !slices.Contains(user.Finalizers, finalizerUser) {
		logger.Debug(
			"finalizer not found; attaching finalizer",
			"finalizer", finalizerUser,
		)
		user.Finalizers = append(user.Finalizers, finalizerUser)
		if err = reconciler.Client.Update(ctx, &user); err != nil {
			err = fmt.Errorf(
				"attaching finalizer: updating user `%s`: %w",
				user.UserName,
				err,
			)
			return
		}
	}

	var awsUser *types.User
	if awsUser, err = reconciler.Users.FetchUser(
		ctx,
		user.UserName,
	); err != nil {
		if as[*UserNotFoundErr](err) != nil {
			// if there is no user with the desired name, create it
			logger.Info(
				"external user not found in AWS; creating external user",
			)
			err = reconciler.Users.CreateUser(ctx, &user)
		} else {
			// otherwise return other errors
			err = fmt.Errorf("fetching aws user `%s`: %w", user.UserName, err)
		}

		return
	}

	if err = matchUsers(&user, awsUser); err != nil {
		return
	}

	err = reconciler.reconcileSecrets(ctx, logger, &user)
	return
}

func (reconciler *UserReconciler) reconcileSecrets(
	ctx context.Context,
	logger *slog.Logger,
	user *v1.User,
) (err error) {
	logger.Debug("reconciling secrets")
	defer func() {
		if err != nil {
			err = fmt.Errorf("reconciling user secrets: %w", err)
		}
	}()

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: user.Namespace,
			Name:      "access-key-" + user.Name,
		},
	}
	if err = reconciler.Client.Get(
		ctx,
		client.ObjectKey{Namespace: user.Namespace, Name: secret.Name},
		&secret,
	); err != nil {
		// if there is no secret corresponding to this user, create one--the
		// access key secret controller will populate it.
		if errors.IsNotFound(err) {
			if err = controllerutil.SetControllerReference(
				user,
				&secret,
				reconciler.Client.Scheme(),
			); err != nil {
				err = fmt.Errorf(
					"creating new user secret: setting owner reference: %w",
					err,
				)
				return
			}
			secret.StringData = map[string]string{envUserName: user.UserName}
			if err = reconciler.Client.Create(ctx, &secret); err != nil {
				err = fmt.Errorf("creating new user secret: %w", err)
				return
			}
			logger.Debug("created user secret", "secret", secret.Name)
		}
		return
	}

	return
}

func matchUsers(local *v1.User, external *types.User) error {
	var apiVersion, kind, namespace, name string
	for i := range external.Tags {
		switch tag := external.Tags[i]; *tag.Key {
		case tagAPIVersion:
			apiVersion = *tag.Value
		case tagKind:
			kind = *tag.Value
		case tagNamespace:
			namespace = *tag.Value
		case tagName:
			name = *tag.Value
		}
	}

	var mismatches [][3]string
	if apiVersion != v1.SchemeGroupVersion.String() {
		mismatches = append(mismatches, [3]string{
			tagAPIVersion,
			v1.SchemeGroupVersion.String(),
			apiVersion,
		})
	}

	if kind != kindUser {
		mismatches = append(mismatches, [3]string{tagKind, kindUser, kind})
	}

	if namespace != local.Namespace {
		mismatches = append(
			mismatches,
			[3]string{tagNamespace, local.Namespace, namespace},
		)
	}

	if name != local.Name {
		mismatches = append(mismatches, [3]string{tagName, local.Name, name})
	}

	if len(mismatches) > 0 {
		return &UserCorrespondenceErr{
			UserName:   local.UserName,
			Mismatches: mismatches,
		}
	}

	return nil
}

const (
	tagAPIVersion = "weberc2.com/api-version"
	tagKind       = "k8s.io/kind"
	tagNamespace  = "k8s.io/namespace"
	tagName       = "k8s.io/name"
	finalizerUser = "users.iam.aws.weberc2.com/cleanup"
	labelUserName = "users.iam.aws.weberc2.com/name"
	kindUser      = "User"
)
