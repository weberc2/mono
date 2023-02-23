package comments

import (
	pz "github.com/weberc2/httpeasy"
)

type AuthWebServer struct {
	WebServer
	Auth Auth
}

func (aws *AuthWebServer) RepliesRoute() pz.Route {
	return aws.Auth.Optional(aws.WebServer.RepliesRoute())
}

func (aws *AuthWebServer) DeleteConfirmRoute() pz.Route {
	return aws.Auth.Auth(aws.WebServer.DeleteConfirmRoute())
}

func (aws *AuthWebServer) DeleteRoute() pz.Route {
	return aws.Auth.Auth(aws.WebServer.DeleteRoute())
}

func (aws *AuthWebServer) ReplyFormRoute() pz.Route {
	return aws.Auth.Auth(aws.WebServer.ReplyFormRoute())
}

func (aws *AuthWebServer) ReplyRoute() pz.Route {
	return aws.Auth.Auth(aws.WebServer.ReplyRoute())
}

func (aws *AuthWebServer) EditFormRoute() pz.Route {
	return aws.Auth.Auth(aws.WebServer.EditFormRoute())
}

func (aws *AuthWebServer) EditRoute() pz.Route {
	return aws.Auth.Auth(aws.WebServer.EditRoute())
}

func (aws *AuthWebServer) Routes() []pz.Route {
	return []pz.Route{
		aws.RepliesRoute(),
		aws.DeleteConfirmRoute(),
		aws.DeleteRoute(),
		aws.ReplyFormRoute(),
		aws.ReplyRoute(),
		aws.EditFormRoute(),
		aws.EditRoute(),
	}
}
