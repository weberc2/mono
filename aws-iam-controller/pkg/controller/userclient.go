package controller

import (
	"context"
	"errors"
	"fmt"
	v1 "iamcontroller/pkg/api/v1"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

type UserClient struct {
	IAM *iam.Client
}

func (client UserClient) FetchUser(
	ctx Context,
	name string,
) (*types.User, error) {
	rsp, err := client.IAM.GetUser(ctx, &iam.GetUserInput{UserName: &name})
	if err != nil {
		if as[*types.NoSuchEntityException](err) != nil {
			return nil, fmt.Errorf("fetching user: %w", &UserNotFoundErr{name})
		}
		return nil, fmt.Errorf("fetching user `%s`: %w", name, err)
	}
	return rsp.User, nil
}

func (client UserClient) CreateUser(ctx Context, user *v1.User) error {
	if _, err := client.IAM.CreateUser(
		ctx,
		&iam.CreateUserInput{
			UserName: &user.UserName,
			Tags: []types.Tag{
				{
					Key:   aws.String(tagAPIVersion),
					Value: aws.String(v1.SchemeGroupVersion.String()),
				},
				{
					Key:   aws.String(tagKind),
					Value: aws.String(kindUser),
				},
				{
					Key:   aws.String(tagNamespace),
					Value: &user.Namespace,
				},
				{
					Key:   aws.String(tagName),
					Value: &user.Name,
				},
			},
		},
	); err != nil {
		var exists *types.EntityAlreadyExistsException
		if errors.As(err, &exists) {
			return fmt.Errorf(
				"creating user: %w",
				&UserExistsErr{User: user.UserName},
			)
		}
		return fmt.Errorf("creating user `%s`: %w", user.UserName, err)
	}
	return nil
}

func (client UserClient) DeleteUser(ctx Context, name string) error {
	if _, err := client.IAM.DeleteUser(
		ctx,
		&iam.DeleteUserInput{UserName: &name},
	); err != nil {
		if as[*types.NoSuchEntityException](err) != nil {
			return fmt.Errorf("deleting user: %w", &UserNotFoundErr{User: name})
		}
		return fmt.Errorf("deleting user `%s`: %w", name, err)
	}

	return nil
}

func (client UserClient) FetchAccessKey(
	ctx Context,
	user string,
	accessKeyID string,
) (*types.AccessKeyMetadata, error) {
	output, err := client.IAM.ListAccessKeys(
		ctx,
		&iam.ListAccessKeysInput{UserName: &user},
	)
	if err != nil {
		if as[*types.NoSuchEntityException](err) != nil {
			return nil, &UserNotFoundErr{User: user}
		}
		return nil, fmt.Errorf(
			"fetching access key `%s` for user `%s`: %w",
			accessKeyID,
			user,
			err,
		)
	}
	for i := range output.AccessKeyMetadata {
		if *output.AccessKeyMetadata[i].AccessKeyId == accessKeyID {
			return &output.AccessKeyMetadata[i], nil
		}
	}
	return nil, fmt.Errorf(
		"fetching access key: %w",
		&AccessKeyNotFoundErr{User: user, AccessKey: accessKeyID},
	)
}

func (client UserClient) ListAccessKeys(
	ctx Context,
	user string,
) ([]types.AccessKeyMetadata, error) {
	rsp, err := client.IAM.ListAccessKeys(
		ctx,
		&iam.ListAccessKeysInput{UserName: &user},
	)
	if err != nil {
		if as[*types.NoSuchEntityException](err) != nil {
			return nil, fmt.Errorf(
				"listing access keys: %w",
				&UserNotFoundErr{User: user},
			)
		}
		return nil, fmt.Errorf(
			"listing access keys for user `%s`: %w",
			user,
			err,
		)
	}

	return rsp.AccessKeyMetadata, nil
}

func (client UserClient) CreateAccessKey(
	ctx Context,
	user string,
) (*types.AccessKey, error) {
	rsp, err := client.IAM.CreateAccessKey(
		ctx,
		&iam.CreateAccessKeyInput{UserName: &user},
	)
	if err != nil {
		if as[*types.NoSuchEntityException](err) != nil {
			return nil, fmt.Errorf(
				"creating access key: %w",
				&UserNotFoundErr{User: user},
			)
		}
		return nil, fmt.Errorf(
			"creating access key for user `%s`: %w",
			user,
			err,
		)
	}
	return rsp.AccessKey, nil
}

func (client UserClient) DeleteAccessKey(
	ctx Context,
	user string,
	accessKeyID string,
) error {
	if _, err := client.IAM.DeleteAccessKey(
		ctx,
		&iam.DeleteAccessKeyInput{UserName: &user, AccessKeyId: &accessKeyID},
	); err != nil {
		if as[*types.NoSuchEntityException](err) != nil {
			return fmt.Errorf(
				"deleting access key: %w",
				&AccessKeyNotFoundErr{User: user, AccessKey: accessKeyID},
			)
		}
		return fmt.Errorf(
			"deleting access key `%s` for user `%s`: %w",
			accessKeyID,
			user,
			err,
		)
	}
	return nil
}

type UserNotFoundErr struct {
	User string
}

func (err *UserNotFoundErr) Error() string {
	return fmt.Sprintf("user not found: %s", err.User)
}

type UserExistsErr struct {
	User string
}

func (err *UserExistsErr) Error() string {
	return fmt.Sprintf("user exists: %s", err.User)
}

type UserCorrespondenceErr struct {
	UserName   string
	Mismatches [][3]string
}

func (err *UserCorrespondenceErr) Error() string {
	var sb strings.Builder
	fmt.Fprintf(
		&sb,
		"user `%s` does not correspond to k8s resource",
		err.UserName,
	)

	if len(err.Mismatches) < 1 {
		return sb.String()
	}

	fmt.Fprintf(
		&sb,
		": tag `%s` [wanted `%s`; found `%s`]",
		err.Mismatches[0][0],
		err.Mismatches[0][1],
		err.Mismatches[0][1],
	)

	for i := range err.Mismatches[1:] {
		fmt.Fprintf(
			&sb,
			", tag `%s` [wanted `%s`; found `%s`]",
			err.Mismatches[i+1][0],
			err.Mismatches[i+1][1],
			err.Mismatches[i+1][1],
		)
	}

	return sb.String()
}

type AccessKeyNotFoundErr struct {
	User      string
	AccessKey string
}

func (err *AccessKeyNotFoundErr) Error() string {
	return fmt.Sprintf(
		"access key not found for user `%s`: %s",
		err.User,
		err.AccessKey,
	)
}

func as[T error](err error) (out T) {
	errors.As(err, &out)
	return
}

type Context = context.Context
