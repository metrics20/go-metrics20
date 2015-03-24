package metrics20

import (
	"errors"
	"fmt"
	"strings"
)

type metricVersion int

const (
	Legacy      metricVersion = iota // bar.bytes or whatever
	M20                              // foo=bar.unit=B
	M20NoEquals                      // foo_is_bar.unit_is_B
)

func (version metricVersion) TagDelimiter() string {
	if version == M20 {
		return "="
	} else if version == M20NoEquals {
		return "_is_"
	}
	panic("TagDelimiter() called on metricVersion" + string(version))
}

// getVersion returns the expected version of a metric, but doesn't validate
func getVersion(metric_in string) metricVersion {
	if strings.Contains(metric_in, "=") {
		return M20
	}
	if strings.Contains(metric_in, "_is_") {
		return M20NoEquals
	}
	return Legacy
}

func IsMetric20(metric_in string) bool {
	v := getVersion(metric_in)
	return v == M20 || v == M20NoEquals
}

// InitialValidation checks the basic form of metric keys
func InitialValidation(metric_id string, version metricVersion) error {
	if version == Legacy {
		// if the metric contains no = or _is_, we don't really care what it does contain.  it can be whatever.
		// except for:
		if strings.Contains(metric_id, "..") {
			return fmt.Errorf("metric '%s' has an empty node", metric_id)
		}
		return nil
	}
	if version == M20 {
		if strings.Contains(metric_id, "_is_") {
			return fmt.Errorf("metric '%s' has both = and _is_", metric_id)
		}
		if !strings.HasPrefix(metric_id, "unit=") && !strings.Contains(metric_id, ".unit=") {
			return fmt.Errorf("metric '%s' has no unit tag", metric_id)
		}
		if !strings.HasPrefix(metric_id, "target_type=") && !strings.Contains(metric_id, ".target_type=") {
			return fmt.Errorf("metric '%s' has no target_type tag", metric_id)
		}
	} else { //version == M20NoEquals
		if strings.Contains(metric_id, "=") {
			return fmt.Errorf("metric '%s' has both = and _is_", metric_id)
		}
		if !strings.HasPrefix(metric_id, "unit_is_") && !strings.Contains(metric_id, ".unit_is_") {
			return fmt.Errorf("metric '%s' has no unit tag", metric_id)
		}
		if !strings.HasPrefix(metric_id, "target_type_is_") && !strings.Contains(metric_id, ".target_type_is_") {
			return fmt.Errorf("metric '%s' has no target_type tag", metric_id)
		}
	}
	if strings.Count(metric_id, ".") < 2 {
		return fmt.Errorf("metric '%s': must have at least one tag_k/tag_v pair beyond unit and target_type", metric_id)
	}
	return nil
}

type MetricSpec struct {
	Id   string
	Tags map[string]string
}

// NewMetricSpec takes a metric key, validates it (unit tag, etc) and
// converts it to a MetricSpec, setting nX tags, cleans up ps to /s unit
func NewMetricSpec(id string) (metric *MetricSpec, err error) {
	version := getVersion(id)
	err = InitialValidation(id, version)
	if err != nil {
		return nil, err
	}
	nodes := strings.Split(id, ".")
	del := version.TagDelimiter()
	tags := make(map[string]string)
	for i, node := range nodes {
		tag := strings.Split(node, del)
		if len(tag) > 2 {
			return nil, errors.New("bad metric spec: more than 1 equals")
		} else if len(tag) < 2 {
			tags[fmt.Sprintf("n%d", i+1)] = node
		} else if tag[0] == "" || tag[1] == "" {
			return nil, errors.New("bad metric spec: tag_k and tag_v must be non-empty strings")
		} else {
			// k=v format, and both are != ""
			key := tag[0]
			val := tag[1]
			if _, ok := tags[key]; ok {
				return nil, fmt.Errorf("duplicate tag key '%s'", key)
			}
			if key == "unit" && strings.HasSuffix(val, "ps") {
				val = val[:len(val)-2] + "/s"
			}
			tags[key] = val
		}
	}
	return &MetricSpec{id, tags}, nil
}

type MetricEs struct {
	Tags []string `json:"tags"`
}

func NewMetricEs(spec MetricSpec) MetricEs {
	tags := make([]string, len(spec.Tags), len(spec.Tags))
	i := 0
	for tag_key, tag_val := range spec.Tags {
		tags[i] = fmt.Sprintf("%s=%s", tag_key, tag_val)
		i++
	}
	return MetricEs{tags}
}
