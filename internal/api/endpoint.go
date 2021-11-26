package api

func (a *API) registerEndpoints() {
	for _, fn := range []func(){
		a.registerGetPostsAll,
		a.registerGetPostsReactions,
	} {
		fn()
	}
}
