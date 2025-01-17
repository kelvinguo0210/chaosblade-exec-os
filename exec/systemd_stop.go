/*
 * Copyright 1999-2020 Alibaba Group Holding Ltd.
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

 package exec

 import (
	 "context"
	 "fmt"
	 "path"
 
	 "github.com/chaosblade-io/chaosblade-spec-go/channel"
	 "github.com/chaosblade-io/chaosblade-spec-go/spec"
	 "github.com/chaosblade-io/chaosblade-spec-go/util"
	 "github.com/chaosblade-io/chaosblade-exec-os/exec/category"
 )
 
 const StopSystemdBin = "chaos_stopsystemd"
 
 type StopSystemdActionCommandSpec struct {
	 spec.BaseExpActionCommandSpec
 }
 
 func NewStopSystemdActionCommandSpec() spec.ExpActionCommandSpec {
	 return &StopSystemdActionCommandSpec{
		 spec.BaseExpActionCommandSpec{
			 ActionMatchers: []spec.ExpFlagSpec{
				 &spec.ExpFlag{
					 Name: "service",
					 Desc: "Service name",
				 },
			 },
			 ActionFlags:    []spec.ExpFlagSpec{},
			 ActionExecutor: &StopSystemdExecutor{},
			 ActionExample: `
 # Stop the service test
 blade create systemd stop --service test`,
			 ActionPrograms:   []string{StopSystemdBin},
			 ActionCategories: []string{category.SystemSystemd},
		 },
	 }
 }
 
 func (*StopSystemdActionCommandSpec) Name() string {
	 return "stop"
 }
 
 func (*StopSystemdActionCommandSpec) Aliases() []string {
	 return []string{"s"}
 }
 
 func (*StopSystemdActionCommandSpec) ShortDesc() string {
	 return "Stop systemd"
 }
 
 func (k *StopSystemdActionCommandSpec) LongDesc() string {
	 if k.ActionLongDesc != "" {
		 return k.ActionLongDesc
	 }
	 return "Stop system by service name"
 }
 
 func (*StopSystemdActionCommandSpec) Categories() []string {
	 return []string{category.SystemSystemd}
 }
 
 type StopSystemdExecutor struct {
	 channel spec.Channel
 }
 
 func (sse *StopSystemdExecutor) Name() string {
	 return "stop"
 }
 
 func (sse *StopSystemdExecutor) Exec(uid string, ctx context.Context, model *spec.ExpModel) *spec.Response {
	 if sse.channel == nil {
		 util.Errorf(uid, util.GetRunFuncName(), spec.ResponseErr[spec.ChannelNil].ErrInfo)
		 return spec.ResponseFail(spec.ChannelNil, spec.ResponseErr[spec.ChannelNil].ErrInfo)
	 }
	 service := model.ActionFlags["service"]
	 if service == "" {
		 util.Errorf(uid, util.GetRunFuncName(), "service name is mandantory")
		 return spec.ResponseFailWaitResult(spec.ParameterLess, fmt.Sprintf(spec.ResponseErr[spec.ParameterLess].Err, "service"),
			 fmt.Sprintf(spec.ResponseErr[spec.ParameterLess].ErrInfo, "service"))
	 }
 

	 flags := fmt.Sprintf("--service %s", service)
	 if _, ok := spec.IsDestroy(ctx); ok {
		return sse.startService(service, ctx)
	} else {
		if response := checkServiceInvalid(uid, service, ctx); response != nil {
	 	    return response
	    }
		return sse.channel.Run(ctx, path.Join(sse.channel.GetScriptPath(), StopSystemdBin), flags)
	}
 }

 func checkServiceInvalid(uid, service string, ctx context.Context) *spec.Response {
	cl := channel.NewLocalChannel()
	if !cl.IsCommandAvailable("systemctl") {
		util.Errorf(uid, util.GetRunFuncName(), spec.ResponseErr[spec.CommandSystemctlNotFound].Err)
		return spec.ResponseFail(spec.CommandSystemctlNotFound, spec.ResponseErr[spec.SystemdNotFound].Err,
			spec.ResponseErr[spec.SystemdNotFound].ErrInfo)
	}
	response := cl.Run(ctx, "systemctl", fmt.Sprintf(`status "%s" | grep 'Active' | grep 'running'`, service))
	if !response.Success {
		util.Errorf(uid, util.GetRunFuncName(), fmt.Sprintf(spec.ResponseErr[spec.SystemdNotFound].Err, service))
		return spec.ResponseFail(spec.SystemdNotFound, fmt.Sprintf(spec.ResponseErr[spec.SystemdNotFound].Err, service),
			fmt.Sprintf(spec.ResponseErr[spec.SystemdNotFound].ErrInfo, service, response.Err))
	}
	return nil
}

func (sse *StopSystemdExecutor) startService(service string, ctx context.Context) *spec.Response {
	return sse.channel.Run(ctx, "systemctl", fmt.Sprintf("start %s", service))
}
 
 func (sse *StopSystemdExecutor) SetChannel(channel spec.Channel) {
	 sse.channel = channel
 }
 