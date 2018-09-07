// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The protoc-gen-go binary is a protoc plugin to generate a Go protocol
// buffer package.
package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/golang/protobuf/proto"
	descpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"google.golang.org/proto/protogen"
)

func main() {
	protogen.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			genFile(gen, f)
		}
		return nil
	})
}

func genFile(gen *protogen.Plugin, f *protogen.File) {
	g := gen.NewGeneratedFile(f.GeneratedFilenamePrefix+".pb.go", f.GoImportPath)
	g.P("// Code generated by protoc-gen-go. DO NOT EDIT.")
	g.P("// source: ", f.Desc.Path())
	g.P()
	g.P("package ", f.GoPackageName)
	g.P()

	for _, m := range f.Messages {
		genMessage(gen, g, m)
	}

	genFileDescriptor(gen, g, f)
}

func genFileDescriptor(gen *protogen.Plugin, g *protogen.GeneratedFile, f *protogen.File) {
	// Determine the name of the var holding the file descriptor:
	//
	//     fileDescriptor_<hash of filename>
	filenameHash := sha256.Sum256([]byte(f.Desc.Path()))
	varName := fmt.Sprintf("fileDescriptor_%s", hex.EncodeToString(filenameHash[:8]))

	// Trim the source_code_info from the descriptor.
	// Marshal and gzip it.
	descProto := proto.Clone(f.Proto).(*descpb.FileDescriptorProto)
	descProto.SourceCodeInfo = nil
	b, err := proto.Marshal(descProto)
	if err != nil {
		gen.Error(err)
		return
	}
	var buf bytes.Buffer
	w, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	w.Write(b)
	w.Close()
	b = buf.Bytes()

	g.P("func init() { proto.RegisterFile(", strconv.Quote(f.Desc.Path()), ", ", varName, ") }")
	g.P()
	g.P("var ", varName, " = []byte{")
	g.P("// ", len(b), " bytes of a gzipped FileDescriptorProto")
	for len(b) > 0 {
		n := 16
		if n > len(b) {
			n = len(b)
		}

		s := ""
		for _, c := range b[:n] {
			s += fmt.Sprintf("0x%02x,", c)
		}
		g.P(s)

		b = b[n:]
	}
	g.P("}")
	g.P()
}

func genMessage(gen *protogen.Plugin, g *protogen.GeneratedFile, m *protogen.Message) {
	g.P("type ", m.GoIdent, " struct {")
	g.P("}")
	g.P()

	for _, nested := range m.Messages {
		genMessage(gen, g, nested)
	}
}
