/*
   DbMan - © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
   Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
   Contributors to this project, hereby assign copyright in this code to the project,
   to be licensed under the same terms as the rest of the code.
*/

package core

// @title DbMan
// @version 1.0.0
// @description Call DbMan's commands using HTTP operations from anywhere.
// @contact.name SouthWinds Tech Ltd
// @contact.url https://www.southwinds.io/
// @contact.email info@southwinds.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	_ "southwinds.dev/dbman/docs" // documentation needed for swagger
	"southwinds.dev/dbman/plugin"
	h "southwinds.dev/http"
	"strings"
)

type Server struct {
	*h.Server
	cfg *Config
}

func NewServer(cfg *Config) *Server {
	s := &Server{}
	s.Server = h.New("dbman", "")
	s.cfg = cfg
	return s
}

// a liveliness probe to prove the http service is listening
func (s *Server) liveHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("OK"))
	if err != nil {
		fmt.Printf("!!! I cannot write response: %v", err)
	}
}

// @Summary Check that DbMan is Ready
// @Description Checks that DbMan is ready to accept calls
// @Tags General
// @Produce  plain
// @Success 200 {string} OK
// @Failure 500 {string} error message
// @Router /ready [get]
func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	ready, err := DM.CheckReady()
	if !ready {
		fmt.Printf("! I am not ready: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
		}
	} else {
		_, _ = w.Write([]byte("OK"))
	}
}

// @Summary Retrieves database server information
// @Description Gets specific information about the database server to which DbMan is configured to connect.
// @Tags Database
// @Produce  application/json, application/yaml
// @Success 200 {json} database server information
// @Failure 500 {string} error message
// @Router /db/info/server [get]
func (s *Server) dbServerHandler(w http.ResponseWriter, r *http.Request) {
	info, err := DM.GetDbInfo()
	if err != nil {
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("I cannot get database server information: %v\n", err))
		return
	}
	h.Write(w, r, info)
}

// @Summary Gets a list of available queries.
// @Description Lists all of the queries declared in the current release manifest.
// @Tags Database
// @Produce  application/json, application/yaml
// @Success 200 {string} configuration variables
// @Failure 500 {string} error message
// @Router /db/info/queries [get]
func (s *Server) queriesHandler(writer http.ResponseWriter, request *http.Request) {
	// get the release manifest for the current application version
	_, manifest, err := DM.GetReleaseInfo(DM.Cfg.GetString(AppVersion))
	if err != nil {
		h.Err(writer, http.StatusInternalServerError, fmt.Sprintf("I cannot fetch release information: %v\n", err))
		return
	}
	h.Write(writer, request, manifest.Queries)
}

// @Summary Runs a query.
// @Description Execute a query defined in the release manifest and return the result as a generic serializable table.
// @Tags Database
// @Produce  application/json, application/yaml, application/xml, text/csv, text/html, application/xhtml+xml
// @Success 200 {Table} a generic table
// @Failure 500 {string} error message
// @Param name path string true "the name of the query as defined in the release manifest"
// @Param params query string false "a string of parameters to be passed to the query in the format 'key1=value1,...,keyN=valueN'"
// @Router /db/query/{name} [get]
func (s *Server) queryHandler(writer http.ResponseWriter, request *http.Request) {
	// get request variables
	vars := mux.Vars(request)
	queryName := vars["name"]
	// if no query name has been specified it cannot continue
	if len(queryName) == 0 {
		h.Err(writer, http.StatusBadRequest, fmt.Sprintf("!!! I cannot run the query as a query name has not been provided\n"))
		return
	}
	// now check the query has parameters
	queryParams := request.URL.Query()["params"]
	params := make(map[string]string)
	if len(queryParams) > 0 {
		parts := strings.Split(queryParams[0], ",")
		for _, part := range parts {
			subPart := strings.Split(part, "=")
			if len(subPart) != 2 {
				fmt.Printf("I cannot break down query parameter '%s': format should be 'key=value'\n", subPart)
				return
			}
			params[strings.Trim(subPart[0], " ")] = strings.Trim(subPart[1], " ")
		}
	}
	table, query, _, err := DM.Query(queryName, params)
	if err != nil {
		h.Err(writer, http.StatusInternalServerError, fmt.Sprintf("!!! I cannot execute the query: %v\n", err))
		return
	}
	accept := request.Header.Get("Accept")
	// if a html representation was requested as part of the Accept http header
	if strings.Index(accept, "text/html") != -1 || strings.Index(accept, "application/xhtml+xml") != -1 {
		writer.Header().Set("Content-Type", "text/html")
		// determines the http scheme
		// assume http by default
		var scheme = "http"
		// if the http server has set the TLS value then changes the scheme to https
		if request.TLS != nil {
			scheme = "https"
		}
		uri := fmt.Sprintf("%s://%s%s", scheme, request.Host, request.URL.Path)
		if len(request.URL.RawQuery) > 0 {
			uri += fmt.Sprintf("?%s", request.URL.RawQuery)
		}
		theme := DM.getTheme(s.cfg.GetString(ThemeName))
		err = table.AsHTML(writer, &plugin.HtmlTableVars{
			Title:       query.Name,
			Description: query.Description,
			QueryURI:    uri,
			Style:       theme.Style,
			Header:      theme.Header,
			Footer:      theme.Footer,
		})
		if err != nil {
			h.Err(writer, http.StatusInternalServerError, fmt.Sprintf("I cannot execute the query: %v\n", err))
			return
		}
		return
	} else {
		// renders any other representations
		h.Write(writer, request, *table)
	}
}

// @Summary Creates a new database
// @Description When the database does not already exists, this operation executes the manifest commands required to create the new database.
// @Tags Database
// @Produce  plain
// @Success 200 {string} execution logs
// @Failure 500 {string} error message
// @Router /db/create [post]
func (s *Server) createHandler(w http.ResponseWriter, r *http.Request) {
	// deploy the schema and functions
	output, err, elapsed := DM.Create()
	w.Write([]byte(output.String()))
	// return an error if failed
	if err != nil {
		h.Err(w, http.StatusInternalServerError, err.Error())
	} else {
		_, err = w.Write([]byte(fmt.Sprintf("? I have completed the action in %v\n", elapsed)))
		if err != nil {
			fmt.Printf("!!! I failed to write error to response: %v", err)
		}
	}
}

// @Summary Deploys the schema and objects in an empty database.
// @Description When the database is empty, this operation executes the manifest commands required to deploy the  database schema and objects.
// @Tags Database
// @Produce  plain
// @Success 200 {string} execution logs
// @Failure 500 {string} error message
// @Router /db/deploy [post]
func (s *Server) deployHandler(w http.ResponseWriter, r *http.Request) {
	// deploy the schema and functions
	output, err, elapsed := DM.Deploy()
	w.Write([]byte(output.String()))
	// return an error if failed
	if err != nil {
		h.Err(w, http.StatusInternalServerError, err.Error())
	} else {
		_, err = w.Write([]byte(fmt.Sprintf("? I have completed the action in %v\n", elapsed)))
		if err != nil {
			fmt.Printf("!!! I failed to write error to response: %v", err)
		}
	}
}

// @Summary Upgrade a database to a specific version.
// @Description This operation executes the manifest commands required to upgrade an existing database schema and objects to a new version. The target version is defined by DbMan's configuration value "AppVersion". This operation support rolling upgrades.
// @Tags Database
// @Produce  plain
// @Success 200 {string} execution logs
// @Failure 500 {string} error message
// @Router /db/upgrade [post]
func (s *Server) upgradeHandler(w http.ResponseWriter, r *http.Request) {
	// deploy the schema and functions
	output, err, elapsed := DM.Upgrade()
	w.Write([]byte(output.String()))
	// return an error if failed
	if err != nil {
		h.Err(w, http.StatusInternalServerError, err.Error())
	} else {
		_, err = w.Write([]byte(fmt.Sprintf("? I have completed the action in %v\n", elapsed)))
		if err != nil {
			fmt.Printf("!!! I failed to write error to response: %v", err)
		}
	}
}

// @Summary Validates the current DbMan's configuration.
// @Description Checks that the information in the current configuration set is ok to connect to backend services and the format of manifest is correct.
// @Tags Configuration
// @Produce  plain
// @Success 200 {string} execution logs
// @Failure 500 {string} error message
// @Router /conf/check [get]
func (s *Server) checkConfigHandler(w http.ResponseWriter, r *http.Request) {
	results := DM.CheckConfigSet()
	for check, result := range results {
		_, _ = w.Write([]byte(fmt.Sprintf("[%v] => %v\n", check, result)))
	}
}

// @Summary Show DbMan's current configuration.
// @Description Lists all variables in DbMan's configuration.
// @Tags Configuration
// @Produce  plain
// @Success 200 {string} configuration variables
// @Failure 500 {string} error message
// @Router /conf [get]
func (s *Server) showConfigHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(DM.ConfigSetAsString()))
}
