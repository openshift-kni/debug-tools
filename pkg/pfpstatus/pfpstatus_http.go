/*
 * Copyright 2023 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pfpstatus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/go-logr/logr"

	authnv1 "k8s.io/api/authentication/v1"
	authzv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func serveLoop(ctx context.Context, env *environ, params HTTPParams) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /pfpstatus", func(w http.ResponseWriter, r *http.Request) {
		pfpstatusMainHandler(env, w, r)
	})
	mux.HandleFunc("GET /pfpstatus/{nodeName}", func(w http.ResponseWriter, r *http.Request) {
		pfpstatusNodeHandler(env, w, r)
	})

	var handle http.Handler = mux
	for _, mw := range params.Middlewares {
		env.lh.Info("Linking middleware", "name", mw.Name)
		handle = mw.Link(handle)
	}
	host := expandHost(params.Host)
	env.lh.Info("Starting PFP server", "host", host, "port", params.Port)
	addr := fmt.Sprintf("%s:%d", host, params.Port)
	err := http.ListenAndServe(addr, handle)
	if err != nil {
		env.lh.Error(err, "cannot serve PFP status")
	}
}

func expandHost(host string) string {
	if host == "" {
		return "0.0.0.0"
	}
	return host
}

func pfpstatusMainHandler(env *environ, w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Nodes int `json:"nodes"`
	}
	env.mu.Lock()
	resp := Response{
		Nodes: env.rec.CountNodes(),
	}
	env.mu.Unlock()

	sendContent(env.lh, w, &resp, "generic")
}

func pfpstatusNodeHandler(env *environ, w http.ResponseWriter, r *http.Request) {
	nodeName := r.PathValue("nodeName")
	if nodeName == "" {
		env.lh.Info("requested pfpstatus for empty node")
		http.Error(w, "missing node name", http.StatusUnprocessableEntity)
		return
	}
	env.mu.Lock()
	content, ok := env.rec.ContentForNode(nodeName)
	env.mu.Unlock()

	if !ok {
		http.Error(w, "unknown node name", http.StatusUnprocessableEntity)
		return
	}
	sendContent(env.lh, w, content, nodeName)
}

func sendContent(lh logr.Logger, w http.ResponseWriter, content any, endpoint string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(content)
	if err != nil {
		lh.Error(err, "sending back content for endpoint %q", endpoint)
	}
}

type TokenBearerAuthenticator struct {
	cs   kubernetes.Interface
	lh   logr.Logger
	next http.Handler
}

func (tba *TokenBearerAuthenticator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tba.lh.V(6).Info("auth middleware start")
	defer tba.lh.V(6).Info("auth middleware stop")

	if isRequestFromLoopback(tba.lh, r.RemoteAddr) {
		tba.lh.V(2).Info("auth bypass for loopback request", "remoteAddr", r.RemoteAddr)
		tba.next.ServeHTTP(w, r)
		return
	}

	err, code := canTokenBearerDebugPFPs(context.Background(), tba.cs, tba.lh, r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	tba.next.ServeHTTP(w, r)
}
func (tba *TokenBearerAuthenticator) WithLink(next http.Handler) *TokenBearerAuthenticator {
	tba.next = next
	return tba
}

func (tba *TokenBearerAuthenticator) Link(next http.Handler) http.Handler {
	tba.next = next
	return tba
}

func NewTokenBearerAuth(cs kubernetes.Interface, lh logr.Logger) *TokenBearerAuthenticator {
	return &TokenBearerAuthenticator{
		cs: cs,
		lh: lh,
	}
}

func isRequestFromLoopback(lh logr.Logger, remoteAddr string) bool {
	// Get the remote host IP, splitting off the port
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// this is so unexpected that we want to be loud
		lh.Error(err, "cannot parse address", "remoteAddr", remoteAddr)
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func canTokenBearerDebugPFPs(ctx context.Context, cs kubernetes.Interface, lh logr.Logger, authHeader string) (error, int) {
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return errors.New("Unauthorized: Missing Authorization Header"), http.StatusUnauthorized
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	tr, err := cs.AuthenticationV1().TokenReviews().Create(ctx, &authnv1.TokenReview{
		Spec: authnv1.TokenReviewSpec{
			Token: token,
		},
	}, metav1.CreateOptions{})

	if err != nil {
		lh.V(2).Error(err, "performing TokenReview")
		return fmt.Errorf("Internal Server Error: %w", err), http.StatusInternalServerError
	}

	if !tr.Status.Authenticated {
		lh.V(2).Info("token authentication failed")
		return errors.New("Unauthorized: Invalid token"), http.StatusUnauthorized
	}

	lh.V(4).Info("token authenticated", "user", tr.Status.User.Username)

	sar := &authzv1.SubjectAccessReview{
		Spec: authzv1.SubjectAccessReviewSpec{
			ResourceAttributes: &authzv1.ResourceAttributes{
				Group:    "topology.node.k8s.io",
				Resource: "noderesourcetopologies",
				Verb:     "sched.openshift-kni.io/debug",
			},
			User:   tr.Status.User.Username,
			Groups: tr.Status.User.Groups,
			Extra:  convertExtra(tr.Status.User.Extra),
		},
	}

	sarResult, err := cs.AuthorizationV1().SubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})
	if err != nil {
		lh.V(2).Error(err, "checking SubjectAccessReview")
		return fmt.Errorf("Internal Server Error: %w", err), http.StatusInternalServerError
	}

	if !sarResult.Status.Allowed {
		lh.V(2).Info("request denied", "username", tr.Status.User.Username, "reason", sarResult.Status.Reason)
		return errors.New("Forbidden: " + sarResult.Status.Reason), http.StatusForbidden
	}

	lh.V(4).Info("request allowed", "username", tr.Status.User.Username)
	return nil, http.StatusOK
}

func convertExtra(extra map[string]authnv1.ExtraValue) map[string]authzv1.ExtraValue {
	if extra == nil {
		return nil
	}
	newExtra := make(map[string]authzv1.ExtraValue)
	for k, v := range extra {
		newExtra[k] = authzv1.ExtraValue(v)
	}
	return newExtra
}
