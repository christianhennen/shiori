package webserver

import (
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/ldap"
	"github.com/go-shiori/shiori/pkg/warc"
	"github.com/julienschmidt/httprouter"
	cch "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

// Options is options for the server
type Options struct {
	DB         database.DB
	DataDir    string
	Address    string
	Port       int
	RootPath      string
	LDAPClient *ldap.Client
}

// ServeApp serves wb interface in specified port
func ServeApp(opts Options) error {
	// Create handler
	hdl := handler{
		DB:           opts.DB,
		DataDir:      opts.DataDir,
		UserCache:    cch.New(time.Hour, 10*time.Minute),
		SessionCache: cch.New(time.Hour, 10*time.Minute),
		ArchiveCache: cch.New(time.Minute, 5*time.Minute),
		RootPath:     cfg.RootPath,
		LDAPClient:   opts.LDAPClient,
	}

	hdl.prepareSessionCache()
	hdl.prepareArchiveCache()

	err := hdl.prepareTemplates()
	if err != nil {
		return fmt.Errorf("failed to prepare templates: %v", err)
	}

	// Create router
	router := httprouter.New()

	// jp here means "join path", as in "join route with root path"
	jp := func(route string) string {
		return path.Join(cfg.RootPath, route)
	}

	router.GET(jp("/js/*filepath"), hdl.serveJsFile)
	router.GET(jp("/res/*filepath"), hdl.serveFile)
	router.GET(jp("/css/*filepath"), hdl.serveFile)
	router.GET(jp("/fonts/*filepath"), hdl.serveFile)

	router.GET(jp("/"), hdl.serveIndexPage)
	router.GET(jp("/login"), hdl.serveLoginPage)
	router.GET(jp("/bookmark/:id/thumb"), hdl.serveThumbnailImage)
	router.GET(jp("/bookmark/:id/content"), hdl.serveBookmarkContent)
	router.GET(jp("/bookmark/:id/archive/*filepath"), hdl.serveBookmarkArchive)

	router.POST(jp("/api/login"), hdl.apiLogin)
	router.POST(jp("/api/logout"), hdl.apiLogout)
	router.GET(jp("/api/bookmarks"), hdl.apiGetBookmarks)
	router.GET(jp("/api/tags"), hdl.apiGetTags)
	router.PUT(jp("/api/tag"), hdl.apiRenameTag)
	router.POST(jp("/api/bookmarks"), hdl.apiInsertBookmark)
	router.DELETE(jp("/api/bookmarks"), hdl.apiDeleteBookmark)
	router.PUT(jp("/api/bookmarks"), hdl.apiUpdateBookmark)
	router.PUT(jp("/api/cache"), hdl.apiUpdateCache)
	router.PUT(jp("/api/bookmarks/tags"), hdl.apiUpdateBookmarkTags)
	router.POST(jp("/api/bookmarks/ext"), hdl.apiInsertViaExtension)
	router.DELETE(jp("/api/bookmarks/ext"), hdl.apiDeleteViaExtension)

	router.GET(jp("/api/accounts"), hdl.apiGetAccounts)
	router.PUT(jp("/api/accounts"), hdl.apiUpdateAccount)
	router.POST(jp("/api/accounts"), hdl.apiInsertAccount)
	router.DELETE(jp("/api/accounts"), hdl.apiDeleteAccount)

	// Route for panic
	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, arg interface{}) {
		http.Error(w, fmt.Sprint(arg), 500)
	}

	// Create server
	url := fmt.Sprintf("%s:%d", opts.Address, opts.Port)
	svr := &http.Server{
		Addr:         url,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: time.Minute,
	}

	// Serve app
	logrus.Infoln("Serve shiori in", url)
	return svr.ListenAndServe()
}
