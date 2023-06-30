/* Copyright (c) 2023 Snowflake Inc. All rights reserved.

   Licensed under the Apache License, Version 2.0 (the
   "License"); you may not use this file except in compliance
   with the License.  You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing,
   software distributed under the License is distributed on an
   "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
   KIND, either express or implied.  See the License for the
   specific language governing permissions and limitations
   under the License.
*/

// Package client provides the client interface for 'httpoverrpc'
package client

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/subcommands"

	"github.com/Snowflake-Labs/sansshell/client"
	pb "github.com/Snowflake-Labs/sansshell/services/httpoverrpc"
	"github.com/Snowflake-Labs/sansshell/services/util"
)

const subPackage = "httpoverrpc"

func init() {
	subcommands.Register(&httpCmd{}, subPackage)
}

func (*httpCmd) GetSubpackage(f *flag.FlagSet) *subcommands.Commander {
	c := client.SetupSubpackage(subPackage, f)
	c.Register(&proxyCmd{}, "")
	return c
}

type httpCmd struct{}

func (*httpCmd) Name() string { return subPackage }
func (p *httpCmd) Synopsis() string {
	return client.GenerateSynopsis(p.GetSubpackage(flag.NewFlagSet("", flag.ContinueOnError)), 2)
}
func (p *httpCmd) Usage() string {
	return client.GenerateUsage(subPackage, p.Synopsis())
}
func (*httpCmd) SetFlags(f *flag.FlagSet) {}

func (p *httpCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	c := p.GetSubpackage(f)
	return c.Execute(ctx, args...)
}

type proxyCmd struct {
	listenAddr string
}

func (*proxyCmd) Name() string     { return "proxy" }
func (*proxyCmd) Synopsis() string { return "Proxies a URL" }
func (*proxyCmd) Usage() string {
	return `proxy:
    Proxy things.
`
}

func (p *proxyCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.listenAddr, "addr", "localhost:0", "Address to listen on, defaults to a random localhost port")
}

// This context detachment is temporary until we use go1.21 and context.WithoutCancel is available.
type noCancel struct {
	ctx context.Context
}

func (c noCancel) Deadline() (time.Time, bool)       { return time.Time{}, false }
func (c noCancel) Done() <-chan struct{}             { return nil }
func (c noCancel) Err() error                        { return nil }
func (c noCancel) Value(key interface{}) interface{} { return c.ctx.Value(key) }

// WithoutCancel returns a context that is never canceled.
func WithoutCancel(ctx context.Context) context.Context {
	return noCancel{ctx: ctx}
}

func sendError(resp http.ResponseWriter, code int, err error) {
	resp.WriteHeader(code)
	if _, err := resp.Write([]byte(err.Error())); err != nil {
		fmt.Println(err)
	}
}

func (p *proxyCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	// Ignore the parent context timeout because we don't want to time out here.
	ctx = WithoutCancel(ctx)
	state := args[0].(*util.ExecuteState)
	if f.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "Please specify a URL to proxy.")
		return subcommands.ExitUsageError
	}
	if len(state.Out) != 1 {
		fmt.Fprintln(os.Stderr, "Proxying can only be done with exactly one target.")
		return subcommands.ExitUsageError
	}
	port, err := strconv.Atoi(f.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Port could not be interpreted as a number.")
		return subcommands.ExitUsageError
	}

	proxy := pb.NewHTTPOverRPCClientProxy(state.Conn)

	m := http.NewServeMux()
	m.HandleFunc("/", func(httpResp http.ResponseWriter, httpReq *http.Request) {
		var reqHeaders []*pb.Header
		for k, v := range httpReq.Header {
			reqHeaders = append(reqHeaders, &pb.Header{Key: k, Values: v})
		}
		body, err := io.ReadAll(httpReq.Body)
		if err != nil {
			sendError(httpResp, http.StatusBadRequest, err)
			return
		}
		req := &pb.HTTPRequest{
			RequestUri: httpReq.RequestURI,
			Method:     httpReq.Method,
			Headers:    reqHeaders,
			Body:       body,
			Port:       int32(port),
		}
		resp, err := proxy.Localhost(ctx, req)
		if err != nil {
			sendError(httpResp, http.StatusInternalServerError, err)
			return
		}
		for _, h := range resp.Headers {
			for _, v := range h.Values {
				httpResp.Header().Add(h.Key, v)
			}
		}
		httpResp.WriteHeader(int(resp.StatusCode))
		httpResp.Write(resp.Body)
	})
	l, err := net.Listen("tcp4", p.listenAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to listen on %v.\n", p.listenAddr)
		return subcommands.ExitFailure
	}
	fmt.Printf("Listening on http://%v, ctrl-c to exit", l.Addr())
	if err := http.Serve(l, m); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return subcommands.ExitUsageError
	}
	return subcommands.ExitSuccess
}
