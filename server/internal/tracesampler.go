// Copyright 2019 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.opencensus.io/trace"
)

// Sampler constructs an object that decides how often to sample traces.
//
// The spec is a string in one of the forms:
//   * `X%` - to sample approximately X percent of requests.
//   * `Xqps` - to produce approximately X samples per second.
//
// Returns an error if the spec can't be parsed.
func Sampler(spec string) (trace.Sampler, error) {
	switch spec = strings.ToLower(spec); {
	case strings.HasSuffix(spec, "%"):
		percent, err := strconv.ParseFloat(strings.TrimSuffix(spec, "%"), 64)
		if err != nil {
			return nil, fmt.Errorf("not a float percent %q", spec)
		}
		// Note: ProbabilitySampler takes care of <=0.0 && >=1.0 cases.
		return trace.ProbabilitySampler(percent / 100.0), nil

	case strings.HasSuffix(spec, "qps"):
		qps, err := strconv.ParseFloat(strings.TrimSuffix(spec, "qps"), 64)
		if err != nil {
			return nil, fmt.Errorf("not a float QPS %q", spec)
		}
		if qps <= 0.0000001 {
			// Semantically the same, but slightly faster.
			return trace.ProbabilitySampler(0), nil
		}
		return (&qpsSampler{
			period: time.Duration(float64(time.Second) / qps),
			now:    time.Now,
		}).Sampler, nil

	default:
		return nil, fmt.Errorf("unrecognized sampling spec string %q - should be either 'X%%' or 'Xqps'", spec)
	}
}

// qpsSampler asks to sample a trace approximately each 'period'.
//
// TODO(vadimsh): Use a token bucket algorithm once we have a reusable
// implementation.
type qpsSampler struct {
	m      sync.RWMutex
	next   time.Time
	period time.Duration
	now    func() time.Time // for mocking time
}

func (s *qpsSampler) Sampler(p trace.SamplingParameters) trace.SamplingDecision {
	if p.ParentContext.IsSampled() {
		return trace.SamplingDecision{Sample: true}
	}

	now := s.now()

	s.m.RLock()
	sample := s.next.IsZero() || now.After(s.next)
	s.m.RUnlock()

	if sample {
		s.m.Lock()
		if sample = s.next.IsZero() || now.After(s.next); sample {
			s.next = now.Add(s.period)
		}
		s.m.Unlock()
	}

	return trace.SamplingDecision{Sample: sample}
}
