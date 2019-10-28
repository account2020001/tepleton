package baseapp

import (
	"encoding/json"
	"fmt"
	"regexp"

	sdk "github.com/tepleton/tepleton-sdk/types"
)

// Router provides handlers for each transaction type.
type Router interface {
	AddRoute(r string, h sdk.Handler, i sdk.InitGenesis) (rtr Router)
	Route(path string) (h sdk.Handler)
	InitGenesis(ctx sdk.Context, data map[string]json.RawMessage) error
}

// map a transaction type to a handler and an initgenesis function
type route struct {
	r string
	h sdk.Handler
	i sdk.InitGenesis
}

type router struct {
	routes []route
}

// nolint
// NewRouter - create new router
// TODO either make Function unexported or make return type (router) Exported
func NewRouter() *router {
	return &router{
		routes: make([]route, 0),
	}
}

var isAlpha = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString

// AddRoute - TODO add description
func (rtr *router) AddRoute(r string, h sdk.Handler, i sdk.InitGenesis) Router {
	if !isAlpha(r) {
		panic("route expressions can only contain alphanumeric characters")
	}
	rtr.routes = append(rtr.routes, route{r, h, i})

	return rtr
}

// Route - TODO add description
// TODO handle expressive matches.
func (rtr *router) Route(path string) (h sdk.Handler) {
	for _, route := range rtr.routes {
		if route.r == path {
			return route.h
		}
	}
	return nil
}

// InitGenesis - call `InitGenesis`, where specified, for all routes
// Return the first error if any, otherwise nil
func (rtr *router) InitGenesis(ctx sdk.Context, data map[string]json.RawMessage) error {
	for _, route := range rtr.routes {
		if route.i != nil {
			encoded, found := data[route.r]
			if !found {
				return sdk.ErrGenesisParse(fmt.Sprintf("Expected module genesis information for module %s but it was not present", route.r))
			}
			if err := route.i(ctx, encoded); err != nil {
				return err
			}
		}
	}
	return nil
}
