package main

import (
	r "github.com/unprofession-al/routing"
)

func (s Server) Routes() r.Route {
	return r.Route{
		R: r.Routes{
			"hook": {
				H: r.Handlers{
					"POST": {F: s.HookHandler, Q: []*r.QueryParam{}},
				},
			},
		},
	}
}
