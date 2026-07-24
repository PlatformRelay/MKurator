//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"

	"github.com/platformrelay/mkurator/test/utils"
)

func e2eVerboseLogs() bool {
	return os.Getenv("KURATOR_E2E_VERBOSE_LOGS") == "1"
}

// e2eStage prints a ci_step-style banner to stdout (visible outside GinkgoWriter).
func e2eStage(stage string) {
	_, _ = fmt.Fprintf(os.Stdout, "\n==> %s %s\n\n", time.Now().Format(time.RFC3339), stage)
}

// e2eBy mirrors Ginkgo By() and echoes a short progress line to stdout.
func e2eBy(msg string) {
	_, _ = fmt.Fprintf(os.Stdout, "[e2e] %s\n", msg)
	By(msg)
}

// e2eSpecLine prints spec start/end markers when ReportBeforeEach/AfterEach fire.
func e2eSpecLine(state, fullText string) {
	_, _ = fmt.Fprintf(os.Stdout, "[e2e] %s %s\n", state, fullText)
}

// dumpHelmDeployFailureDiagnostics captures the manager rollout state at the moment the
// Helm same-cluster deploy fails, before the AfterSuite tears the cluster state down. It is a
// temporary diagnostic to discriminate why the controller-manager never reaches Ready on the
// v1beta1 storage flip: conversion-webhook/TLS failure vs leftover not-Ready CRs vs pod crash.
func dumpHelmDeployFailureDiagnostics() {
	e2eBy("collecting Helm deploy failure diagnostics")

	dump := func(label string, args ...string) {
		cmd := exec.Command("kubectl", args...)
		out, err := utils.Run(cmd)
		if err != nil {
			_, _ = fmt.Fprintf(GinkgoWriter, "==DIAG %s== error: %s\n%s\n", label, err, out)
			return
		}
		_, _ = fmt.Fprintf(GinkgoWriter, "==DIAG %s==\n%s\n", label, out)
	}

	dump("pods", "get", "pods", "-n", namespace, "-o", "wide")
	dump("describe-deploy", "describe", "deployment", "mkurator-controller-manager", "-n", namespace)
	dump("manager-logs", "logs", "deployment/mkurator-controller-manager", "-n", namespace, "--tail=120")
	dump("manager-logs-prev", "logs", "deployment/mkurator-controller-manager", "-n", namespace, "--previous", "--tail=120")
	dump("events", "get", "events", "-n", namespace, "--sort-by=.lastTimestamp")

	// The single best discriminator: does listing the non-storage (v1alpha1) version across all
	// namespaces error (conversion/TLS), return leftover objects, or return empty?
	for _, kind := range mkuratorE2ECRDs {
		plural := kind[:len(kind)-len(".messaging.mkurator.dev")]
		dump("list-v1alpha1-"+plural, "get", plural+".v1alpha1.messaging.mkurator.dev", "-A")
		dump("list-v1beta1-"+plural, "get", plural+".v1beta1.messaging.mkurator.dev", "-A")
	}

	// CRD storedVersions + conversion caBundle presence for one representative kind.
	dump("crd-storedversions", "get", "crd", "queuemanagerconnections.messaging.mkurator.dev",
		"-o", "jsonpath={.status.storedVersions}{\"\\n\"}{.spec.conversion.webhook.clientConfig.caBundle}")
}

// dumpFailureDiagnostics collects kubectl context on spec failure without full JSON logs by default.
func dumpFailureDiagnostics(controllerPodName string) {
	e2eBy("collecting failure diagnostics")

	if controllerPodName != "" {
		args := []string{"logs", controllerPodName, "-n", namespace}
		if !e2eVerboseLogs() {
			args = append(args, "--tail=40")
		}
		cmd := exec.Command("kubectl", args...)
		controllerLogs, err := utils.Run(cmd)
		if err == nil {
			hint := "last 40 lines"
			if e2eVerboseLogs() {
				hint = "full log"
			}
			_, _ = fmt.Fprintf(GinkgoWriter, "Controller logs (%s; KURATOR_E2E_VERBOSE_LOGS=1 for full dump):\n%s",
				hint, controllerLogs)
		} else {
			_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get controller logs: %s\n", err)
		}
	}

	cmd := exec.Command("kubectl", "get", "events", "-n", namespace, "--sort-by=.lastTimestamp")
	eventsOutput, err := utils.Run(cmd)
	if err == nil {
		_, _ = fmt.Fprintf(GinkgoWriter, "Kubernetes events:\n%s", eventsOutput)
	} else {
		_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get Kubernetes events: %s\n", err)
	}

	cmd = exec.Command("kubectl", "logs", "curl-metrics", "-n", namespace, "--tail=20")
	metricsOutput, err := utils.Run(cmd)
	if err == nil {
		_, _ = fmt.Fprintf(GinkgoWriter, "Metrics curl logs (last 20 lines):\n%s", metricsOutput)
	} else if e2eVerboseLogs() {
		_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get curl-metrics logs: %s\n", err)
	}

	if controllerPodName != "" {
		cmd = exec.Command("kubectl", "describe", "pod", controllerPodName, "-n", namespace)
		podDescription, err := utils.Run(cmd)
		if err == nil {
			_, _ = fmt.Fprintf(GinkgoWriter, "Pod description:\n%s", podDescription)
		} else {
			_, _ = fmt.Fprintf(GinkgoWriter, "Failed to describe controller pod: %s\n", err)
		}
	}
}
