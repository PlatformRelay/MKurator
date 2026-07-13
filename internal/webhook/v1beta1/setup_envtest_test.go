package webhookv1beta1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

var _ = Describe("v1beta1 validating webhook setup", func() {
	It("registers every hub kind's validating webhook with the manager", func() {
		mgr, err := newWebhookManager(scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())

		Expect(SetupWithManager(mgr)).To(Succeed())
	})

	It("wraps the failing kind when a hub type is missing from the scheme", func() {
		// An empty scheme has no v1beta1 hub types registered, so the very first
		// setup call (QueueManagerConnection) fails to resolve its GVK and the
		// error must name that kind.
		mgr, err := newWebhookManager(runtime.NewScheme())
		Expect(err).NotTo(HaveOccurred())

		err = SetupWithManager(mgr)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("setup QueueManagerConnection webhook"))
	})
})
