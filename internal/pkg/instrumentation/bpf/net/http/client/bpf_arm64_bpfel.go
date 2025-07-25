// Code generated by bpf2go; DO NOT EDIT.
//go:build arm64

package client

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"structs"

	"github.com/cilium/ebpf"
)

type bpfHttpRequestT struct {
	_           structs.HostLayout
	StartTime   uint64
	EndTime     uint64
	Sc          bpfSpanContext
	Psc         bpfSpanContext
	Host        [128]int8
	Proto       [8]int8
	StatusCode  uint64
	Method      [16]int8
	Path        [128]int8
	Scheme      [8]int8
	Opaque      [8]int8
	RawPath     [8]int8
	Username    [8]int8
	RawQuery    [128]int8
	Fragment    [56]int8
	RawFragment [56]int8
	ForceQuery  uint8
	OmitHost    uint8
	_           [6]byte
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
	UprobeTransportRoundTrip        *ebpf.ProgramSpec `ebpf:"uprobe_Transport_roundTrip"`
	UprobeTransportRoundTripReturns *ebpf.ProgramSpec `ebpf:"uprobe_Transport_roundTrip_Returns"`
	UprobeWriteSubset               *ebpf.ProgramSpec `ebpf:"uprobe_writeSubset"`
}

// bpfMapSpecs contains maps before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type bpfMapSpecs struct {
	AllocMap                   *ebpf.MapSpec `ebpf:"alloc_map"`
	Events                     *ebpf.MapSpec `ebpf:"events"`
	GoContextToSc              *ebpf.MapSpec `ebpf:"go_context_to_sc"`
	HttpClientUprobeStorageMap *ebpf.MapSpec `ebpf:"http_client_uprobe_storage_map"`
	HttpEvents                 *ebpf.MapSpec `ebpf:"http_events"`
	HttpHeaders                *ebpf.MapSpec `ebpf:"http_headers"`
	ProbeActiveSamplerMap      *ebpf.MapSpec `ebpf:"probe_active_sampler_map"`
	SamplersConfigMap          *ebpf.MapSpec `ebpf:"samplers_config_map"`
	SliceArrayBuffMap          *ebpf.MapSpec `ebpf:"slice_array_buff_map"`
	TrackedSpansBySc           *ebpf.MapSpec `ebpf:"tracked_spans_by_sc"`
}

// bpfVariableSpecs contains global variables before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type bpfVariableSpecs struct {
	CtxPtrPos         *ebpf.VariableSpec `ebpf:"ctx_ptr_pos"`
	EndAddr           *ebpf.VariableSpec `ebpf:"end_addr"`
	ForceQueryPos     *ebpf.VariableSpec `ebpf:"force_query_pos"`
	FragmentPos       *ebpf.VariableSpec `ebpf:"fragment_pos"`
	HeadersPtrPos     *ebpf.VariableSpec `ebpf:"headers_ptr_pos"`
	Hex               *ebpf.VariableSpec `ebpf:"hex"`
	IoWriterBufPtrPos *ebpf.VariableSpec `ebpf:"io_writer_buf_ptr_pos"`
	IoWriterN_pos     *ebpf.VariableSpec `ebpf:"io_writer_n_pos"`
	MethodPtrPos      *ebpf.VariableSpec `ebpf:"method_ptr_pos"`
	OmitHostPos       *ebpf.VariableSpec `ebpf:"omit_host_pos"`
	OpaquePos         *ebpf.VariableSpec `ebpf:"opaque_pos"`
	PathPtrPos        *ebpf.VariableSpec `ebpf:"path_ptr_pos"`
	RawFragmentPos    *ebpf.VariableSpec `ebpf:"raw_fragment_pos"`
	RawPathPos        *ebpf.VariableSpec `ebpf:"raw_path_pos"`
	RawQueryPos       *ebpf.VariableSpec `ebpf:"raw_query_pos"`
	RequestHostPos    *ebpf.VariableSpec `ebpf:"request_host_pos"`
	RequestProtoPos   *ebpf.VariableSpec `ebpf:"request_proto_pos"`
	SchemePos         *ebpf.VariableSpec `ebpf:"scheme_pos"`
	StartAddr         *ebpf.VariableSpec `ebpf:"start_addr"`
	StatusCodePos     *ebpf.VariableSpec `ebpf:"status_code_pos"`
	TotalCpus         *ebpf.VariableSpec `ebpf:"total_cpus"`
	UrlHostPos        *ebpf.VariableSpec `ebpf:"url_host_pos"`
	UrlPtrPos         *ebpf.VariableSpec `ebpf:"url_ptr_pos"`
	UserPtrPos        *ebpf.VariableSpec `ebpf:"user_ptr_pos"`
	UsernamePos       *ebpf.VariableSpec `ebpf:"username_pos"`
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
	AllocMap                   *ebpf.Map `ebpf:"alloc_map"`
	Events                     *ebpf.Map `ebpf:"events"`
	GoContextToSc              *ebpf.Map `ebpf:"go_context_to_sc"`
	HttpClientUprobeStorageMap *ebpf.Map `ebpf:"http_client_uprobe_storage_map"`
	HttpEvents                 *ebpf.Map `ebpf:"http_events"`
	HttpHeaders                *ebpf.Map `ebpf:"http_headers"`
	ProbeActiveSamplerMap      *ebpf.Map `ebpf:"probe_active_sampler_map"`
	SamplersConfigMap          *ebpf.Map `ebpf:"samplers_config_map"`
	SliceArrayBuffMap          *ebpf.Map `ebpf:"slice_array_buff_map"`
	TrackedSpansBySc           *ebpf.Map `ebpf:"tracked_spans_by_sc"`
}

func (m *bpfMaps) Close() error {
	return _BpfClose(
		m.AllocMap,
		m.Events,
		m.GoContextToSc,
		m.HttpClientUprobeStorageMap,
		m.HttpEvents,
		m.HttpHeaders,
		m.ProbeActiveSamplerMap,
		m.SamplersConfigMap,
		m.SliceArrayBuffMap,
		m.TrackedSpansBySc,
	)
}

// bpfVariables contains all global variables after they have been loaded into the kernel.
//
// It can be passed to loadBpfObjects or ebpf.CollectionSpec.LoadAndAssign.
type bpfVariables struct {
	CtxPtrPos         *ebpf.Variable `ebpf:"ctx_ptr_pos"`
	EndAddr           *ebpf.Variable `ebpf:"end_addr"`
	ForceQueryPos     *ebpf.Variable `ebpf:"force_query_pos"`
	FragmentPos       *ebpf.Variable `ebpf:"fragment_pos"`
	HeadersPtrPos     *ebpf.Variable `ebpf:"headers_ptr_pos"`
	Hex               *ebpf.Variable `ebpf:"hex"`
	IoWriterBufPtrPos *ebpf.Variable `ebpf:"io_writer_buf_ptr_pos"`
	IoWriterN_pos     *ebpf.Variable `ebpf:"io_writer_n_pos"`
	MethodPtrPos      *ebpf.Variable `ebpf:"method_ptr_pos"`
	OmitHostPos       *ebpf.Variable `ebpf:"omit_host_pos"`
	OpaquePos         *ebpf.Variable `ebpf:"opaque_pos"`
	PathPtrPos        *ebpf.Variable `ebpf:"path_ptr_pos"`
	RawFragmentPos    *ebpf.Variable `ebpf:"raw_fragment_pos"`
	RawPathPos        *ebpf.Variable `ebpf:"raw_path_pos"`
	RawQueryPos       *ebpf.Variable `ebpf:"raw_query_pos"`
	RequestHostPos    *ebpf.Variable `ebpf:"request_host_pos"`
	RequestProtoPos   *ebpf.Variable `ebpf:"request_proto_pos"`
	SchemePos         *ebpf.Variable `ebpf:"scheme_pos"`
	StartAddr         *ebpf.Variable `ebpf:"start_addr"`
	StatusCodePos     *ebpf.Variable `ebpf:"status_code_pos"`
	TotalCpus         *ebpf.Variable `ebpf:"total_cpus"`
	UrlHostPos        *ebpf.Variable `ebpf:"url_host_pos"`
	UrlPtrPos         *ebpf.Variable `ebpf:"url_ptr_pos"`
	UserPtrPos        *ebpf.Variable `ebpf:"user_ptr_pos"`
	UsernamePos       *ebpf.Variable `ebpf:"username_pos"`
}

// bpfPrograms contains all programs after they have been loaded into the kernel.
//
// It can be passed to loadBpfObjects or ebpf.CollectionSpec.LoadAndAssign.
type bpfPrograms struct {
	UprobeTransportRoundTrip        *ebpf.Program `ebpf:"uprobe_Transport_roundTrip"`
	UprobeTransportRoundTripReturns *ebpf.Program `ebpf:"uprobe_Transport_roundTrip_Returns"`
	UprobeWriteSubset               *ebpf.Program `ebpf:"uprobe_writeSubset"`
}

func (p *bpfPrograms) Close() error {
	return _BpfClose(
		p.UprobeTransportRoundTrip,
		p.UprobeTransportRoundTripReturns,
		p.UprobeWriteSubset,
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
