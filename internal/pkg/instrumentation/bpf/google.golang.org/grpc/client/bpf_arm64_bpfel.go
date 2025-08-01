// Code generated by bpf2go; DO NOT EDIT.
//go:build arm64

package grpc

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"structs"

	"github.com/cilium/ebpf"
)

type bpfGrpcRequestT struct {
	_          structs.HostLayout
	StartTime  uint64
	EndTime    uint64
	Sc         bpfSpanContext
	Psc        bpfSpanContext
	ErrMsg     [128]int8
	Method     [50]int8
	Target     [50]int8
	StatusCode uint32
}

type bpfSliceArrayBuff struct {
	_    structs.HostLayout
	Buff [1024]uint8
}

type bpfSpanContext struct {
	_          structs.HostLayout
	TraceID    [16]uint8
	SpanID     [8]uint8
	TraceFlags uint8
	Padding    [7]uint8
}

// loadBpf returns the embedded CollectionSpec for bpf.
func loadBpf() (*ebpf.CollectionSpec, error) {
	reader := bytes.NewReader(_BpfBytes)
	spec, err := ebpf.LoadCollectionSpecFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("can't load bpf: %w", err)
	}

	return spec, err
}

// loadBpfObjects loads bpf and converts it into a struct.
//
// The following types are suitable as obj argument:
//
//	*bpfObjects
//	*bpfPrograms
//	*bpfMaps
//
// See ebpf.CollectionSpec.LoadAndAssign documentation for details.
func loadBpfObjects(obj interface{}, opts *ebpf.CollectionOptions) error {
	spec, err := loadBpf()
	if err != nil {
		return err
	}

	return spec.LoadAndAssign(obj, opts)
}

// bpfSpecs contains maps and programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type bpfSpecs struct {
	bpfProgramSpecs
	bpfMapSpecs
	bpfVariableSpecs
}

// bpfProgramSpecs contains programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type bpfProgramSpecs struct {
	UprobeClientConnInvoke         *ebpf.ProgramSpec `ebpf:"uprobe_ClientConn_Invoke"`
	UprobeClientConnInvokeReturns  *ebpf.ProgramSpec `ebpf:"uprobe_ClientConn_Invoke_Returns"`
	UprobeLoopyWriterHeaderHandler *ebpf.ProgramSpec `ebpf:"uprobe_LoopyWriter_HeaderHandler"`
	UprobeHttp2ClientNewStream     *ebpf.ProgramSpec `ebpf:"uprobe_http2Client_NewStream"`
}

// bpfMapSpecs contains maps before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type bpfMapSpecs struct {
	AllocMap               *ebpf.MapSpec `ebpf:"alloc_map"`
	Events                 *ebpf.MapSpec `ebpf:"events"`
	GoContextToSc          *ebpf.MapSpec `ebpf:"go_context_to_sc"`
	GrpcEvents             *ebpf.MapSpec `ebpf:"grpc_events"`
	ProbeActiveSamplerMap  *ebpf.MapSpec `ebpf:"probe_active_sampler_map"`
	SamplersConfigMap      *ebpf.MapSpec `ebpf:"samplers_config_map"`
	SliceArrayBuffMap      *ebpf.MapSpec `ebpf:"slice_array_buff_map"`
	StreamidToSpanContexts *ebpf.MapSpec `ebpf:"streamid_to_span_contexts"`
	TrackedSpansBySc       *ebpf.MapSpec `ebpf:"tracked_spans_by_sc"`
}

// bpfVariableSpecs contains global variables before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type bpfVariableSpecs struct {
	ClientconnTargetPtrPos *ebpf.VariableSpec `ebpf:"clientconn_target_ptr_pos"`
	EndAddr                *ebpf.VariableSpec `ebpf:"end_addr"`
	ErrorStatusPos         *ebpf.VariableSpec `ebpf:"error_status_pos"`
	HeaderFrameHfPos       *ebpf.VariableSpec `ebpf:"headerFrame_hf_pos"`
	HeaderFrameStreamidPos *ebpf.VariableSpec `ebpf:"headerFrame_streamid_pos"`
	Hex                    *ebpf.VariableSpec `ebpf:"hex"`
	HttpclientNextidPos    *ebpf.VariableSpec `ebpf:"httpclient_nextid_pos"`
	StartAddr              *ebpf.VariableSpec `ebpf:"start_addr"`
	StatusCodePos          *ebpf.VariableSpec `ebpf:"status_code_pos"`
	StatusMessagePos       *ebpf.VariableSpec `ebpf:"status_message_pos"`
	StatusS_pos            *ebpf.VariableSpec `ebpf:"status_s_pos"`
	TotalCpus              *ebpf.VariableSpec `ebpf:"total_cpus"`
	WriteStatusSupported   *ebpf.VariableSpec `ebpf:"write_status_supported"`
}

// bpfObjects contains all objects after they have been loaded into the kernel.
//
// It can be passed to loadBpfObjects or ebpf.CollectionSpec.LoadAndAssign.
type bpfObjects struct {
	bpfPrograms
	bpfMaps
	bpfVariables
}

func (o *bpfObjects) Close() error {
	return _BpfClose(
		&o.bpfPrograms,
		&o.bpfMaps,
	)
}

// bpfMaps contains all maps after they have been loaded into the kernel.
//
// It can be passed to loadBpfObjects or ebpf.CollectionSpec.LoadAndAssign.
type bpfMaps struct {
	AllocMap               *ebpf.Map `ebpf:"alloc_map"`
	Events                 *ebpf.Map `ebpf:"events"`
	GoContextToSc          *ebpf.Map `ebpf:"go_context_to_sc"`
	GrpcEvents             *ebpf.Map `ebpf:"grpc_events"`
	ProbeActiveSamplerMap  *ebpf.Map `ebpf:"probe_active_sampler_map"`
	SamplersConfigMap      *ebpf.Map `ebpf:"samplers_config_map"`
	SliceArrayBuffMap      *ebpf.Map `ebpf:"slice_array_buff_map"`
	StreamidToSpanContexts *ebpf.Map `ebpf:"streamid_to_span_contexts"`
	TrackedSpansBySc       *ebpf.Map `ebpf:"tracked_spans_by_sc"`
}

func (m *bpfMaps) Close() error {
	return _BpfClose(
		m.AllocMap,
		m.Events,
		m.GoContextToSc,
		m.GrpcEvents,
		m.ProbeActiveSamplerMap,
		m.SamplersConfigMap,
		m.SliceArrayBuffMap,
		m.StreamidToSpanContexts,
		m.TrackedSpansBySc,
	)
}

// bpfVariables contains all global variables after they have been loaded into the kernel.
//
// It can be passed to loadBpfObjects or ebpf.CollectionSpec.LoadAndAssign.
type bpfVariables struct {
	ClientconnTargetPtrPos *ebpf.Variable `ebpf:"clientconn_target_ptr_pos"`
	EndAddr                *ebpf.Variable `ebpf:"end_addr"`
	ErrorStatusPos         *ebpf.Variable `ebpf:"error_status_pos"`
	HeaderFrameHfPos       *ebpf.Variable `ebpf:"headerFrame_hf_pos"`
	HeaderFrameStreamidPos *ebpf.Variable `ebpf:"headerFrame_streamid_pos"`
	Hex                    *ebpf.Variable `ebpf:"hex"`
	HttpclientNextidPos    *ebpf.Variable `ebpf:"httpclient_nextid_pos"`
	StartAddr              *ebpf.Variable `ebpf:"start_addr"`
	StatusCodePos          *ebpf.Variable `ebpf:"status_code_pos"`
	StatusMessagePos       *ebpf.Variable `ebpf:"status_message_pos"`
	StatusS_pos            *ebpf.Variable `ebpf:"status_s_pos"`
	TotalCpus              *ebpf.Variable `ebpf:"total_cpus"`
	WriteStatusSupported   *ebpf.Variable `ebpf:"write_status_supported"`
}

// bpfPrograms contains all programs after they have been loaded into the kernel.
//
// It can be passed to loadBpfObjects or ebpf.CollectionSpec.LoadAndAssign.
type bpfPrograms struct {
	UprobeClientConnInvoke         *ebpf.Program `ebpf:"uprobe_ClientConn_Invoke"`
	UprobeClientConnInvokeReturns  *ebpf.Program `ebpf:"uprobe_ClientConn_Invoke_Returns"`
	UprobeLoopyWriterHeaderHandler *ebpf.Program `ebpf:"uprobe_LoopyWriter_HeaderHandler"`
	UprobeHttp2ClientNewStream     *ebpf.Program `ebpf:"uprobe_http2Client_NewStream"`
}

func (p *bpfPrograms) Close() error {
	return _BpfClose(
		p.UprobeClientConnInvoke,
		p.UprobeClientConnInvokeReturns,
		p.UprobeLoopyWriterHeaderHandler,
		p.UprobeHttp2ClientNewStream,
	)
}

func _BpfClose(closers ...io.Closer) error {
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Do not access this directly.
//
//go:embed bpf_arm64_bpfel.o
var _BpfBytes []byte
