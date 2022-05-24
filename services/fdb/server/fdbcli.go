/* Copyright (c) 2019 Snowflake Inc. All rights reserved.

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

package server

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/Snowflake-Labs/sansshell/services"
	pb "github.com/Snowflake-Labs/sansshell/services/fdb"
	"github.com/Snowflake-Labs/sansshell/services/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	// FDBCLI is the location of the fdbcli binary. Binding this to a flag is often useful.
	FDBCLI string

	// FDBCLIUser is the user to become before running fdbcli (if non blank). Binding this to a flag is often useful.
	FDBCLIUser string

	// FDBCLIGroup is the group to become before running fdbcli (if non blank). Binding this to a flag is often useful.
	FDBCLIGroup string

	// FDBCLIEnvList ia a list of environment variables to retain before running fdbcli (such as TLS). Binding this to a flag is often useful.
	FDBCLIEnvList []string

	// generateFDBCLIArgs exists as a var for testing purposes
	generateFDBCLIArgs = generateFDBCLIArgsImpl
)

// server is used to implement the gRPC server
type server struct {
	mu        sync.Mutex
	fdbCLIUid int
	fdbCLIGid int
}

type creds struct {
	uid int
	gid int
}

// processCreds does a cached load of the uid/gid and is thread safe.
func (s *server) processCreds() (*creds, error) {
	c := &creds{s.fdbCLIUid, s.fdbCLIGid}
	s.mu.Lock()
	defer s.mu.Unlock()

	var usr, grp string
	if s.fdbCLIUid == -1 && FDBCLIUser != "" {
		usr = FDBCLIUser
	}
	if s.fdbCLIGid == -1 && FDBCLIGroup != "" {
		grp = FDBCLIGroup
	}

	if usr != "" {
		u, err := user.Lookup(usr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "unknown username %s: %v", usr, err)
		}
		id, err := strconv.Atoi(u.Uid)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "can't parse uid %s from lookup: %v", u.Uid, err)
		}
		s.fdbCLIUid, c.uid = id, id
	}
	if grp != "" {
		g, err := user.LookupGroup(grp)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "unknown group %s: %v", grp, err)
		}
		id, err := strconv.Atoi(g.Gid)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "can't parse gid %s from lookup: %v", g.Gid, err)
		}
		s.fdbCLIGid, c.gid = id, id
	}
	return c, nil
}

func (s *server) generateCommandOpts() ([]util.Option, error) {
	creds, err := s.processCreds()
	if err != nil {
		return nil, err
	}
	var opts []util.Option
	if creds.uid != -1 {
		opts = append(opts, util.CommandUser(uint32(creds.uid)))
	}
	if creds.gid != -1 {
		opts = append(opts, util.CommandGroup(uint32(creds.gid)))
	}
	return opts, nil
}

// captureLogs contains a context for determining a log file which may be emitted by a command
// to send back in the response as a Log entry.
type captureLogs struct {
	Path    string // Either a path to a file or a directory
	IsDir   bool   // Whether path is a directory (false means it's a file)
	Cleanup bool   // If true, os.RemoveAll on path
	Suffix  string // If a directory and this is not "" use as a suffix matcher for files to transmit
}

func parseLogs(logs []captureLogs) ([]*pb.Log, error) {
	var retLogs []*pb.Log
	// Parse captureLogs to see if there's something we need to transmit back.
	var files []string
	for _, l := range logs {
		if l.IsDir {
			dir := l.Path
			fs, err := os.ReadDir(dir)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "unable to list dir %s: %v", dir, err)
			}
			for _, f := range fs {
				if strings.HasSuffix(f.Name(), l.Suffix) {
					files = append(files, path.Join(dir, f.Name()))
				}
			}
			continue
		}
		files = append(files, l.Path)
	}
	for _, f := range files {
		retLogs = append(retLogs, &pb.Log{
			Filename: f,
		})
	}
	return retLogs, nil
}

func stringFlag(args []string, flag *wrapperspb.StringValue, value string) []string {
	if flag != nil {
		if value != "" {
			args = append(args, value)
		}
		args = append(args, flag.GetValue())
	}
	return args
}

func stringKVFlag(args []string, flag *wrapperspb.StringValue, value string) []string {
	a := stringFlag(nil, flag, value)
	if len(a) != 0 {
		args = append(args, strings.Join(a, "="))
	}
	return args
}

func uint32Flag(args []string, flag *wrapperspb.UInt32Value, value string) []string {
	if flag != nil {
		if value != "" {
			args = append(args, value)
		}
		args = append(args, fmt.Sprintf("%d", flag.GetValue()))
	}
	return args
}

func uint32KVFlag(args []string, flag *wrapperspb.UInt32Value, value string) []string {
	a := uint32Flag(nil, flag, value)
	if len(a) != 0 {
		args = append(args, strings.Join(a, "="))
	}
	return args
}

func int32Flag(args []string, flag *wrapperspb.Int32Value, value string) []string {
	if flag != nil {
		if value != "" {
			args = append(args, value)
		}
		args = append(args, fmt.Sprintf("%d", flag.GetValue()))
	}
	return args
}

func boolFlag(args []string, flag *wrapperspb.BoolValue, value string) []string {
	if flag != nil {
		// These are java flags and so they are false unless set. i.e. arity of 0 (no additional data).
		if flag.GetValue() {
			args = append(args, value)
		}
	}
	return args
}

func (s *server) FDBCLI(req *pb.FDBCLIRequest, stream pb.CLI_FDBCLIServer) error {
	if req.Request == nil {
		return status.Error(codes.InvalidArgument, "must fill in request")
	}

	command, logs, err := generateFDBCLIArgs(req)
	if err != nil {
		return err
	}
	opts, err := s.generateCommandOpts()
	if err != nil {
		return err
	}
	// Add env vars from flag
	for _, e := range FDBCLIEnvList {
		opts = append(opts, util.EnvVar(e))
	}

	run, err := util.RunCommand(stream.Context(), command[0], command[1:], opts...)
	if err != nil {
		return status.Errorf(codes.Internal, "error running fdbcli cmd (%+v): %v", command, err)
	}
	if run.Error != nil {
		return status.Errorf(codes.Internal, "can't exec fdbcli: %v", run.Error)
	}
	resp := &pb.FDBCLIResponse{
		Response: &pb.FDBCLIResponse_Output{
			Output: &pb.FDBCLIResponseOutput{
				RetCode: int32(run.ExitCode),
				Stderr:  run.Stderr.Bytes(),
				Stdout:  run.Stdout.Bytes(),
			},
		},
	}
	if err := stream.Send(resp); err != nil {
		return status.Errorf(codes.Internal, "can't send on stream: %v", err)
	}

	// Process any logs
	retLogs, err := parseLogs(logs)
	if err != nil {
		return err
	}
	// We get a list of filenames. So now we go over each and do chunk reads and send them back.
	for _, log := range retLogs {
		f, err := os.Open(log.Filename)
		if err != nil {
			return status.Errorf(codes.Internal, "can't open file %s: %v", log.Filename, err)
		}
		defer f.Close()
		buf := make([]byte, util.StreamingChunkSize)

		for {
			n, err := f.Read(buf)

			// If we got EOF we're done for this file
			if err == io.EOF {
				break
			}

			if err != nil {
				return status.Errorf(codes.Internal, "can't read file %s: %v", log.Filename, err)
			}
			// Only send over the number of bytes we actually read or
			// else we'll send over garbage in the last packet potentially.
			resp := &pb.FDBCLIResponse{
				Response: &pb.FDBCLIResponse_Log{
					Log: &pb.Log{
						Filename: log.Filename,
						Contents: buf[:n],
					},
				},
			}
			if err := stream.Send(resp); err != nil {
				return status.Errorf(codes.Internal, "can't send on stream for file %s: %v", log.Filename, err)
			}
		}
	}
	for _, l := range logs {
		if l.Cleanup {
			os.RemoveAll(l.Path)
		}
	}
	return nil
}

func generateFDBCLIArgsImpl(req *pb.FDBCLIRequest) ([]string, []captureLogs, error) {
	var logs []captureLogs
	topLevelArgs, l, err := generateFDBCLITopArgs(req)
	if err != nil {
		return nil, nil, err
	}
	logs = append(logs, l...)

	var args []string
	switch req.Request.(type) {
	case *pb.FDBCLIRequest_Command:
		args, l, err = parseFDBCommand(req.GetCommand())
	case *pb.FDBCLIRequest_Transaction:
		if len(req.GetTransaction().Commands) == 0 {
			return nil, nil, status.Error(codes.InvalidArgument, "transaction must have commands filled in")
		}
		args = append(args, "begin")
		args = append(args, ";")
		for _, c := range req.GetTransaction().Commands {
			a, l1, err := parseFDBCommand(c)
			if err != nil {
				return nil, nil, err
			}
			args = append(args, a...)
			l = append(l, l1...)
			args = append(args, ";")
		}
		args = append(args, "commit")
	default:
		return nil, nil, status.Errorf(codes.InvalidArgument, "unknown type %T for request", req.Request)
	}
	if err != nil {
		return nil, nil, err
	}
	logs = append(logs, l...)

	// fdbcli expects --exec to be passed one arg
	final := strings.Join(args, " ")
	args = []string{"--exec", final}

	command := []string{FDBCLI}
	command = append(command, topLevelArgs...)
	command = append(command, args...)
	return command, logs, nil
}

func generateFDBCLITopArgs(req *pb.FDBCLIRequest) ([]string, []captureLogs, error) {
	var args []string
	var logs []captureLogs

	args = stringFlag(args, req.Config, "--cluster-file")
	args = boolFlag(args, req.Log, "--log")
	if req.Log != nil && req.Log.Value {
		// Generate a captureLog
		dir, err := os.MkdirTemp("", "fdbcli")
		if err != nil {
			return nil, nil, status.Errorf(codes.Internal, "can't create tmp dir: %v", err)
		}
		logs = append(logs, captureLogs{
			Path:    dir,
			Cleanup: true,
			IsDir:   true,
		})

		args = append(args, "--log-dir", dir)
	}
	args = stringFlag(args, req.TraceFormat, "--trace_format")
	args = stringFlag(args, req.TlsCertificateFile, "--tls_certificate_file")
	args = stringFlag(args, req.TlsCaFile, "--tls_ca_file")
	args = stringFlag(args, req.TlsKeyFile, "--tls_key_file")
	args = stringFlag(args, req.TlsPassword, "--tls_password")
	args = stringFlag(args, req.TlsVerifyPeers, "--tls_verify_peers")
	args = boolFlag(args, req.DebugTls, "--debug-tls")
	args = boolFlag(args, req.Version, "--version")
	args = stringFlag(args, req.LogGroup, "--log-group")
	args = boolFlag(args, req.NoStatus, "--no-status")
	args = stringFlag(args, req.Memory, "--memory")
	args = boolFlag(args, req.BuildFlags, "--build-flags")
	args = int32Flag(args, req.Timeout, "--timeout")
	return args, logs, nil
}

func parseFDBCommand(req *pb.FDBCLICommand) (args []string, l []captureLogs, err error) {
	if req.Command == nil {
		return nil, nil, status.Error(codes.InvalidArgument, "command must be filled in")
	}
	switch req.Command.(type) {
	case *pb.FDBCLICommand_Advanceversion:
		args, l, err = parseFDBCLIAdvanceversion(req.GetAdvanceversion())
	case *pb.FDBCLICommand_Clear:
		args, l, err = parseFDBCLIClear(req.GetClear())
	case *pb.FDBCLICommand_Clearrange:
		args, l, err = parseFDBCLIClearrange((req.GetClearrange()))
	case *pb.FDBCLICommand_Configure:
		args, l, err = parseFDBCLIConfigure(req.GetConfigure())
	case *pb.FDBCLICommand_Consistencycheck:
		args, l, err = parseFDBCLIConsistencycheck(req.GetConsistencycheck())
	case *pb.FDBCLICommand_Coordinators:
		args, l, err = parseFDBCLICoordinators(req.GetCoordinators())
	case *pb.FDBCLICommand_Createtenant:
		args, l, err = parseFDBCLICreateTentant(req.GetCreatetenant())
	case *pb.FDBCLICommand_Defaulttenant:
		args, l, err = parseFDBCLIDefaulttenant(req.GetDefaulttenant())
	case *pb.FDBCLICommand_Deletetenant:
		args, l, err = parseFDBCLIDeleteTenant(req.GetDeletetenant())
	case *pb.FDBCLICommand_Exclude:
		args, l, err = parseFDBCLIExclude(req.GetExclude())
	case *pb.FDBCLICommand_Fileconfigure:
		args, l, err = parseFDBCLIFileconfigure(req.GetFileconfigure())
	case *pb.FDBCLICommand_ForceRecoveryWithDataLoss:
		args, l, err = parseFDBCLIForceRecoveryWithDataLoss(req.GetForceRecoveryWithDataLoss())
	case *pb.FDBCLICommand_Get:
		args, l, err = parseFDBCLIGet(req.GetGet())
	case *pb.FDBCLICommand_Getrange:
		args, l, err = parseFDBCLIGetrange(req.GetGetrange())
	case *pb.FDBCLICommand_Getrangekeys:
		args, l, err = parseFDBCLIGetrangekeys(req.GetGetrangekeys())
	case *pb.FDBCLICommand_Gettenant:
		args, l, err = parseFDBCLIGettenant(req.GetGettenant())
	case *pb.FDBCLICommand_Getversion:
		args, l, err = parseFDBCLIGetversion(req.GetGetversion())
	case *pb.FDBCLICommand_Help:
		args, l, err = parseFDBCLIHelp(req.GetHelp())
	case *pb.FDBCLICommand_Include:
		args, l, err = parseFDBCLIInclude(req.GetInclude())
	case *pb.FDBCLICommand_Kill:
		args, l, err = parseFDBCLIKill(req.GetKill())
	case *pb.FDBCLICommand_Listtenants:
		args, l, err = parseFDBCLIListtenants(req.GetListtenants())
	case *pb.FDBCLICommand_Lock:
		args, l, err = parseFDBCLILock(req.GetLock())
	case *pb.FDBCLICommand_Maintenance:
		args, l, err = parseFDBCLIMaintenance(req.GetMaintenance())
	case *pb.FDBCLICommand_Option:
		args, l, err = parseFDBCLIOption(req.GetOption())
	case *pb.FDBCLICommand_Profile:
		args, l, err = parseFDBCLIProfile(req.GetProfile())
	case *pb.FDBCLICommand_Set:
		args, l, err = parseFDBCLISet(req.GetSet())
	case *pb.FDBCLICommand_Setclass:
		args, l, err = parseFDBCLISetclass(req.GetSetclass())
	case *pb.FDBCLICommand_Sleep:
		args, l, err = parseFDBCLISleep(req.GetSleep())
	case *pb.FDBCLICommand_Status:
		args, l, err = parseFDBCLIStatus(req.GetStatus())
	case *pb.FDBCLICommand_Throttle:
		args, l, err = parseFDBCLIThrottle(req.GetThrottle())
	case *pb.FDBCLICommand_Triggerddteaminfolog:
		args, l, err = parseFDBCLITriggerddteaminfolog(req.GetTriggerddteaminfolog())
	case *pb.FDBCLICommand_Unlock:
		args, l, err = parseFDBCLIUnlock(req.GetUnlock())
	case *pb.FDBCLICommand_Usetenant:
		args, l, err = parseFDBCLIUseTenant(req.GetUsetenant())
	case *pb.FDBCLICommand_Writemode:
		args, l, err = parseFDBCLIWritemode(req.GetWritemode())
	case *pb.FDBCLICommand_Tssq:
		args, l, err = parseFDBCLITssq(req.GetTssq())
	default:
		return nil, nil, status.Errorf(codes.InvalidArgument, "unknown type %T for command", req.Command)
	}
	return
}

func parseFDBCLIAdvanceversion(req *pb.FDBCLIAdvanceversion) ([]string, []captureLogs, error) {
	return []string{"advanceversion", fmt.Sprintf("%d", req.Version)}, nil, nil
}

func parseFDBCLIClear(req *pb.FDBCLIClear) ([]string, []captureLogs, error) {
	if req.Key == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "clear requires key to be set")
	}
	return []string{"clear", req.Key}, nil, nil
}

func parseFDBCLIClearrange(req *pb.FDBCLIClearrange) ([]string, []captureLogs, error) {
	if req.BeginKey == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "clearrange requires begin_key to be set")
	}
	if req.EndKey == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "clearrange requires end_key to be set")
	}
	return []string{"clearrange", req.BeginKey, req.EndKey}, nil, nil
}

func parseFDBCLIConfigure(req *pb.FDBCLIConfigure) ([]string, []captureLogs, error) {
	args := []string{"configure"}

	// Technically one of these is required but we'll let the CLI complain in that case.
	args = stringFlag(args, req.NewOrTss, "")
	args = stringFlag(args, req.RedundancyMode, "")
	args = stringFlag(args, req.StorageEngine, "")
	args = uint32KVFlag(args, req.GrvProxies, "grv_proxies")
	args = uint32KVFlag(args, req.CommitProxies, "commit_proxies")
	args = uint32KVFlag(args, req.Resolvers, "resolvers")
	args = uint32KVFlag(args, req.Logs, "logs")
	args = uint32KVFlag(args, req.Count, "count")
	args = uint32KVFlag(args, req.PerpetualStorageWiggle, "perpetual_storage_wiggle")
	args = stringKVFlag(args, req.PerpetualStorageWiggleLocality, "perpetual_storage_wiggle_locality")
	args = stringKVFlag(args, req.StorageMigrationType, "storage_migration_type")
	args = stringKVFlag(args, req.TenantMode, "tenant_mode")
	return args, nil, nil
}

func parseFDBCLIConsistencycheck(req *pb.FDBCLIConsistencycheck) ([]string, []captureLogs, error) {
	args := []string{"consistencycheck"}

	// This one is special. If it's set we must set a value for both cases.
	// No value is also legit (returns status).
	if req.Mode != nil {
		if req.Mode.Value {
			args = append(args, "on")
		} else {
			args = append(args, "off")
		}
	}
	return args, nil, nil
}

func parseFDBCLICoordinators(req *pb.FDBCLICoordinators) ([]string, []captureLogs, error) {
	args := []string{"coordinators"}

	if req.Request == nil {
		return nil, nil, status.Errorf(codes.InvalidArgument, "coordinators requires request to be set")
	}
	if req.GetAuto() != nil {
		args = append(args, "auto")
	}
	if addr := req.GetAddresses(); addr != nil {
		args = append(args, addr.Addresses...)
	}
	args = stringKVFlag(args, req.Description, "description")
	return args, nil, nil
}

func parseFDBCLICreateTentant(req *pb.FDBCLICreatetenant) ([]string, []captureLogs, error) {
	args := []string{"createtenant"}

	if req.Name == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "createtenant must fill in name")
	}
	args = append(args, req.Name)
	return args, nil, nil
}

func parseFDBCLIDefaulttenant(req *pb.FDBCLIDefaulttenant) ([]string, []captureLogs, error) {
	return []string{"defaulttenant"}, nil, nil
}

func parseFDBCLIDeleteTenant(req *pb.FDBCLIDeletetenant) ([]string, []captureLogs, error) {
	args := []string{"deletetenant"}

	if req.Name == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "deletetenant must fill in name")
	}
	args = append(args, req.Name)
	return args, nil, nil
}

func parseFDBCLIExclude(req *pb.FDBCLIExclude) ([]string, []captureLogs, error) {
	args := []string{"exclude"}

	args = boolFlag(args, req.Failed, "failed")
	args = append(args, req.Addresses...)
	return args, nil, nil
}

func parseFDBCLIFileconfigure(req *pb.FDBCLIFileconfigure) ([]string, []captureLogs, error) {
	args := []string{"fileconfigure"}

	args = boolFlag(args, req.New, "new")

	if req.File == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "fileconfigure requires file to be filled in")
	}
	if err := util.ValidPath(req.File); err != nil {
		return nil, nil, err
	}
	args = append(args, req.File)
	return args, nil, nil
}

func parseFDBCLIForceRecoveryWithDataLoss(req *pb.FDBCLIForceRecoveryWithDataLoss) ([]string, []captureLogs, error) {
	args := []string{"force_recovery_with_data_loss"}

	if req.Dcid == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "force_recovery_with_data_loss requires dcid to be filled in")
	}
	args = append(args, req.Dcid)
	return args, nil, nil
}

func parseFDBCLIGet(req *pb.FDBCLIGet) ([]string, []captureLogs, error) {
	args := []string{"get"}

	if req.Key == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "get requires key to be filled in")
	}
	args = append(args, req.Key)
	return args, nil, nil
}

func parseFDBCLIGetrange(req *pb.FDBCLIGetrange) ([]string, []captureLogs, error) {
	args := []string{"getrange"}

	if req.BeginKey == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "getrange requires begin_key to be filled in")
	}
	args = append(args, req.BeginKey)
	args = stringFlag(args, req.EndKey, "")
	args = uint32Flag(args, req.Limit, "")
	return args, nil, nil
}

func parseFDBCLIGetrangekeys(req *pb.FDBCLIGetrangekeys) ([]string, []captureLogs, error) {
	args := []string{"getrangekeys"}

	if req.BeginKey == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "getrange requires begin_key to be filled in")
	}
	args = append(args, req.BeginKey)
	args = stringFlag(args, req.EndKey, "")
	args = uint32Flag(args, req.Limit, "")
	return args, nil, nil
}

func parseFDBCLIGettenant(req *pb.FDBCLIGettenant) ([]string, []captureLogs, error) {
	args := []string{"gettenant"}

	if req.Name == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "gettenant requires name to be filled in")
	}
	args = append(args, req.Name)
	return args, nil, nil
}

func parseFDBCLIGetversion(req *pb.FDBCLIGetversion) ([]string, []captureLogs, error) {
	return []string{"getversion"}, nil, nil
}

func parseFDBCLIHelp(req *pb.FDBCLIHelp) ([]string, []captureLogs, error) {
	args := []string{"help"}
	args = append(args, req.Options...)

	return args, nil, nil
}

func parseFDBCLIInclude(req *pb.FDBCLIInclude) ([]string, []captureLogs, error) {
	args := []string{"include"}

	if req.Request == nil {
		return nil, nil, status.Error(codes.InvalidArgument, "include requires request to be filled in")
	}
	args = boolFlag(args, req.Failed, "failed")

	switch req.Request.(type) {
	case *pb.FDBCLIInclude_All:
		args = append(args, "all")
	case *pb.FDBCLIInclude_Addresses:
		addr := req.GetAddresses().GetAddresses()
		if len(addr) == 0 {
			return nil, nil, status.Error(codes.InvalidArgument, "include without all requires addresses to be filled in")
		}
		args = append(args, addr...)
	}
	return args, nil, nil
}

func parseFDBCLIKill(req *pb.FDBCLIKill) ([]string, []captureLogs, error) {
	args := []string{"kill"}

	for _, a := range req.Addresses {
		args = append(args, fmt.Sprintf("; kill %s", a))
	}

	return args, nil, nil
}

func parseFDBCLIListtenants(req *pb.FDBCLIListtenants) ([]string, []captureLogs, error) {
	args := []string{"listtenants"}

	args = stringFlag(args, req.Begin, "")
	args = stringFlag(args, req.End, "")
	args = uint32Flag(args, req.Limit, "")

	return args, nil, nil
}

func parseFDBCLILock(req *pb.FDBCLILock) ([]string, []captureLogs, error) {
	return []string{"lock"}, nil, nil
}

func parseFDBCLIMaintenance(req *pb.FDBCLIMaintenance) ([]string, []captureLogs, error) {
	args := []string{"maintenance"}

	if req.Request == nil {
		return nil, nil, status.Error(codes.InvalidArgument, "maintenance requires request to be filled in")
	}

	switch req.Request.(type) {
	case *pb.FDBCLIMaintenance_Status:
		return args, nil, nil
	case *pb.FDBCLIMaintenance_On:
		on := req.GetOn()
		if on.Zoneid == "" {
			return nil, nil, status.Error(codes.InvalidArgument, "maintenance requires zoneid to be filled in if using on")
		}
		args = append(args, "on", on.Zoneid, fmt.Sprintf("%d", on.Seconds))
	case *pb.FDBCLIMaintenance_Off:
		args = append(args, "off")
	}
	return args, nil, nil
}

func parseFDBCLIOption(req *pb.FDBCLIOption) ([]string, []captureLogs, error) {
	args := []string{"option"}

	if req.Request == nil {
		return nil, nil, status.Error(codes.InvalidArgument, "option requires request to be filled in")
	}

	switch req.Request.(type) {
	case *pb.FDBCLIOption_Blank:
	// The easy case. Just want option with no args.
	case *pb.FDBCLIOption_Arg:
		arg := req.GetArg()
		if arg.State == "" {
			return nil, nil, status.Error(codes.InvalidArgument, "option when called with args requires state to be filled in")
		}
		if arg.Option == "" {
			return nil, nil, status.Error(codes.InvalidArgument, "option when called with args requires option to be filled in")
		}
		args = append(args, arg.State)
		args = append(args, arg.Option)
		args = stringFlag(args, arg.Arg, "")
	}
	return args, nil, nil
}

func parseFDBCLIProfile(req *pb.FDBCLIProfile) ([]string, []captureLogs, error) {
	args := []string{"profile"}
	var logs []captureLogs

	if req.Request == nil {
		return nil, nil, status.Error(codes.InvalidArgument, "profile requires request to be filled in")
	}
	switch req.Request.(type) {
	case *pb.FDBCLIProfile_Client:
		args = append(args, "client")
		client := req.GetClient()
		if client.Request == nil {
			return nil, nil, status.Errorf(codes.InvalidArgument, "profile client requires request to be filled in")
		}
		switch client.Request.(type) {
		case *pb.FDBCLIProfileActionClient_Get:
			args = append(args, "get")
		case *pb.FDBCLIProfileActionClient_Set:
			set := client.GetSet()
			if set.Rate == nil {
				return nil, nil, status.Errorf(codes.InvalidArgument, "profile client set requires rate to be filled in ")
			}
			if set.Size == nil {
				return nil, nil, status.Errorf(codes.InvalidArgument, "profile client set requires size to be filled in ")
			}

			args = append(args, "set")

			args = append(args, "rate")
			switch set.GetRate().(type) {
			case *pb.FDBCLIProfileActionClientSet_DefaultRate:
				args = append(args, "default")
			case *pb.FDBCLIProfileActionClientSet_ValueRate:
				args = append(args, fmt.Sprintf("%g", set.GetValueRate()))
			}

			args = append(args, "size")
			switch set.GetSize().(type) {
			case *pb.FDBCLIProfileActionClientSet_DefaultSize:
				args = append(args, "default")
			case *pb.FDBCLIProfileActionClientSet_ValueSize:
				args = append(args, fmt.Sprintf("%d", set.GetValueSize()))
			}
		}
	case *pb.FDBCLIProfile_List:
		args = append(args, "list")
	case *pb.FDBCLIProfile_Flow:
		flow := req.GetFlow()

		if len(flow.Processes) == 0 {
			return nil, nil, status.Error(codes.InvalidArgument, "profile flow must supply at least one process")
		}

		args = append(args, "flow")
		args = append(args, "run")
		args = append(args, fmt.Sprintf("%d", flow.Duration))

		// Create a log file for logging this.
		dir, err := os.MkdirTemp("", "fdbcli.profile")
		if err != nil {
			return nil, nil, status.Errorf(codes.Internal, "can't make tmp dir for profile: %v", err)
		}
		path := path.Join(dir, "profile.out")
		args = append(args, path)
		logs = append(logs, captureLogs{
			Path:    path,
			Cleanup: true,
		})

		args = append(args, flow.Processes...)
	case *pb.FDBCLIProfile_Heap:
		heap := req.GetHeap()
		if heap.Process == "" {
			return nil, nil, status.Error(codes.InvalidArgument, "profile heap requires process to be filled in")
		}
		args = append(args, "heap", heap.Process)
	}
	return args, logs, nil
}

func parseFDBCLISet(req *pb.FDBCLISet) ([]string, []captureLogs, error) {
	args := []string{"set"}

	if req.Key == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "set requires key to be filled in")
	}
	if req.Value == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "set requires value to be filled in")
	}

	args = append(args, req.Key, req.Value)

	return args, nil, nil
}

func parseFDBCLISetclass(req *pb.FDBCLISetclass) ([]string, []captureLogs, error) {
	args := []string{"setclass"}

	if req.Request == nil {
		return nil, nil, status.Error(codes.InvalidArgument, "setclass requires request to be filled in ")
	}
	switch req.Request.(type) {
	case *pb.FDBCLISetclass_List:
		return args, nil, nil
	case *pb.FDBCLISetclass_Arg:
		arg := req.GetArg()
		if arg.Address == "" {
			return nil, nil, status.Error(codes.InvalidArgument, "setclass requires address to be filled in when specifying arg")
		}
		if arg.Class == "" {
			return nil, nil, status.Error(codes.InvalidArgument, "setclass requires class to be filled in when specifying arg")
		}
		args = append(args, arg.Address, arg.Class)
	}
	return args, nil, nil
}

func parseFDBCLISleep(req *pb.FDBCLISleep) ([]string, []captureLogs, error) {
	args := []string{"sleep"}

	args = append(args, fmt.Sprintf("%d", req.Seconds))
	return args, nil, nil
}

func parseFDBCLIStatus(req *pb.FDBCLIStatus) ([]string, []captureLogs, error) {
	args := []string{"status"}

	args = stringFlag(args, req.Style, "")
	return args, nil, nil
}

func parseFDBCLIThrottle(req *pb.FDBCLIThrottle) ([]string, []captureLogs, error) {
	args := []string{"throttle"}

	if req.Request == nil {
		return nil, nil, status.Error(codes.InvalidArgument, "throttle requires request to be filled in")
	}
	switch req.Request.(type) {
	case *pb.FDBCLIThrottle_On:
		on := req.GetOn()
		if on.Tag == "" {
			return nil, nil, status.Error(codes.InvalidArgument, "throttle requires tag to be filled in when specifying on")
		}
		args = append(args, "on", "tag", on.Tag)
		args = uint32Flag(args, on.Rate, "")
		args = stringFlag(args, on.Duration, "")
		args = stringFlag(args, on.Priority, "")
	case *pb.FDBCLIThrottle_Off:
		off := req.GetOff()
		args = append(args, "off")
		args = stringFlag(args, off.Type, "")
		args = stringFlag(args, off.Tag, "tag")
		args = stringFlag(args, off.Priority, "")
	case *pb.FDBCLIThrottle_Enable:
		args = append(args, "enable", "auto")
	case *pb.FDBCLIThrottle_Disable:
		args = append(args, "disable", "auto")
	case *pb.FDBCLIThrottle_List:
		list := req.GetList()
		args = append(args, "list")
		args = stringFlag(args, list.Type, "")
		args = uint32Flag(args, list.Limit, "")
	}
	return args, nil, nil
}

func parseFDBCLITriggerddteaminfolog(req *pb.FDBCLITriggerddteaminfolog) ([]string, []captureLogs, error) {
	return []string{"triggerddteaminfolog"}, nil, nil
}

func parseFDBCLIUnlock(req *pb.FDBCLIUnlock) ([]string, []captureLogs, error) {
	args := []string{"unlock"}

	if req.Uid == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "unlock requires uid to be filled in")
	}
	args = append(args, req.Uid)
	return args, nil, nil
}

func parseFDBCLIUseTenant(req *pb.FDBCLIUsetenant) ([]string, []captureLogs, error) {
	args := []string{"usetenant"}

	if req.Name == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "usetenant requires name to be filled in")
	}
	args = append(args, req.Name)
	return args, nil, nil
}

func parseFDBCLIWritemode(req *pb.FDBCLIWritemode) ([]string, []captureLogs, error) {
	args := []string{"writemode"}

	if req.Mode == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "writemode requires mode to be filled in")
	}
	args = append(args, req.Mode)
	return args, nil, nil
}

func parseFDBCLITssq(req *pb.FDBCLITssq) ([]string, []captureLogs, error) {
	args := []string{"tssq"}

	if req.Request == nil {
		return nil, nil, status.Error(codes.InvalidArgument, "tssq requires request to be filled in")
	}

	switch req.Request.(type) {
	case *pb.FDBCLITssq_Start:
		start := req.GetStart()
		if start.StorageUid == "" {
			return nil, nil, status.Error(codes.InvalidArgument, "tssq requires storage_uid to be filled in when start is set")
		}
		args = append(args, "start", start.StorageUid)
	case *pb.FDBCLITssq_Stop:
		stop := req.GetStop()
		if stop.StorageUid == "" {
			return nil, nil, status.Error(codes.InvalidArgument, "tssq requires storage_uid to be filled in when stop is set")
		}
		args = append(args, "stop", stop.StorageUid)
	case *pb.FDBCLITssq_List:
		args = append(args, "list")
	}
	return args, nil, nil
}

// Register is called to expose this handler to the gRPC server
func (s *server) Register(gs *grpc.Server) {
	pb.RegisterCLIServer(gs, s)
}

func init() {
	services.RegisterSansShellService(&server{
		fdbCLIUid: -1,
		fdbCLIGid: -1,
	})
}