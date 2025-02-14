// spec package is introduced to avoid circular dependencies since this
// particular test requires to depend on routing directly to expose the API and
// the APP server.
package apps_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andybalholm/brotli"
	apps "github.com/cozy/cozy-stack/model/app"
	"github.com/cozy/cozy-stack/model/instance"
	"github.com/cozy/cozy-stack/model/instance/lifecycle"
	"github.com/cozy/cozy-stack/model/intent"
	"github.com/cozy/cozy-stack/model/oauth"
	"github.com/cozy/cozy-stack/model/session"
	"github.com/cozy/cozy-stack/model/vfs"
	"github.com/cozy/cozy-stack/pkg/assets"
	"github.com/cozy/cozy-stack/pkg/assets/dynamic"
	"github.com/cozy/cozy-stack/pkg/assets/model"
	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/pkg/consts"
	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/tests/testutils"
	"github.com/cozy/cozy-stack/web"
	webApps "github.com/cozy/cozy-stack/web/apps"
	"github.com/labstack/echo/v4"

	"github.com/stretchr/testify/assert"
)

const domain = "cozywithapps.example.net"
const slug = "mini"

var ts *httptest.Server
var testInstance *instance.Instance
var token string

var jar http.CookieJar
var client *http.Client

func compress(content string) []byte {
	buf := &bytes.Buffer{}
	bw := brotli.NewWriter(buf)
	_, _ = bw.Write([]byte(content))
	_ = bw.Close()
	return buf.Bytes()
}

func createFile(dir, filename, content string) error {
	abs := path.Join(dir, filename+".br")
	file, err := vfs.Create(testInstance.VFS(), abs)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(compress(content))
	return err
}

func installMiniApp() error {
	manifest := &couchdb.JSONDoc{
		Type: consts.Apps,
		M: map[string]interface{}{
			"_id":    consts.Apps + "/" + slug,
			"name":   "Mini",
			"icon":   "icon.svg",
			"slug":   slug,
			"source": "git://github.com/cozy/mini.git",
			"state":  apps.Ready,
			"intents": []apps.Intent{
				{
					Action: "PICK",
					Types:  []string{"io.cozy.foos"},
					Href:   "/foo",
				},
			},
			"routes": apps.Routes{
				"/foo": apps.Route{
					Folder: "/",
					Index:  "index.html",
					Public: false,
				},
				"/bar": apps.Route{
					Folder: "/bar",
					Index:  "index.html",
					Public: false,
				},
				"/public": apps.Route{
					Folder: "/public",
					Index:  "index.html",
					Public: true,
				},
			},
		},
	}

	err := couchdb.CreateNamedDoc(testInstance, manifest)
	if err != nil {
		return err
	}

	appdir := path.Join(vfs.WebappsDirName, slug)
	_, err = vfs.MkdirAll(testInstance.VFS(), appdir)
	if err != nil {
		return err
	}
	bardir := path.Join(appdir, "bar")
	_, err = vfs.Mkdir(testInstance.VFS(), bardir, nil)
	if err != nil {
		return err
	}
	pubdir := path.Join(appdir, "public")
	_, err = vfs.Mkdir(testInstance.VFS(), pubdir, nil)
	if err != nil {
		return err
	}

	err = createFile(appdir, "icon.svg", "<svg>...</svg>")
	if err != nil {
		return err
	}
	err = createFile(appdir, "index.html", `this is index.html. <a lang="{{.Locale}}" href="https://{{.Domain}}/status/">Status</a> {{.Favicon}}`)
	if err != nil {
		return err
	}
	err = createFile(bardir, "index.html", "{{.CozyBar}}")
	if err != nil {
		return err
	}
	err = createFile(appdir, "hello.html", "world {{.Token}}")
	if err != nil {
		return err
	}
	err = createFile(pubdir, "index.html", "this is a file in public/")
	return err
}

func doGet(path string, auth bool) (*http.Response, error) {
	c := client
	if !auth {
		c = &http.Client{CheckRedirect: noRedirect}
	}
	req, err := http.NewRequest("GET", ts.URL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Host = slug + "." + testInstance.Domain
	return c.Do(req)
}

func doGetAll(t *testing.T, path string, auth bool) []byte {
	res, err := doGet(path, auth)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	return body
}

func assertGet(t *testing.T, contentType, content string, res *http.Response) {
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, contentType, res.Header.Get("Content-Type"))
	body, _ := ioutil.ReadAll(res.Body)
	assert.Contains(t, string(body), content)
}

func assertAuthGet(t *testing.T, path, contentType, content string) {
	res, err := doGet(path, true)
	assert.NoError(t, err)
	assertGet(t, contentType, content, res)
}

func assertAnonGet(t *testing.T, path, contentType, content string) {
	res, err := doGet(path, false)
	assert.NoError(t, err)
	assertGet(t, contentType, content, res)
}

func assertNotPublic(t *testing.T, path string, code int, location string) {
	res, err := doGet(path, false)
	assert.NoError(t, err)
	assert.Equal(t, code, res.StatusCode)
	if 300 <= code && code < 400 {
		assert.Equal(t, location, res.Header.Get("location"))
	}
}

func assertNotFound(t *testing.T, path string) {
	res, err := doGet(path, true)
	assert.NoError(t, err)
	assert.Equal(t, 404, res.StatusCode)
}

func TestServe(t *testing.T) {
	assertAuthGet(t, "/foo/", "text/html; charset=utf-8", `this is index.html. <a lang="en" href="https://cozywithapps.example.net/status/">Status</a>`)
	assertAuthGet(t, "/foo/hello.html", "text/html; charset=utf-8", "world {{.Token}}")
	assertAuthGet(t, "/public", "text/html; charset=utf-8", "this is a file in public/")
	assertAuthGet(t, "/public/index.html", "text/html; charset=utf-8", "this is a file in public/")
	assertAnonGet(t, "/public", "text/html; charset=utf-8", "this is a file in public/")
	assertAnonGet(t, "/public/index.html", "text/html; charset=utf-8", "this is a file in public/")
	assertNotPublic(t, "/foo", 302, "https://cozywithapps.example.net/auth/login?redirect=https%3A%2F%2Fmini.cozywithapps.example.net%2Ffoo")
	assertNotPublic(t, "/foo/hello.tml", 401, "")
	assertNotFound(t, "/404")
	assertNotFound(t, "/")
	assertNotFound(t, "/index.html")
	assertNotFound(t, "/public/hello.html")
}

func TestCozyBar(t *testing.T) {
	body := doGetAll(t, "/bar/", true)
	assert.Contains(t, string(body), `<link rel="stylesheet" type="text/css" href="//cozywithapps.example.net/assets/css/cozy-bar`)
	assert.Contains(t, string(body), `<script src="//cozywithapps.example.net/assets/js/cozy-bar`)
}

func TestServeWithAnIntents(t *testing.T) {
	intent := &intent.Intent{
		Action: "PICK",
		Type:   "io.cozy.foos",
		Client: "io.cozy.apps/test-app",
	}
	err := intent.Save(testInstance)
	assert.NoError(t, err)
	err = intent.FillServices(testInstance)
	assert.NoError(t, err)
	assert.Len(t, intent.Services, 1)
	err = intent.Save(testInstance)
	assert.NoError(t, err)

	path := strings.Replace(intent.Services[0].Href, "https://mini.cozywithapps.example.net", "", 1)
	res, err := doGet(path, true)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	h := res.Header.Get(echo.HeaderContentSecurityPolicy)
	assert.Contains(t, h, "frame-ancestors 'self' https://test-app.cozywithapps.example.net/;")
}

func TestFaviconWithContext(t *testing.T) {
	context := "foo"

	asset, ok := assets.Get("/favicon.ico", context)
	if ok {
		_ = assets.Remove(asset.Name, asset.Context)
	}
	// Create and insert an asset in foo context
	tmpdir := os.TempDir()
	_, err := os.OpenFile(filepath.Join(tmpdir, "custom_favicon.png"), os.O_RDWR|os.O_CREATE, 0600)
	assert.NoError(t, err)

	assetsOptions := []model.AssetOption{{
		URL:     fmt.Sprintf("file://%s", filepath.Join(tmpdir, "custom_favicon.png")),
		Name:    "/favicon.ico",
		Context: context,
	}}
	err = dynamic.RegisterCustomExternals(assetsOptions, 1)
	assert.NoError(t, err)

	// Test the theme
	assert.NoError(t, lifecycle.Patch(testInstance, &lifecycle.Options{
		ContextName: context,
	}))
	assert.NoError(t, err)

	res, err := doGet("/foo", true)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	body, _ := ioutil.ReadAll(res.Body)
	expected := `this is index.html. <a lang="en" href="https://cozywithapps.example.net/status/">Status</a>`
	assert.Contains(t, string(body), expected)
	assert.Contains(t, string(body), fmt.Sprintf("/assets/ext/%s/favicon.ico", context))
	assert.NotContains(t, string(body), "/assets/favicon.ico")
}

func TestServeAppsWithJWTNotLogged(t *testing.T) {
	config.GetConfig().Subdomains = config.FlatSubdomains
	appHost := "cozywithapps-mini.example.net"

	req, _ := http.NewRequest("GET", ts.URL+"/foo?jwt=abc", nil)
	req.Host = appHost
	c := &http.Client{CheckRedirect: noRedirect}
	res, err := c.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 302, res.StatusCode)
	location, err := url.Parse(res.Header.Get("Location"))
	assert.NoError(t, err)
	assert.Equal(t, "/auth/login", location.Path)

	assert.Equal(t, testInstance.Domain, location.Host)
	assert.NotEmpty(t, location.Query().Get("redirect"))
	assert.Equal(t, "abc", location.Query().Get("jwt"))
}

func TestOauthAppCantInstallApp(t *testing.T) {
	req, _ := http.NewRequest("POST", ts.URL+"/apps/mini-bis?Source=git://github.com/nono/cozy-mini.git", nil)
	req.Header.Add("Authorization", "Bearer "+token)
	req.Host = testInstance.Domain
	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 403, res.StatusCode)
}

func TestOauthAppCantUpdateApp(t *testing.T) {
	req, _ := http.NewRequest("PUT", ts.URL+"/apps/mini", nil)
	req.Header.Add("Authorization", "Bearer "+token)
	req.Host = testInstance.Domain
	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 403, res.StatusCode)
}

func TestListApps(t *testing.T) {
	req, _ := http.NewRequest("GET", ts.URL+"/apps/", nil)
	req.Header.Add("Authorization", "Bearer "+token)
	req.Host = testInstance.Domain
	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

	var results map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&results)
	assert.NoError(t, err)
	objs := results["data"].([]interface{})
	assert.Len(t, objs, 1)
	data := objs[0].(map[string]interface{})
	id := data["id"].(string)
	assert.NotEmpty(t, id)
	typ := data["type"].(string)
	assert.Equal(t, "io.cozy.apps", typ)

	attrs := data["attributes"].(map[string]interface{})
	name := attrs["name"].(string)
	assert.Equal(t, "Mini", name)
	slug := attrs["slug"].(string)
	assert.Equal(t, "mini", slug)

	links := data["links"].(map[string]interface{})
	self := links["self"].(string)
	assert.Equal(t, "/apps/mini", self)
	related := links["related"].(string)
	assert.Equal(t, "https://cozywithapps-mini.example.net/", related)
	icon := links["icon"].(string)
	assert.Equal(t, "/apps/mini/icon/", icon)
}

func TestIconForApp(t *testing.T) {
	req, _ := http.NewRequest("GET", ts.URL+"/apps/mini/icon", nil)
	req.Header.Add("Authorization", "Bearer "+token)
	req.Host = testInstance.Domain
	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	body, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, "<svg>...</svg>", string(body))
}

func TestUninstallAppWithLinkedClient(t *testing.T) {
	// Install drive app
	installer, err := apps.NewInstaller(testInstance, apps.Copier(consts.WebappType, testInstance),
		&apps.InstallerOptions{
			Operation:  apps.Install,
			Type:       consts.WebappType,
			Slug:       "drive",
			SourceURL:  "registry://drive",
			Registries: testInstance.Registries(),
		},
	)
	assert.NoError(t, err)
	_, err = installer.RunSync()
	assert.NoError(t, err)

	// Link an OAuthClient to drive
	oauthClient := &oauth.Client{
		ClientName:   "test-linked",
		RedirectURIs: []string{"https://foobar"},
		SoftwareID:   "registry://drive",
	}

	oauthClient.Create(testInstance)
	// Forcing the oauthClient to get a couchID for the purpose of later deletion
	oauthClient, err = oauth.FindClient(testInstance, oauthClient.ClientID)
	assert.NoError(t, err)

	scope := "io.cozy.apps:ALL"
	token, err := testInstance.MakeJWT("cli", "drive", scope, "", time.Now())
	assert.NoError(t, err)

	// Trying to remove this app
	req, _ := http.NewRequest("DELETE", ts.URL+"/apps/drive", nil)
	req.Host = testInstance.Domain

	req.Header.Add("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, res.StatusCode)
	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "linked OAuth client exists")

	// Cleaning
	uninstaller, err := apps.NewInstaller(testInstance, apps.Copier(consts.WebappType, testInstance),
		&apps.InstallerOptions{
			Operation:  apps.Delete,
			Type:       consts.WebappType,
			Slug:       "drive",
			SourceURL:  "registry://drive",
			Registries: testInstance.Registries(),
		},
	)
	assert.NoError(t, err)
	_, err = uninstaller.RunSync()
	assert.NoError(t, err)
	errc := oauthClient.Delete(testInstance)
	assert.Nil(t, errc)
}

func TestUninstallAppWithoutLinkedClient(t *testing.T) {
	// Install drive app
	installer, err := apps.NewInstaller(testInstance, apps.Copier(consts.WebappType, testInstance),
		&apps.InstallerOptions{
			Operation:  apps.Install,
			Type:       consts.WebappType,
			Slug:       "drive",
			SourceURL:  "registry://drive",
			Registries: testInstance.Registries(),
		},
	)
	assert.NoError(t, err)
	_, err = installer.RunSync()
	assert.NoError(t, err)

	// Create an OAuth client not linked to drive
	oauthClient := &oauth.Client{
		ClientName:   "test-linked",
		RedirectURIs: []string{"https://foobar"},
		SoftwareID:   "foobarclient",
	}
	oauthClient.Create(testInstance)
	// Forcing the oauthClient to get a couchID for the purpose of later deletion
	oauthClient, err = oauth.FindClient(testInstance, oauthClient.ClientID)
	assert.NoError(t, err)

	scope := "io.cozy.apps:ALL"
	token, err := testInstance.MakeJWT("cli", "drive", scope, "", time.Now())
	assert.NoError(t, err)

	// Trying to remove this app
	req, _ := http.NewRequest("DELETE", ts.URL+"/apps/drive", nil)
	req.Host = testInstance.Domain

	req.Header.Add("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

	// Cleaning
	errc := oauthClient.Delete(testInstance)
	assert.Nil(t, errc)
}

func TestMain(m *testing.M) {
	config.UseTestFile()
	config.GetConfig().Assets = "../../assets"
	testutils.NeedCouchdb()
	setup := testutils.NewSetup(m, "apps_test")
	err := setup.SetupSwiftTest()
	if err != nil {
		panic("Could not init Swift test")
	}
	err = dynamic.InitDynamicAssetFS()
	if err != nil {
		panic("Could not init dynamic FS")
	}
	tempdir := setup.GetTmpDirectory()

	cfg := config.GetConfig()
	cfg.Fs.URL = &url.URL{
		Scheme: "file",
		Host:   "localhost",
		Path:   tempdir,
	}
	was := cfg.Subdomains
	cfg.Subdomains = config.NestedSubdomains
	defer func() { cfg.Subdomains = was }()

	pass := "aephe2Ei"
	testInstance = setup.GetTestInstance(&lifecycle.Options{Domain: domain})
	params := lifecycle.PassParameters{
		Key:        "fake-encrypt-key",
		Iterations: 0,
	}
	_ = lifecycle.ForceUpdatePassphrase(testInstance, []byte(pass), params)
	testInstance.RegisterToken = nil
	testInstance.OnboardingFinished = true
	_ = testInstance.Update()

	err = installMiniApp()
	if err != nil {
		setup.CleanupAndDie("Could not install mini app.", err)
	}

	ts = setup.GetTestServer("/apps", webApps.WebappsRoutes, func(r *echo.Echo) *echo.Echo {
		r.POST("/login", func(c echo.Context) error {
			longRunSession := true
			sess, _ := session.New(testInstance, longRunSession)
			cookie, _ := sess.ToCookie()
			c.SetCookie(cookie)
			return c.HTML(http.StatusOK, "OK")
		})
		router, err := web.CreateSubdomainProxy(r, webApps.Serve)
		if err != nil {
			setup.CleanupAndDie("Cant start subdoman proxy", err)
		}
		return router
	})

	jar = setup.GetCookieJar()
	client = &http.Client{Jar: jar}

	// Login
	req, _ := http.NewRequest("POST", ts.URL+"/login", bytes.NewBufferString("passphrase="+pass))
	req.Host = testInstance.Domain
	_, _ = client.Do(req)

	_, token = setup.GetTestClient(consts.Apps + " io.cozy.registry.webapps")

	os.Exit(setup.Run())
}

func noRedirect(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}
