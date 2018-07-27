package llb

import (
	"github.com/moby/buildkit/solver/pb"
	"github.com/opencontainers/go-digest"
)

type ActionType int

const (
	ActionCopy ActionType = iota
	ActionMkFile
	ActionMkDir
	ActionRm
)

type FileActions struct {
	Type            ActionType
	InputFromSource Output
	InputFromTarget FileTarget
	Target          int
	MkDir           FileActionMkDir
	MkFile          FileActionMkFile
}

type FileTarget struct {
	Input  int
	Output int
}

type FileActionMkDir struct {
	path  string
	mode  int
	owner ChownOpt
}

type FileActionMkFile struct {
	path  string
	mode  int
	data  []byte
	owner ChownOpt
}

type ChownOpt struct {
}

type FileOp struct {
	MarshalCache
	input      Output
	output     Output
	fileTarget []FileTarget
	fileAction []FileActions
	//constraints Constraints
}

func NewFileOp(input Output, actions []FileActions) *FileOp {
	f := &FileOp{
		input:      input,
		fileAction: actions,
	}
	o := &output{vertex: f}
	f.output = o
	return f
}

type FileOption interface {
	SetFileOption(fa []FileActions)
}

type fileOptionFunc func(fa []FileActions)

func (fn fileOptionFunc) SetFileOption(fa []FileActions) {
	fn(fa)
}

//todo: propagate ChownOpt
func Mkdir(path string, mode int) FileOption {
	return fileOptionFunc(func(fa []FileActions) {
		mkdir := FileActionMkDir{path, mode, ChownOpt{}}
		fa = append(fa, FileActions{Type: ActionMkDir, MkDir: mkdir})
	})
}

func (f *FileOp) Validate() error {
	//todo: implement
	return nil
}

func (f *FileOp) Marshal(c *Constraints) (digest.Digest, []byte, *pb.OpMetadata, error) {
	if f.Cached(c) {
		return f.Load()
	}
	if err := f.Validate(); err != nil {
		return "", nil, nil, err
	}
	fo := &pb.FileOp{}
	proto, opMeta := MarshalConstraints(c, nil)
	for i, fa := range f.fileAction {
		inputIndex := pb.InputIndex(len(proto.Inputs))
		pfa := &pb.FileAction{}
		switch fa.Type {
		case ActionMkDir:
			pfa.Action = &pb.FileAction_Mkdir{
				Mkdir: &pb.FileActionMkDir{
					Path:  fa.MkDir.path,
					Mode:  uint32(fa.MkDir.mode),
					Owner: &pb.ChownOpt{}, //todo: propagate ChownOpt{}
				},
			}
		case ActionCopy:
		case ActionMkFile:
		case ActionRm:
		}

		fo.Actions = append(fo.Actions, pfa)
	}
	proto.Op = &pb.Op_File{
		File: fo,
	}

	dt, err := proto.Marshal()
	if err != nil {
		return "", nil, nil, err
	}
	f.Store(dt, opMeta, c)
	return f.Load()
}

func (f *FileOp) Output() Output {
	return f.output
}

func (f *FileOp) Inputs() (inputs []Output) {
	inputs = append(inputs, f.input)
	return
}
