package httpctx_test

import (
	"net/http"

	"github.com/spkg/httpctx"
	"golang.org/x/net/context"
)

func Example() {
	public := httpctx.Use(ensureHttps)
	authenticate := public.Use(authenticate, ensureAdmin)

	http.Handle("/admin", authenticate.HandleFunc(admin))
	http.Handle("/", public.HandleFunc(index))
	http.ListenAndServe(":8080", nil)
}

func index(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Write([]byte("index page"))
	return nil
}

func admin(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Write([]byte("admin page, userid=" + useridFrom(ctx)))
	return nil
}

// ensureHttps is an example of middleware that ensures that the request scheme is https.
func ensureHttps(h httpctx.Handler) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.URL.Scheme != "https" {
			u := *r.URL
			u.Scheme = "https"
			http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
			return nil
		}
		return h.ServeHTTPContext(ctx, w, r)
	})
}

// authenticate is an example of middleware that authenticates using basic authentication
func authenticate(h httpctx.Handler) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		credentials := r.Header.Get("WWW-Authenticate")
		ctx, err := checkCredentials(ctx, credentials)
		if err != nil {
			return err
		}
		return h.ServeHTTPContext(ctx, w, r)
	})
}

func ensureAdmin(h httpctx.Handler) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		// ... get the userid from the context and ensure that the user has admin privilege ...
		return h.ServeHTTPContext(ctx, w, r)
	})
}

// checkCredentials is a placeholder for a function that checks
// credentials, and if successful returns a context with the
// identity of the user attached as a value.
func checkCredentials(ctx context.Context, credentials string) (context.Context, error) {
	// ... just an example ...
	// Note that you would not normally use a string as the key to context.WithValue
	return context.WithValue(ctx, "userid", "username"), nil
}

func useridFrom(ctx context.Context) string {
	userid, _ := ctx.Value("userid").(string)
	return userid
}
