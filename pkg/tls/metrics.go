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

package tls

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/kuadrant/kcp-glbc/pkg/metrics"
)

const (
	issuerLabel = "issuer"
)

var (
	// CertificateRequestCount is a prometheus metric which holds the number of
	// pending TLS certificate requests.
	// TODO: make it package private once the certificate management is better encapsulated
	CertificateRequestCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "glbc_tls_certificate_pending_request_count",
			Help: "GLBC TLS certificate pending request count",
		},
		[]string{
			issuerLabel,
		},
	)
)

func init() {
	// Register metrics with the global prometheus registry
	metrics.Registry.MustRegister(
		CertificateRequestCount,
	)
}

func InitMetrics(provider Provider) {
	// Initialize metrics
	issuer := provider.IssuerID()
	CertificateRequestCount.WithLabelValues(issuer).Set(0)
}
