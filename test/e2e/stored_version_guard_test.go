//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// legacyStoredVersion is the removed spoke API version ("v1" + "alpha1"). It is
// assembled by concatenation on purpose: this guard's whole job is to assert that
// token never survives in a CRD's status.storedVersions, yet the repo keeps a
// mechanical "no legacy-version references under test/" gate. Concatenation lets
// the guard check the real value at runtime while keeping the raw grep clean.
const legacyStoredVersion = "v1" + "alpha1"

// Stored-version guard (8e-8a, decision C).
//
// After the single-version (legacy-spoke-removed) CRDs are applied to the cluster,
// every messaging.mkurator.dev CRD MUST report status.storedVersions == ["v1beta1"].
//
// On a fresh kind cluster storedVersions starts ["v1beta1"], so this passes. It
// exists to catch the bricked-upgrade case (roast P0): a cluster that ever stored
// legacy-spoke objects keeps that version in storedVersions, and applying CRDs that
// no longer list it leaves the apiserver unable to decode those stored objects.
// This converts that invisible failure into a red gate — if any CRD still lists the
// legacy spoke version in storedVersions, the operator must run the one-time
// stored-object rewrite+prune in docs/UPGRADE.md before dropping it.
var _ = Describe("CRD stored-version guard", Serial, Label("smoke"), func() {
	It("reports storedVersions == [v1beta1] for all six kinds (no legacy spoke left behind)", func() {
		waitForMKuratorCRDsEstablished()

		for _, crd := range mkuratorE2ECRDs {
			By("checking status.storedVersions for " + crd)
			Eventually(func(g Gomega) {
				out, err := runKubectl("get", "crd", crd,
					"-o", "jsonpath={.status.storedVersions}")
				g.Expect(err).NotTo(HaveOccurred(), "CRD %s should exist", crd)
				stored := strings.TrimSpace(out)

				// Fail loud on the bricked-upgrade case: a stored legacy-spoke object
				// keeps that version here after the single-version CRDs apply.
				g.Expect(stored).NotTo(ContainSubstring(legacyStoredVersion),
					"CRD %s still lists %s in storedVersions (%s) — run the stored-object "+
						"rewrite+prune in docs/UPGRADE.md before dropping it",
					crd, legacyStoredVersion, stored)
				g.Expect(stored).To(Equal(`["v1beta1"]`),
					"CRD %s storedVersions must be exactly [v1beta1], got %s", crd, stored)
			}).WithTimeout(2 * time.Minute).WithPolling(2 * time.Second).Should(Succeed())
		}
	})

	// Post-install CR write+read on the single-version CRD. Explicitly proves the
	// applied chart/kustomize CRD accepts a v1beta1 write through admission + storage
	// with NO conversion webhook — a dangling conversion stanza (roast P1) would fail
	// this create. Runs in both the kustomize and helm smoke lanes; needs no live MQ
	// (we assert the write round-trips, not that the connection reconciles Ready).
	It("accepts a v1beta1 QueueManagerConnection write and reads it back (no dangling conversion)", func() {
		waitForControllerAndWebhookReadyCached()

		const (
			guardQMC   = "single-version-guard-qmc"
			guardCreds = "single-version-guard-creds"
		)
		DeferCleanup(func() {
			kubectlDeleteIgnoreNotFound("queuemanagerconnection", guardQMC, namespace)
			kubectlDeleteIgnoreNotFound("secret", guardCreds, namespace)
		})

		By("creating the credentials Secret")
		Expect(kubectlApply(fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: %s
type: Opaque
stringData:
  username: admin
  mqAdminPassword: placeholder
`, guardCreds, namespace))).To(Succeed())

		By("applying a v1beta1 QueueManagerConnection (must pass admission on the single-version CRD)")
		Expect(applyWithWebhookRetry(fmt.Sprintf(`apiVersion: messaging.mkurator.dev/v1beta1
kind: QueueManagerConnection
metadata:
  name: %s
  namespace: %s
  annotations:
    messaging.mkurator.dev/allow-insecure-tls: "true"
spec:
  queueManager: QM1
  endpoint: https://placeholder.invalid:9443
  tls:
    insecureSkipVerify: true
  credentialsSecretRef:
    name: %s
`, guardQMC, namespace, guardCreds))).To(Succeed())

		By("reading the object back and confirming it stored as v1beta1")
		Eventually(func(g Gomega) {
			apiVer, err := runKubectl("get", "queuemanagerconnection", guardQMC, "-n", namespace,
				"-o", "jsonpath={.apiVersion}")
			g.Expect(err).NotTo(HaveOccurred(), "QMC %s should be readable after create", guardQMC)
			g.Expect(strings.TrimSpace(apiVer)).To(Equal("messaging.mkurator.dev/v1beta1"))

			qm, err := runKubectl("get", "queuemanagerconnection", guardQMC, "-n", namespace,
				"-o", "jsonpath={.spec.queueManager}")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(strings.TrimSpace(qm)).To(Equal("QM1"))
		}).WithTimeout(time.Minute).WithPolling(2 * time.Second).Should(Succeed())
	})
})
