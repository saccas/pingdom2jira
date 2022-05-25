package main

import (
	r "github.com/unprofession-al/routing"
)

func (s Server) Routes() r.Route {
	return r.Route{
		R: r.Routes{
			"hooks": {
				H: r.Handlers{
					"POST": {F: s.HookHandler, Q: []*r.QueryParam{}},
				},
			},
		},
	}
}
