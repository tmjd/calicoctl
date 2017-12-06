// Copyright (c) 2016 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/docopt/docopt-go"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/projectcalico/calicoctl/calicoctl/commands/argutils"
	"github.com/projectcalico/calicoctl/calicoctl/commands/constants"
	"github.com/projectcalico/calicoctl/calicoctl/commands/v1resourceloader"
	"github.com/projectcalico/libcalico-go/lib/apis/v1/unversioned"
	conversion "github.com/projectcalico/libcalico-go/lib/upgrade/etcd/conversionv1v3"
)

func Convert(args []string) {
	doc := constants.DatastoreIntro + `Usage:
  calicoctl convert --filename=<FILENAME>
                [--output=<OUTPUT>]

Examples:
  # Create a policy using the data in policy.yaml.
  calicoctl convert -f ./policy.yaml -o yaml

  # Create a policy based on the JSON passed into stdin.
  cat policy.json | calicoctl convert -f -

Options:
  -h --help                     Show this screen.
  -f --filename=<FILENAME>      Filename to use to create the resource.  If set to
                                "-" loads from stdin.
  -o --output=<OUTPUT FORMAT>   Output format. One of: yaml or json.
                                [Default: yaml]


Description:
  Convert config files between different API versions. Both YAML and JSON formats are accepted.

  The default output will be printed to stdout in YAML format.
`
	parsedArgs, err := docopt.Parse(doc, args, true, "", false, false)
	if err != nil {
		fmt.Printf("Invalid option: 'calicoctl %s'. Use flag '--help' to read about a specific subcommand.\n", strings.Join(args, " "))
		os.Exit(1)
	}
	if len(parsedArgs) == 0 {
		return
	}

	var rp resourcePrinter
	output := parsedArgs["--output"].(string)
	// Only supported output formats are yaml (default) and json.
	switch output {
	case "yaml", "yml":
		rp = resourcePrinterYAML{}
	case "json":
		rp = resourcePrinterJSON{}
	default:
		rp = nil
	}

	if rp == nil {
		fmt.Printf("unrecognized output format '%s'\n", output)
		os.Exit(1)
	}

	filename := argutils.ArgStringOrBlank(parsedArgs, "--filename")

	// Load the V1 resource from file and convert to a slice
	// of resources for easier handling.
	resV1, err := v1resourceloader.CreateResourcesFromFile(filename)
	if err != nil {
		fmt.Printf("Failed to execute command: %v\n", err)
		os.Exit(1)
	}

	var results []runtime.Object
	for _, v1Resource := range resV1 {
		v3Resource, err := convertResource(v1Resource)
		if err != nil {
			fmt.Printf("Failed to execute command: %v\n", err)
			os.Exit(1)
		}
		results = append(results, v3Resource)
	}

	log.Infof("results: %+v", results)

	err = rp.print(nil, results)
	if err != nil {
		fmt.Println(err)
	}
}

func convertResource(v1resource unversioned.Resource) (runtime.Object, error) {
	switch strings.ToLower(v1resource.GetTypeMetadata().Kind) {
	case "node":
		return convert(conversion.Node{}, v1resource)
	case "hostendpoint":
		return convert(conversion.HostEndpoint{}, v1resource)
	case "workloadendpoint":
		return convert(conversion.WorkloadEndpoint{}, v1resource)
	case "profile":
		return convert(conversion.Profile{}, v1resource)
	case "policy":
		return convert(conversion.Policy{}, v1resource)
	case "ippool":
		return convert(conversion.IPPool{}, v1resource)
	case "bgppeer":
		return convert(conversion.BGPPeer{}, v1resource)

	default:
		return nil, fmt.Errorf("conversion for the resource type '%s' is not supported", v1resource.GetTypeMetadata().Kind)
	}
}

func convert(convRes conversion.Converter, v1resource unversioned.Resource) (conversion.Resource, error) {
	// Convert v1 API resource to v1 backend KVPair.
	kvp, err := convRes.APIV1ToBackendV1(v1resource)
	if err != nil {
		return nil, err
	}

	// Convert v1 backend KVPair to v3 API resource.
	res, err := convRes.BackendV1ToAPIV3(kvp)
	if err != nil {
		return nil, err
	}

	return res, nil
}
