package interfaces

import (
	"errors"
	"regexp"
	"strings"

	"github.com/lwlcom/cisco_exporter/rpc"
	"github.com/lwlcom/cisco_exporter/util"
)

// Parse parses cli output and tries to find interfaces with related stats
func (c *interfaceCollector) Parse(ostype string, output string) ([]Interface, error) {
	if ostype != rpc.IOSXE && ostype != rpc.NXOS && ostype != rpc.IOS {
		return nil, errors.New("'show interface' is not implemented for " + ostype)
	}
	items := []Interface{}
	newIfRegexp, _ := regexp.Compile(`(?:^!?(?: |admin|show|.+#).*$|^$)`)
	macRegexp, _ := regexp.Compile(`^\s+Hardware(?: is|:) .+, address(?: is|:) (.*) \(.*\)$`)
	deviceNameRegexp, _ := regexp.Compile(`^([a-zA-Z0-9\/\.-]+) is.*$`)
	adminStatusRegexp, _ := regexp.Compile(`^.+ is (administratively)?\s*(up|down).*, line protocol is.*$`)
	adminStatusNXOSRegexp, _ := regexp.Compile(`^\S+ is (up|down)(?:\s|,)?(\(Administratively down\))?.*$`)
	descRegexp, _ := regexp.Compile(`^\s+Description: (.*)$`)
	dropsRegexp, _ := regexp.Compile(`^\s+Input queue: \d+\/\d+\/(\d+)\/\d+ .+ Total output drops: (\d+)$`)
	inputBytesRegexp, _ := regexp.Compile(`^\s+\d+ (?:packets input,|input packets)\s+(\d+) bytes.*$`)
	outputBytesRegexp, _ := regexp.Compile(`^\s+\d+ (?:packets output,|output packets)\s+(\d+) bytes.*$`)
	inputErrorsRegexp, _ := regexp.Compile(`^\s+(\d+) input error(?:s,)? .*$`)
	outputErrorsRegexp, _ := regexp.Compile(`^\s+(\d+) output error(?:s,)? .*$`)

	current := Interface{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if !newIfRegexp.MatchString(line) {
			if current != (Interface{}) {
				items = append(items, current)
			}
			matches := deviceNameRegexp.FindStringSubmatch(line)
			if matches == nil {
				continue
			}
			current = Interface{
				Name: matches[1],
			}
		}
		if current == (Interface{}) {
			continue
		}

		if matches := adminStatusRegexp.FindStringSubmatch(line); matches != nil {
			if matches[1] == "" {
				current.AdminStatus = "up"
			} else {
				current.AdminStatus = "down"
			}
			current.OperStatus = matches[2]
		} else if matches := adminStatusNXOSRegexp.FindStringSubmatch(line); matches != nil {
			if matches[2] == "" {
				current.AdminStatus = "up"
			} else {
				current.AdminStatus = "down"
			}
			current.OperStatus = matches[1]
		} else if matches := descRegexp.FindStringSubmatch(line); matches != nil {
			current.Description = matches[1]
		} else if matches := macRegexp.FindStringSubmatch(line); matches != nil {
			current.MacAddress = matches[1]
		} else if matches := dropsRegexp.FindStringSubmatch(line); matches != nil {
			current.InputDrops = util.Str2float64(matches[1])
			current.OutputDrops = util.Str2float64(matches[2])
		} else if matches := inputBytesRegexp.FindStringSubmatch(line); matches != nil {
			current.InputBytes = util.Str2float64(matches[1])
		} else if matches := outputBytesRegexp.FindStringSubmatch(line); matches != nil {
			current.OutputBytes = util.Str2float64(matches[1])
		} else if matches := inputErrorsRegexp.FindStringSubmatch(line); matches != nil {
			current.InputErrors = util.Str2float64(matches[1])
		} else if matches := outputErrorsRegexp.FindStringSubmatch(line); matches != nil {
			current.OutputErrors = util.Str2float64(matches[1])
		}
	}
	return append(items, current), nil
}

// ParseVlans parses cli output and tries to find vlans with related traffic stats
func (c *interfaceCollector) ParseVlans(ostype string, output string) ([]Interface, error) {
	if ostype != rpc.IOSXE {
		return nil, errors.New("'show vlans' is not implemented for " + ostype)
	}
	items := []Interface{}
	deviceNameRegexp, _ := regexp.Compile(`^([a-zA-Z0-9\/-]+\.[a-zA-Z0-9\/-]+) \(:?\d+\).*$`)
	inputBytesRegexp, _ := regexp.Compile(`^\s+Total \d+ packets, (\d+) bytes input.*$`)
	outputBytesRegexp, _ := regexp.Compile(`^\s+Total \d+ packets, (\d+) bytes output.*$`)

	current := Interface{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if matches := deviceNameRegexp.FindStringSubmatch(line); matches != nil {
			if current != (Interface{}) {
				items = append(items, current)
			}
			current = Interface{
				Name: matches[1],
			}
		}
		if current == (Interface{}) {
			continue
		}
		if matches := inputBytesRegexp.FindStringSubmatch(line); matches != nil {
			current.InputBytes = util.Str2float64(matches[1])
		} else if matches := outputBytesRegexp.FindStringSubmatch(line); matches != nil {
			current.OutputBytes = util.Str2float64(matches[1])
		}
	}
	return append(items, current), nil
}
