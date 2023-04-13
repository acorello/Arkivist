package notion

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"dev.acorello.it/go/arkivist/notion/database"
	"dev.acorello.it/go/arkivist/notion/page"
	"dev.acorello.it/go/arkivist/osutil"
	"github.com/pelletier/go-toml/v2"
)

type Notion struct {
	BaseURI   url.URL
	AuthToken string
}

func LoadTOMLConfig(filepath string) Notion {
	tomlConf := struct {
		Notion struct {
			AuthToken string
			BaseURI   string
		}
	}{}
	config := osutil.MustOpen(filepath)
	defer config.Close()
	toml.NewDecoder(config).Decode(&tomlConf)
	n := tomlConf.Notion
	baseURI, err := url.Parse(n.BaseURI)
	if err != nil {
		log.Fatalf("Invalid URI %q", n.BaseURI)
	}
	return Notion{
		AuthToken: n.AuthToken,
		BaseURI:   *baseURI,
	}
}

func (me Notion) AddPage(p page.Page) (page.PageResponse, error) {
	_ = me.newRequest("POST", "pages")
	return page.PRZero, nil
}

func (me Notion) Database(id string) (database.Database, error) {
	getDB := me.newRequest("GET", "databases", id)
	res, err := http.DefaultClient.Do(getDB)
	if err != nil {
		return database.Zero, err
	}
	defer res.Body.Close()
	return database.ParseJSON(res.Body)
}

func (me Notion) newRequest(method string, path ...string) *http.Request {
	url := me.BaseURI.JoinPath(path...)
	req, err := http.NewRequest(method, url.String(), nil)
	if err != nil {
		// TODO: verify if there is a case to return the error
		log.Fatal(err)
	}
	me.addHeaders(req)
	return req
}
func (me Notion) addHeaders(req *http.Request) {
	authHeader := fmt.Sprintf("Bearer %s", me.AuthToken)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", "application/json")
	// ❓ required? should I always choose the latest version? how do I know what is the latest version?
	req.Header.Add("Notion-Version", "2021-08-16")
}
