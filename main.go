package main

import (
        "encoding/json"
        "fmt"
        "os"
        "sort"
        "strings"

        "sigs.k8s.io/kustomize/kyaml/fn/framework"
        "sigs.k8s.io/kustomize/kyaml/fn/framework/command"
        "sigs.k8s.io/kustomize/kyaml/yaml"
)

type Block struct {
        Name  string   `json:"name"`
        IPs   []string `json:"ips"`
        Ports []int    `json:"ports"`
}


func main() {
        fn := framework.ResourceListProcessorFunc(func(rl *framework.ResourceList) error {
                // 1. Get the 'data' field from functionConfig
                dataNode, err := rl.FunctionConfig.Pipe(yaml.Lookup("data"))
                if err != nil || dataNode == nil {
                        return fmt.Errorf("functionConfig.data is missing")
                }

                // 2. Get 'jsonPath' from inside the 'data' node
                jsonPathNode, err := dataNode.Pipe(yaml.Lookup("jsonPath"))
                if err != nil || jsonPathNode == nil {
                        return fmt.Errorf("jsonPath is missing in functionConfig.data")
                }


                // 3. Extract the value from the underlying yaml.Node
                // This gets the raw value without extra quotes
                jsonPath := strings.TrimSpace(jsonPathNode.YNode().Value)


                data, err := os.ReadFile(jsonPath)
                if err != nil {
                        return fmt.Errorf("failed to read json file at '%s': %w", jsonPath, err)
                }

                var blocks []Block
                if err := json.Unmarshal(data, &blocks); err != nil {
                        return fmt.Errorf("failed to unmarshal json: %w", err)
                }

                // Group IPs by Port
                portMap := make(map[int][]string)
                for _, b := range blocks {
                        for _, p := range b.Ports {
                                portMap[p] = append(portMap[p], b.IPs...)
                        }
                }

                // Generate Policies
                for port, ips := range portMap {
                        uniqueIPs := removeDuplicateStr(ips)
                        netPolYAML := generatePortSpecificNetPol(port, uniqueIPs)
                        node, err := yaml.Parse(netPolYAML)
                        if err != nil {
                                return err
                        }
                        rl.Items = append(rl.Items, node)
                }
                return nil
        })

        cmd := command.Build(fn, command.StandaloneEnabled, true)

        if err := cmd.Execute(); err != nil {
                fmt.Fprintf(os.Stderr, "%v\n", err)
                os.Exit(1)
        }
}

func generatePortSpecificNetPol(port int, ips []string) string {
        var ipBlocks []string
        for _, ip := range ips {
                ipBlocks = append(ipBlocks, fmt.Sprintf("    - ipBlock:\n        cidr: %s", ip))
        }

        return fmt.Sprintf(`
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: egress-allow-port-%d
spec:
  podSelector: {}
  policyTypes:
  - Egress
  egress:
  - to:
%s
    ports:
    - port: %d
      protocol: TCP`, port, strings.Join(ipBlocks, "\n"), port)
}

func removeDuplicateStr(strSlice []string) []string {
        allKeys := make(map[string]bool)
        list := []string{}
        for _, item := range strSlice {
                if _, value := allKeys[item]; !value {
                        allKeys[item] = true
                        list = append(list, item)
                }
        }
        sort.Strings(list)
        return list
}
