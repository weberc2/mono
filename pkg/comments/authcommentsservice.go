package comments

import (
	pz "github.com/weberc2/httpeasy"
)

type AuthCommentsService struct {
	CommentsService
	Auth Auth
}

func (acs *AuthCommentsService) Routes() []pz.Route {
	return []pz.Route{
		acs.RepliesRoute(),
		acs.PutRoute(),
		acs.GetRoute(),
		acs.DeleteRoute(),
		acs.UpdateRoute(),
	}
}

func (acs *AuthCommentsService) PutRoute() pz.Route {
	return acs.Auth.Auth(acs.CommentsService.PutRoute())
}

func (acs *AuthCommentsService) DeleteRoute() pz.Route {
	return acs.Auth.Auth(acs.CommentsService.DeleteRoute())
}

func (acs *AuthCommentsService) UpdateRoute() pz.Route {
	return acs.Auth.Auth(acs.CommentsService.UpdateRoute())
}
