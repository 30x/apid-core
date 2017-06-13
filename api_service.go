// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apid

import "net/http"

type APIService interface {
	Listen() error
	Handle(path string, handler http.Handler) Route
	HandleFunc(path string, handlerFunc http.HandlerFunc) Route
	Vars(r *http.Request) map[string]string

	// for testing
	Router() Router
}

type Route interface {
	Methods(methods ...string) Route
}

// for testing
type Router interface {
	Handle(path string, handler http.Handler) Route
	HandleFunc(path string, handlerFunc http.HandlerFunc) Route
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}
