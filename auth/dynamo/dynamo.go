package dynamo

import (
	"encoding/json"
	"fmt"

	"github.com/negz/kubehook/auth"
	"github.com/prometheus/common/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	// DefaultUserTable is the default DynamoDB table in which to look for
	// users.
	DefaultUserTable string = "kubehook-users"

	tokenKey    string = "token"
	userInfoKey string = "userInfo"
)

// ErrNoToken is returned when an empty token is passed.
var ErrNoToken = errors.New("you must provide a token")

type dynamo struct {
	log       *zap.Logger
	d         dynamodbiface.DynamoDBAPI
	userTable string
}

// A Option represents an argument to NewBackend
type Option func(*dynamo) error

// UserTable specifies the Dynamo table in which the set of valid users is
// stored.
func UserTable(t string) Option {
	return func(a *dynamo) error {
		a.userTable = t
		return nil
	}
}

// Logger allows the use of a custom Zap logger.
func Logger(l *zap.Logger) Option {
	return func(a *dynamo) error {
		a.log = l
		return nil
	}
}

// NewAuthenticator returns an authenticator that authenticates any supplied
// token as the username passed to it as a token.
func NewAuthenticator(api dynamodbiface.DynamoDBAPI, ao ...Option) (auth.Authenticator, error) {
	l, err := zap.NewProduction()
	if err != nil {
		return nil, errors.Wrap(err, "cannot create default logger")
	}
	a := &dynamo{log: l, d: api, userTable: DefaultUserTable}
	for _, o := range ao {
		if err := o(a); err != nil {
			return nil, errors.Wrap(err, "cannot apply dynamo authenticator option")
		}
	}
	return a, nil
}

func (a *dynamo) getUser(token string) (string, error) {
	rsp, err := a.d.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(a.userTable),
		Key: map[string]*dynamodb.AttributeValue{
			tokenKey: {S: aws.String(token)},
		},
		ProjectionExpression: aws.String(userInfoKey),
	})
	if err != nil {
		msg := "cannot get user from Dynamo"
		log.Error(msg, zap.Error(err), zap.String(tokenKey, tokenKey))
		return "", errors.Wrapf(err, msg)
	}

	if rsp.Item != nil {
		for _, attr := range rsp.Item {
			if attr.S != nil {
				return aws.StringValue(attr.S), nil
			}
		}
	}
	return "", errors.Errorf("token not found in %s", a.userTable)
}

func (a *dynamo) Authenticate(token string) (*auth.User, error) {
	if token == "" {
		a.log.Info("authentication", zap.Bool("success", false), zap.Error(ErrNoToken))
		return nil, errors.Wrap(ErrNoToken, "authentication failed")
	}

	j, err := a.getUser(token)
	if err != nil {
		a.log.Info("authentication", zap.Bool("success", false))
		return nil, errors.Wrap(err, "authentication failed")
	}

	u := &auth.User{}
	if err := json.Unmarshal([]byte(j), u); err != nil {
		a.log.Info("authentication", zap.Bool("success", false), zap.Error(err))
		return nil, errors.Wrap(err, "authentication failed")
	}

	// Overwrite the UID with something (hopefully) unique.
	u.UID = fmt.Sprintf("%s/%s", a.userTable, u.Username)

	a.log.Info("authentication", zap.Bool("success", true), zap.String("uid", u.UID))
	return u, nil
}
