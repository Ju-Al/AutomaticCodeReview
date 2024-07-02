/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This code embeds the CRD manifests in pkg/generated/crds/manifests in
// pkg/generated/crds/crds.go.

package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"text/template"
)

const goHeaderFile = "../../../hack/boilerplate.go.txt"

const tpl = `{{.GoHeader}}
// Code generated by crds_generate.go; DO NOT EDIT.

package crds

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"

	apiextinstall "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/install"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes/scheme"
)

var rawCRDs = [][]byte{
{{- range .RawCRDs }}
	[]byte({{ . }}),
{{- end }}
}

var CRDs = crds()

func crds() []*apiextv1beta1.CustomResourceDefinition {
	apiextinstall.Install(scheme.Scheme)
	decode := scheme.Codecs.UniversalDeserializer().Decode
	var objs []*apiextv1beta1.CustomResourceDefinition
	for _, crd := range rawCRDs {
		gzr, err := gzip.NewReader(bytes.NewReader(crd))
		if err != nil {
			panic(err)
		}
		bytes, err := ioutil.ReadAll(gzr)
		if err != nil {
			panic(err)
		}
		gzr.Close()

		obj, _, err := decode(bytes, nil, nil)
		if err != nil {
			panic(err)
		}
		objs = append(objs, obj.(*apiextv1beta1.CustomResourceDefinition))
	}
	return objs
}
`

type templateData struct {
	GoHeader string
	RawCRDs  []string
}

func main() {
	headerBytes, err := ioutil.ReadFile(goHeaderFile)
	if err != nil {
		log.Fatalln(err)
	}

	data := templateData{
		GoHeader: string(headerBytes),
	}

	manifests, err := ioutil.ReadDir("manifests")
	if err != nil {
		log.Fatalln(err)
	}

	for _, crd := range manifests {
		file, err := os.Open("manifests/" + crd.Name())
		if err != nil {
			log.Fatalln(err)
		}

		// gzip compress manifest
		var buf bytes.Buffer
		gzw := gzip.NewWriter(&buf)
		if _, err := io.Copy(gzw, file); err != nil {
			log.Fatalln(err)
		}
		file.Close()
		gzw.Close()

		data.RawCRDs = append(data.RawCRDs, fmt.Sprintf("%q", buf.Bytes()))
	}

	t, err := template.New("crds").Parse(tpl)
	if err != nil {
		log.Fatalln(err)
	}

	out, err := os.Create("crds.go")
	if err != nil {
		log.Fatalln(err)
	}

	if err := t.Execute(out, data); err != nil {
		log.Fatalln(err)
	}
}
